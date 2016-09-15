package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"lib/db"
	"lib/marshal"

	"lib/metrics"

	"log"
	"net/http"
	"os"
	"policy-server/config"
	"policy-server/handlers"
	"policy-server/server_metrics"
	"policy-server/store"
	"policy-server/uaa_client"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/dropsonde"
	"github.com/jmoiron/sqlx"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
	"github.com/tedsuo/rata"
)

const (
	dropsondeOrigin = "policy-server"
	emitInterval    = 30 * time.Second
)

func main() {
	conf := &config.Config{}

	configFilePath := flag.String("config-file", "", "path to config file")
	flag.Parse()

	logger := lager.NewLogger("policy-server")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

	configData, err := ioutil.ReadFile(*configFilePath)
	if err != nil {
		log.Fatal("error reading config")
	}

	err = json.Unmarshal(configData, conf)
	if err != nil {
		log.Fatal("error unmarshalling config")
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: conf.SkipSSLValidation,
			},
		},
	}

	uaaRequestClient := &uaa_client.Client{
		Host:       conf.UAAURL,
		Name:       conf.UAAClient,
		Secret:     conf.UAAClientSecret,
		HTTPClient: httpClient,
		Logger:     logger,
	}
	whoamiHandler := &handlers.WhoAmIHandler{
		Client:    uaaRequestClient,
		Logger:    logger.Session("external"),
		Marshaler: marshal.MarshalFunc(json.Marshal),
	}
	uptimeHandler := &handlers.UptimeHandler{
		StartTime: time.Now(),
	}

	storeGroup := &store.Group{}
	destination := &store.Destination{}
	policy := &store.Policy{}

	retriableConnector := db.RetriableConnector{
		Connector:     db.GetConnectionPool,
		Sleeper:       db.SleeperFunc(time.Sleep),
		RetryInterval: 3 * time.Second,
		MaxRetries:    10,
	}

	type dbConnection struct {
		ConnectionPool *sqlx.DB
		Err            error
	}
	channel := make(chan dbConnection)
	go func() {
		connection, err := retriableConnector.GetConnectionPool(conf.Database)
		channel <- dbConnection{connection, err}
	}()
	var connectionResult dbConnection
	select {
	case connectionResult = <-channel:
	case <-time.After(5 * time.Second):
		log.Fatal("db connection timeout")
	}
	if connectionResult.Err != nil {
		log.Fatalf("db connect: %s", connectionResult.Err)
	}

	dataStore, err := store.New(
		connectionResult.ConnectionPool,
		storeGroup,
		destination,
		policy,
		conf.TagLength,
	)
	if err != nil {
		log.Fatalf("failed to construct datastore: %s", err)
	}

	unmarshaler := marshal.UnmarshalFunc(json.Unmarshal)

	authenticator := handlers.Authenticator{
		Client: uaaRequestClient,
		Logger: logger,
	}

	validator := &handlers.Validator{}

	createPolicyHandler := &handlers.PoliciesCreate{
		Logger:      logger.Session("policies-create"),
		Store:       dataStore,
		Unmarshaler: unmarshaler,
		Validator:   validator,
	}

	deletePolicyHandler := &handlers.PoliciesDelete{
		Logger:      logger.Session("policies-create"),
		Store:       dataStore,
		Unmarshaler: unmarshaler,
		Validator:   validator,
	}

	policiesIndexHandler := &handlers.PoliciesIndex{
		Logger:    logger.Session("policies-index"),
		Store:     dataStore,
		Marshaler: marshal.MarshalFunc(json.Marshal),
	}

	tagsIndexHandler := &handlers.TagsIndex{
		Logger:    logger.Session("tags-index"),
		Store:     dataStore,
		Marshaler: marshal.MarshalFunc(json.Marshal),
	}

	internalPoliciesHandler := &handlers.PoliciesIndexInternal{
		Logger:    logger.Session("policies-index-internal"),
		Store:     dataStore,
		Marshaler: marshal.MarshalFunc(json.Marshal),
	}

	routes := rata.Routes{
		{Name: "uptime", Method: "GET", Path: "/"},
		{Name: "uptime", Method: "GET", Path: "/networking"},
		{Name: "whoami", Method: "GET", Path: "/networking/v0/external/whoami"},
		{Name: "create_policies", Method: "POST", Path: "/networking/v0/external/policies"},
		{Name: "delete_policies", Method: "DELETE", Path: "/networking/v0/external/policies"},
		{Name: "policies_index", Method: "GET", Path: "/networking/v0/external/policies"},
		{Name: "internal_policies", Method: "GET", Path: "/networking/v0/internal/policies"},
		{Name: "tags_index", Method: "GET", Path: "/networking/v0/external/tags"},
	}

	handlers := rata.Handlers{
		"uptime":            uptimeHandler,
		"create_policies":   authenticator.Wrap(createPolicyHandler),
		"delete_policies":   authenticator.Wrap(deletePolicyHandler),
		"policies_index":    authenticator.Wrap(policiesIndexHandler),
		"tags_index":        authenticator.Wrap(tagsIndexHandler),
		"whoami":            whoamiHandler,
		"internal_policies": internalPoliciesHandler,
	}
	router, err := rata.NewRouter(routes, handlers)
	if err != nil {
		log.Fatalf("unable to create rata Router: %s", err) // not tested
	}

	addr := fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)
	server := http_server.New(addr, router)

	err = dropsonde.Initialize(conf.MetronAddress, dropsondeOrigin)
	if err != nil {
		log.Fatalf("initializing dropsonde: %s", err)
	}

	totalPoliciesSource := server_metrics.NewTotalPoliciesSource(dataStore)
	uptimeSource := metrics.NewUptimeSource()
	metricsEmitter := metrics.NewMetricsEmitter(logger, emitInterval, uptimeSource, totalPoliciesSource)

	members := grouper.Members{
		{"metrics_emitter", metricsEmitter},
		{"http_server", server},
	}

	logger.Info("starting", lager.Data{"listen-address": conf.ListenHost, "port": conf.ListenPort})

	group := grouper.NewOrdered(os.Interrupt, members)
	monitor := ifrit.Invoke(sigmon.New(group))

	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}

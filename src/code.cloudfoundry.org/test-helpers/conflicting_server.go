package testhelpers

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func LaunchConflictingServer(port int) *http.Server {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conflictingServer := &http.Server{Addr: address, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		err := conflictingServer.ListenAndServe()
		if err != nil {
			fmt.Fprintf(GinkgoWriter, "conflictingServer closed with error: %s\n", err)
		}
	}()
	client := &http.Client{}
	Eventually(func() bool {
		resp, err := client.Get(fmt.Sprintf("http://%s", address))
		if err != nil {
			return false
		}
		return resp.StatusCode == 404
	}).Should(BeTrue())
	return conflictingServer
}

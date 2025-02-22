package port_allocator_test

import (
	"encoding/json"
	"fmt"
	"time"

	"code.cloudfoundry.org/garden-external-networker/port_allocator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Tracker", func() {
	var (
		pool    *port_allocator.Pool
		tracker *port_allocator.Tracker
	)
	BeforeEach(func() {
		pool = &port_allocator.Pool{}
		tracker = &port_allocator.Tracker{
			StartPort: 100,
			Capacity:  10,
		}
	})

	Describe("AcquireOne", func() {
		It("reserves and returns a port from the pool", func() {
			newPort, err := tracker.AcquireOne(pool, "some-handle")
			Expect(err).NotTo(HaveOccurred())
			Expect(newPort).To(BeInRange(100, 110))
			Expect(pool.AcquiredPorts).To(Equal(map[uint32]string{newPort: "some-handle"}))
		})

		Context("when acquiring multiple ports", func() {
			It("gives unique ports", func() {
				firstPort, err := tracker.AcquireOne(pool, "some-handle")
				Expect(err).NotTo(HaveOccurred())
				secondPort, err := tracker.AcquireOne(pool, "some-handle")
				Expect(err).NotTo(HaveOccurred())

				Expect(pool.AcquiredPorts).To(HaveLen(2))
				Expect(firstPort).NotTo(Equal(secondPort))
				Expect(pool.AcquiredPorts).To(HaveKey(firstPort))
				Expect(pool.AcquiredPorts).To(HaveKey(secondPort))
			})
		})

		Context("when the only unacquired port is in the middle of the range", func() {
			BeforeEach(func() {
				tracker.Capacity = 3
				pool.AcquiredPorts = map[uint32]string{
					100: "some-handle",
					102: "some-handle",
				}
			})

			It("reserves and returns that unacquired port", func() {
				port, err := tracker.AcquireOne(pool, "some-handle")
				Expect(err).NotTo(HaveOccurred())
				Expect(port).To(Equal(uint32(101)))
				Expect(pool.AcquiredPorts).To(HaveKey(uint32(101)))
			})
		})

		Context("when the pool has reached capacity", func() {
			BeforeEach(func() {
				tracker.Capacity = 2
				pool.AcquiredPorts = map[uint32]string{
					100: "some-handle",
					101: "some-handle",
				}
			})

			It("returns a useful error", func() {
				_, err := tracker.AcquireOne(pool, "some-handle")
				Expect(err).To(Equal(port_allocator.ErrorPortPoolExhausted))
			})
		})

		Describe("performance", func() {
			It("should aquire all the ports quickly", Serial, func() {
				exp := gmeasure.NewExperiment("Acquiring Ports")
				AddReportEntry(exp.Name, exp)

				sampleSize := uint32(4_000)

				tracker.Capacity = sampleSize
				exp.Sample(func(idx int) {
					if idx%1_000 == 0 {
						fmt.Fprintf(GinkgoWriter, "Iteration %d\n", idx)
					}

					exp.MeasureDuration("runtime", func() {
						_, err := tracker.AcquireOne(pool, "some-handle")
						Expect(err).NotTo(HaveOccurred())
					})
				}, gmeasure.SamplingConfig{N: int(sampleSize)})

				stats := exp.GetStats("runtime")
				// no more than 1.5ms on average
				Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 1500*time.Microsecond))
			})
		})
	})

	Describe("acquire and release lifecycle", func() {
		It("can re-acquire ports which have been acquired and then released", func() {
			var err error
			for i := uint32(0); i < tracker.Capacity; i++ {
				if i%2 == 0 {
					_, err = tracker.AcquireOne(pool, "some-handle")
				} else {
					_, err = tracker.AcquireOne(pool, "some-handle2")
				}
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(tracker.ReleaseAll(pool, "some-handle")).To(Succeed())
			reacquired, err := tracker.AcquireOne(pool, "some-handle")
			Expect(err).NotTo(HaveOccurred())
			Expect(reacquired).To(Equal(uint32(100)))
		})
	})

	Describe("InRange", func() {
		It("returns true if the given port is in the allocation range", func() {
			for i := uint32(100); i < 110; i++ {
				Expect(tracker.InRange(i)).To(BeTrue())
			}
		})
		It("otherwise returns false", func() {
			Expect(tracker.InRange(110)).To(BeFalse())
		})
	})

	Describe("serializing the pool", func() {
		It("can be roud-tripped through JSON intact", func() {
			pool.AcquiredPorts = map[uint32]string{
				42:  "some-handle",
				105: "some-handle2",
			}

			bytes, err := json.Marshal(pool)
			Expect(err).NotTo(HaveOccurred())

			var newPool port_allocator.Pool
			Expect(json.Unmarshal(bytes, &newPool)).To(Succeed())

			Expect(newPool.AcquiredPorts).To(Equal(pool.AcquiredPorts))
		})

		It("marshals as a map from container handle to list of allocated ports", func() {
			pool.AcquiredPorts = map[uint32]string{
				42:  "some-handle",
				105: "some-handle2",
			}

			bytes, err := json.Marshal(pool)
			Expect(err).NotTo(HaveOccurred())

			Expect(bytes).To(MatchJSON(`{ "acquired_ports": {
				"some-handle": [ 42 ],
				"some-handle2": [ 105 ]
			} }`))
		})
	})
})

func BeInRange(min, max uint32) types.GomegaMatcher {
	return SatisfyAll(
		BeNumerically(">=", min),
		BeNumerically("<", max))
}

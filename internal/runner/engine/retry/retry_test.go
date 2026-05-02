package retry_test

import (
	"errors"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/runner/engine/retry"
)

func TestRetry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Retry Handler Suite")
}

var _ = Describe("Retry Handler", func() {
	var (
		handler *retry.Handler
	)

	BeforeEach(func() {
		handler = retry.NewRetryHandler(3, 100*time.Millisecond)
	})

	Describe("Execute", func() {
		It("should succeed without retries if the operation succeeds", func() {
			err := handler.Execute(func() error {
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(handler.GetStats().Attempts).To(Equal(1))
			Expect(handler.GetStats().Failures).To(Equal(0))
		})

		It("should retry the operation until it succeeds", func() {
			attempts := 0
			err := handler.Execute(func() error {
				attempts++
				if attempts < 3 {
					return errors.New("temporary error")
				}
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(handler.GetStats().Attempts).To(Equal(3))
			Expect(handler.GetStats().Failures).To(Equal(2))
		})

		It("should fail after max retries", func() {
			err := handler.Execute(func() error {
				return errors.New("permanent error")
			})
			Expect(err).To(HaveOccurred())
			Expect(handler.GetStats().Attempts).To(Equal(4))
			Expect(handler.GetStats().Failures).To(Equal(4))
		})
	})

	Describe("GetStats", func() {
		It("should return the correct stats", func() {
			err := handler.Execute(func() error {
				return errors.New("error")
			})
			Expect(err).To(HaveOccurred())
			stats := handler.GetStats()
			Expect(stats.Attempts).To(Equal(4))
			Expect(stats.Failures).To(Equal(4))
		})
	})

	Describe("Retryable", func() {
		It("should return true if attempts are within max retries", func() {
			err := handler.Execute(func() error {
				return errors.New("error")
			})
			Expect(err).To(HaveOccurred())
			Expect(handler.Retryable()).To(BeFalse())
		})
	})

	Describe("Reset", func() {
		It("should reset the stats", func() {
			err := handler.Execute(func() error {
				return errors.New("error")
			})
			Expect(err).To(HaveOccurred())
			handler.Reset()
			stats := handler.GetStats()
			Expect(stats.Attempts).To(Equal(0))
			Expect(stats.Failures).To(Equal(0))
		})
	})
})

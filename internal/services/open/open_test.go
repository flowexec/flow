package open_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/services/open"
)

func TestOpen(t *testing.T) {
	RegisterFailHandler(Fail)
	t.Setenv(open.DisabledEnvKey, "true")
	RunSpecs(t, "Open Suite")
}

var _ = Describe("Open", func() {
	It("should not return an error", func() {
		err := open.Open("http://example.com")
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("OpenWith", func() {
	It("should not return an error", func() {
		err := open.OpenWith("firefox", "http://example.com")
		Expect(err).NotTo(HaveOccurred())
	})
})

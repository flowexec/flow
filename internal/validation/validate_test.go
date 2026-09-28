package validation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/validation"
	"github.com/flowexec/flow/v2/types/executable"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validation Suite")
}

var _ = Describe("ValidateBytes verbs", func() {
	const flowFile = `executables:
  - verb: status
    name: services
    verbAliases: [health]
    exec:
      cmd: echo ok
`

	AfterEach(func() {
		executable.RegisterCustomVerbs()
	})

	It("accepts built-in verbs", func() {
		res, err := validation.ValidateBytes(
			[]byte("executables:\n  - verb: build\n    exec:\n      cmd: echo ok\n"),
			validation.FileTypeFlowFile, true,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Valid).To(BeTrue(), "%v", res.Errors)
	})

	It("rejects verbs that are not registered", func() {
		res, err := validation.ValidateBytes([]byte(flowFile), validation.FileTypeFlowFile, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Valid).To(BeFalse())
		Expect(res.Errors).To(ContainElements(
			validation.Issue{Path: "/executables/0/verb", Message: "invalid executable verb status"},
			validation.Issue{Path: "/executables/0/verbAliases/0", Message: "invalid executable verb health"},
		))
	})

	It("accepts registered custom verbs", func() {
		executable.RegisterCustomVerbs("status", "health")
		res, err := validation.ValidateBytes([]byte(flowFile), validation.FileTypeFlowFile, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Valid).To(BeTrue(), "%v", res.Errors)
	})

	It("rejects malformed verbs at the schema level", func() {
		res, err := validation.ValidateBytes(
			[]byte("executables:\n  - verb: Not Valid\n    exec:\n      cmd: echo ok\n"),
			validation.FileTypeFlowFile, false,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Valid).To(BeFalse())
	})
})

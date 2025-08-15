package flags_test

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal/flags"
)

func TestFlags(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flags Suite")
}

var _ = Describe("ToPflag", func() {
	var (
		cmd      *cobra.Command
		metadata flags.Metadata
	)

	BeforeEach(func() {
		cmd = &cobra.Command{}
	})

	DescribeTable("should correctly convert Metadata to pflag.Flagset",
		func(defaultValue interface{}, expectedType string) {
			metadata = flags.Metadata{
				Name:     "test",
				Default:  defaultValue,
				Required: false,
			}

			flagSet, err := flags.ToPflag(cmd, metadata, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(flagSet.Lookup("test").Value.Type()).To(Equal(expectedType))
		},
		Entry("string", "default", "string"),
		Entry("bool", true, "bool"),
		Entry("slice", []string{"default"}, "stringArray"),
		Entry("int", 1, "int"),
	)
})

var _ = Describe("ValueFor", func() {
	var (
		cmd      *cobra.Command
		metadata flags.Metadata
	)

	BeforeEach(func() {
		cmd = &cobra.Command{}
	})

	DescribeTable("should correctly return the value for the given Metadata",
		func(defaultValue interface{}, expectedValue interface{}) {
			metadata = flags.Metadata{
				Name:     "test",
				Default:  defaultValue,
				Required: false,
			}

			flagset, err := flags.ToPflag(cmd, metadata, false)
			Expect(err).NotTo(HaveOccurred())
			cmd.Flags().AddFlagSet(flagset)
			if reflect.TypeOf(expectedValue).Kind() == reflect.Slice {
				//nolint:errcheck
				err = cmd.ParseFlags([]string{"--test", expectedValue.([]string)[0]})
			} else {
				err = cmd.ParseFlags([]string{"--test=" + fmt.Sprintf("%v", expectedValue)})
			}
			Expect(err).NotTo(HaveOccurred())

			//nolint:exhaustive
			switch reflect.TypeOf(expectedValue).Kind() {
			case reflect.String:
				Expect(flags.ValueFor[string](cmd, metadata, false)).To(Equal(expectedValue))
			case reflect.Bool:
				Expect(flags.ValueFor[bool](cmd, metadata, false)).To(Equal(expectedValue))
			case reflect.Slice:
				Expect(flags.ValueFor[[]string](cmd, metadata, false)).To(Equal(expectedValue))
			case reflect.Int:
				Expect(flags.ValueFor[int](cmd, metadata, false)).To(Equal(expectedValue))
			}
		},
		Entry("string", "default", "default"),
		Entry("bool", true, true),
		Entry("slice", []string{"default"}, []string{"default"}),
		Entry("int", 1, 1),
	)
})

package fileparser_test

import (
	"os"
	"path/filepath"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/internal/fileparser"
	"github.com/flowexec/flow/types/executable"
)

var _ = Describe("ExtractExecConfig", func() {
	DescribeTable("should correctly parse simple configurations",
		func(file string, expectedFields map[string]string) {
			filePath := filepath.Join("testdata", file)
			fileBytes, err := os.ReadFile(filepath.Clean(filePath))
			Expect(err).ToNot(HaveOccurred())

			result, err := fileparser.ExtractExecConfig(string(fileBytes), "# ")
			Expect(err).ToNot(HaveOccurred())
			Expect(result.SimpleFields).To(Equal(expectedFields))
			Expect(result.Params).To(BeEmpty())
			Expect(result.Args).To(BeEmpty())
		},
		Entry(
			"simple key-value pairs",
			"simple.sh",
			map[string]string{
				fileparser.NameConfigurationKey: "hello",
				fileparser.VerbConfigurationKey: "show",
			}),
		Entry(
			"values with spaces in quotes",
			"quoted.sh",
			map[string]string{
				fileparser.NameConfigurationKey:        "value 1",
				fileparser.VerbConfigurationKey:        "value2",
				fileparser.DescriptionConfigurationKey: "value 3",
			}),
		Entry(
			"values with escaped characters",
			"escaped.sh",
			map[string]string{
				fileparser.NameConfigurationKey:        "value 1' one",
				fileparser.DescriptionConfigurationKey: "'value two'",
				fileparser.TagConfigurationKey:         "tag1|tag2",
			}),
		Entry(
			"repeated key configurations",
			"repeated.sh",
			map[string]string{
				fileparser.TagConfigurationKey:         "tag1|tag2|tag3|tag4|tag5",
				fileparser.AliasConfigurationKey:       "alias",
				fileparser.DescriptionConfigurationKey: "first line\nsecond line\nthird line",
			}),
		Entry(
			"complex configuration parsing",
			"complex.sh",
			map[string]string{
				fileparser.NameConfigurationKey:        "name",
				fileparser.VerbConfigurationKey:        "verb",
				fileparser.DescriptionConfigurationKey: "first line\nsecond line\nthird line with 'quotes', and commas\nclosin'",
			}),
		Entry(
			"multi-line description",
			"multiline.sh",
			map[string]string{
				fileparser.DescriptionConfigurationKey: expectedMultiLineDescription,
			}),
		Entry(
			"directory and log mode configuration",
			"dir-logmode.sh",
			map[string]string{
				fileparser.NameConfigurationKey:    "test-dir-logmode",
				fileparser.VerbConfigurationKey:    "test",
				fileparser.DirConfigurationKey:     "./subdir",
				fileparser.LogModeConfigurationKey: "text",
			}),
	)

	DescribeTable("should correctly parse parameter configurations",
		func(file string, expectedFields map[string]string, expectedParams executable.ParameterList) {
			filePath := filepath.Join("testdata", file)
			fileBytes, err := os.ReadFile(filepath.Clean(filePath))
			Expect(err).ToNot(HaveOccurred())

			result, err := fileparser.ExtractExecConfig(string(fileBytes), "# ")
			Expect(err).ToNot(HaveOccurred())
			Expect(result.SimpleFields).To(Equal(expectedFields))
			Expect(result.Params).To(Equal(expectedParams))
			Expect(result.Args).To(BeEmpty())
		},
		Entry(
			"parameters configuration",
			"params.sh",
			map[string]string{
				fileparser.NameConfigurationKey: "test-params",
				fileparser.VerbConfigurationKey: "test",
			},
			executable.ParameterList{
				{SecretRef: "my-secret", EnvKey: "SECRET_VAR"},
				{Prompt: "Enter name", EnvKey: "NAME_VAR"},
				{Text: "default-value", EnvKey: "DEFAULT_VAR"},
			}),
	)

	var pos1 = 1

	DescribeTable("should correctly parse argument configurations",
		func(file string, expectedFields map[string]string, expectedArgs executable.ArgumentList) {
			filePath := filepath.Join("testdata", file)
			fileBytes, err := os.ReadFile(filepath.Clean(filePath))
			Expect(err).ToNot(HaveOccurred())

			result, err := fileparser.ExtractExecConfig(string(fileBytes), "# ")
			Expect(err).ToNot(HaveOccurred())
			Expect(result.SimpleFields).To(Equal(expectedFields))
			Expect(result.Params).To(BeEmpty())
			for _, arg := range expectedArgs {
				Expect(slices.ContainsFunc(result.Args, func(argument executable.Argument) bool {
					return argument.EnvKey == arg.EnvKey
				})).To(BeTrue())
			}
		},
		Entry(
			"arguments configuration",
			"args.sh",
			map[string]string{
				fileparser.NameConfigurationKey: "test-args",
				fileparser.VerbConfigurationKey: "test",
			},
			executable.ArgumentList{
				{Flag: "verbose", EnvKey: "VERBOSE"},
				{Pos: &pos1, EnvKey: "FILENAME"},
				{Flag: "count", EnvKey: "COUNT"},
			}),
	)

	Describe("should handle parsing errors gracefully", func() {
		It("should return error for malformed params", func() {
			content := "# f:params=invalid:format"
			_, err := fileparser.ExtractExecConfig(content, "# ")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("param 1 requires at least 3 fields"))
		})

		It("should return error for malformed args", func() {
			content := "# f:args=invalid:format"
			_, err := fileparser.ExtractExecConfig(content, "# ")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("arg 1 requires exactly 3 fields"))
		})

		It("should return error for invalid param type", func() {
			content := "# f:params=invalid:value:ENV_KEY"
			_, err := fileparser.ExtractExecConfig(content, "# ")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid parameter type"))
		})

		It("should return error for invalid arg type", func() {
			content := "# f:args=invalid:value:ENV_KEY"
			_, err := fileparser.ExtractExecConfig(content, "# ")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid argument type"))
		})

		It("should return error for invalid position in pos arg", func() {
			content := "# f:args=pos:notanumber:ENV_KEY"
			_, err := fileparser.ExtractExecConfig(content, "# ")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid position number"))
		})
	})

	Describe("should handle singular and plural key forms", func() {
		It("should normalize singular forms to plural", func() {
			content := `# f:tag=production f:alias=prod-deploy f:param=text:value:ENV f:arg=flag:test:TEST`
			result, err := fileparser.ExtractExecConfig(content, "# ")
			Expect(err).ToNot(HaveOccurred())

			Expect(result.SimpleFields).To(HaveKey(fileparser.TagConfigurationKey))
			Expect(result.SimpleFields).To(HaveKey(fileparser.AliasConfigurationKey))
			Expect(result.SimpleFields[fileparser.TagConfigurationKey]).To(Equal("production"))
			Expect(result.SimpleFields[fileparser.AliasConfigurationKey]).To(Equal("prod-deploy"))

			Expect(result.Params).To(HaveLen(1))
			Expect(result.Args).To(HaveLen(1))
		})

		It("should handle mixed singular/plural usage", func() {
			content := `# f:tag=production f:tags=deployment|staging f:alias=prod f:aliases=deploy`
			result, err := fileparser.ExtractExecConfig(content, "# ")
			Expect(err).ToNot(HaveOccurred())

			Expect(result.SimpleFields[fileparser.TagConfigurationKey]).To(Equal("production|deployment|staging"))
			Expect(result.SimpleFields[fileparser.AliasConfigurationKey]).To(Equal("prod|deploy"))
		})
	})
})

const expectedMultiLineDescription = `first line
second line
    third line, with commas and 'quotes'
fourth line
fifth line
sixth line
seventh line
eighth line`

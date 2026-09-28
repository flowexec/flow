package executable_test

import (
	"os"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/types/executable"
)

var _ = Describe("ParseExecutableID", func() {
	DescribeTable("valid IDs",
		func(id, wantWs, wantNs, wantName string) {
			ws, ns, name, err := executable.ParseExecutableID(id)
			Expect(err).NotTo(HaveOccurred())
			Expect(ws).To(Equal(wantWs))
			Expect(ns).To(Equal(wantNs))
			Expect(name).To(Equal(wantName))
		},
		Entry("empty", "", "*", "", ""),
		Entry("name only", "build", "*", "*", "build"),
		Entry("namespace and name", "api:build", "*", "api", "build"),
		Entry("explicit empty namespace", ":build", "*", "", "build"),
		Entry("workspace and name", "ws/build", "ws", "*", "build"),
		Entry("fully qualified", "ws/api:build", "ws", "api", "build"),
		Entry("workspace only", "ws/", "ws", "*", ""),
		Entry("workspace and namespace, no name", "ws/api:", "ws", "api", ""),
		Entry("nested namespace", "ws/api/v2:build", "ws", "api/v2", "build"),
		Entry("deeply nested namespace", "ws/a/b/c:build", "ws", "a/b/c", "build"),
		Entry("nested namespace, no name", "ws/api/v2:", "ws", "api/v2", ""),
		Entry("current workspace shorthand", "./api/v2:build", ".", "api/v2", "build"),
		Entry("dots and at signs", "my.ws/team@x/v1.2:db.migrate", "my.ws", "team@x/v1.2", "db.migrate"),
	)

	DescribeTable("invalid IDs",
		func(id string) {
			_, _, _, err := executable.ParseExecutableID(id)
			Expect(err).To(HaveOccurred())
		},
		Entry("nested path without namespace separator", "ws/api/build"),
		Entry("colon in name", "ws/api:build:extra"),
		Entry("empty namespace segment", "ws/api//v2:build"),
		Entry("trailing namespace separator", "ws/api/:build"),
		Entry("dot namespace segment", "ws/api/./v2:build"),
		Entry("dot-dot namespace segment", "ws/../v2:build"),
		Entry("unsupported character in name", "ws/api:build+x"),
	)

	It("round-trips nested namespaces through NewExecutableID", func() {
		id := executable.NewExecutableID("ws", "api/v2", "build")
		Expect(id).To(Equal("ws/api/v2:build"))
		ws, ns, name := executable.MustParseExecutableID(id)
		Expect([]string{ws, ns, name}).To(Equal([]string{"ws", "api/v2", "build"}))
	})
})

var _ = Describe("ValidateNamespace", func() {
	DescribeTable("namespaces",
		func(ns string, valid bool) {
			err := executable.ValidateNamespace(ns)
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},
		Entry("root", "", true),
		Entry("single segment", "api", true),
		Entry("nested", "api/v2", true),
		Entry("leading separator", "/api", false),
		Entry("trailing separator", "api/", false),
		Entry("empty segment", "api//v2", false),
		Entry("colon", "api:v2", false),
		Entry("space", "api v2", false),
		Entry("wildcard", "*", false),
		Entry("dots and at signs", "team@x/v1.2", true),
		Entry("dot segment", "api/.", false),
		Entry("dot-dot segment", "../api", false),
	)
})

var _ = Describe("NamespaceMatches", func() {
	DescribeTable("filters",
		func(ns, filter string, want bool) {
			Expect(executable.NamespaceMatches(ns, filter)).To(Equal(want))
		},
		Entry("wildcard", "api/v2", "*", true),
		Entry("exact", "api", "api", true),
		Entry("exact does not include children", "api/v2", "api", false),
		Entry("subtree includes parent", "api", "api/*", true),
		Entry("subtree includes child", "api/v2", "api/*", true),
		Entry("subtree includes grandchild", "api/v2/x", "api/*", true),
		Entry("subtree excludes sibling prefix", "apix", "api/*", false),
		Entry("nested subtree", "api/v2/x", "api/v2/*", true),
		Entry("root exact", "", "", true),
	)
})

var _ = Describe("ExecutableIDPattern", func() {
	pattern := regexp.MustCompile(executable.ExecutableIDPattern)

	DescribeTable("matching",
		func(id string, want bool) {
			Expect(pattern.MatchString(id)).To(Equal(want))
		},
		Entry("name", "build", true),
		Entry("ns:name", "api:build", true),
		Entry("ws/name", "ws/build", true),
		Entry("ws/", "ws/", true),
		Entry("ws/ns:", "ws/api:", true),
		Entry("nested", "ws/api/v2:build", true),
		Entry("current workspace", "./api/v2:build", true),
		Entry("nested without ws", "api/v2:build", true), // parses as ws=api, ns=v2
		Entry("dots and at signs", "my.ws/team@x/v1.2:db.migrate", true),
		Entry("nested path without colon", "ws/api/build", false),
		Entry("space", "ws/api:bu ild", false),
	)
})

var _ = Describe("Custom verbs", func() {
	AfterEach(func() {
		executable.RegisterCustomVerbs()
	})

	It("accepts a registered custom verb", func() {
		Expect(executable.Verb("status").Validate()).To(HaveOccurred())
		executable.RegisterCustomVerbs("status")
		Expect(executable.Verb("status").Validate()).To(Succeed())
		Expect(executable.SortedValidVerbs()).To(ContainElement("status"))
	})

	It("replaces the registered set on each call", func() {
		executable.RegisterCustomVerbs("status")
		executable.RegisterCustomVerbs("health")
		Expect(executable.Verb("status").Validate()).To(HaveOccurred())
		Expect(executable.Verb("health").Validate()).To(Succeed())
	})

	It("ignores invalid names, built-ins, and duplicates", func() {
		builtins := len(executable.ValidVerbs())
		executable.RegisterCustomVerbs("status", "status", "build", "Bad")
		Expect(executable.ValidVerbs()).To(HaveLen(builtins + 1))
	})

	It("keeps custom verbs out of built-in alias groups", func() {
		executable.RegisterCustomVerbs("status")
		Expect(executable.Verb("status").Equals(executable.VerbView)).To(BeFalse())
		Expect(executable.RelatedVerbs("status")).To(BeEmpty())
	})

	DescribeTable("ValidateCustomVerbName",
		func(name string, valid bool) {
			err := executable.ValidateCustomVerbName(name)
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},
		Entry("simple word", "status", true),
		Entry("hyphenated", "health-check", true),
		Entry("with digits", "check2", true),
		Entry("empty", "", false),
		Entry("uppercase", "Status", false),
		Entry("space", "my verb", false),
		Entry("leading digit", "2fa", false),
		Entry("wildcard", "*", false),
		Entry("built-in", "build", false),
	)
})

var _ = Describe("Built-in verbs", func() {
	It("match the schema's verb enum and custom verb pattern", func() {
		data, err := os.ReadFile("executable_schema.yaml")
		Expect(err).NotTo(HaveOccurred())
		var schema struct {
			Definitions struct {
				Verb struct {
					Enum    []string `yaml:"enum"`
					Docsgen struct {
						OpenPattern string `yaml:"openPattern"`
					} `yaml:"docsgen"`
				} `yaml:"Verb"`
			} `yaml:"definitions"`
		}
		Expect(yaml.Unmarshal(data, &schema)).To(Succeed())
		Expect(schema.Definitions.Verb.Enum).NotTo(BeEmpty())
		Expect(executable.SortedValidVerbs()).To(ConsistOf(schema.Definitions.Verb.Enum))
		Expect(schema.Definitions.Verb.Docsgen.OpenPattern).To(Equal(executable.CustomVerbPattern))
	})
})

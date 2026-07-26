package executable_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/types/executable"
)

var _ = Describe("Template", func() {
	var (
		template *executable.Template
	)

	BeforeEach(func() {
		template = &executable.Template{
			Artifacts: []executable.Artifact{
				{SrcName: "main.go"},
				{SrcName: "go.mod"},
			},
			Form: executable.FormFields{
				&executable.Field{
					Key:     "testKey",
					Prompt:  "testPrompt",
					Default: "testDefault",
				},
			},
			Template: `namespace: test
description: {{ .testKey }}
tags: [test]
`,
		}
		template.SetContext("flowfile", "flowfile.flow.tmpl")
	})

	Describe("SetContext", func() {
		It("should set the context correctly", func() {
			template.SetContext("newName", "new/flowfile.flow.tmpl")
			Expect(template.Name()).To(Equal("newName"))
			Expect(template.Location()).To(Equal("new/flowfile.flow.tmpl"))
		})

		It("should set the name from the location when empty", func() {
			template.SetContext("", "new/flowfile.flow.tmpl")
			Expect(template.Name()).To(Equal("flowfile"))
			Expect(template.Location()).To(Equal("new/flowfile.flow.tmpl"))
		})
	})

	Describe("Validate", func() {
		It("should validate the form config correctly", func() {
			Expect(template.Validate()).To(Succeed())
		})

		It("should error when there is an invalid form field", func() {
			template.Form = append(template.Form, &executable.Field{Description: "i have missing fields"})
			Expect(template.Validate()).To(HaveOccurred())
		})
	})

	Describe("Format Methods", func() {
		It("JSON should return the JSON representation of the template", func() {
			str, err := template.JSON()
			Expect(err).NotTo(HaveOccurred())
			Expect(str).ToNot(BeEmpty())
		})
		It("YAML should return the YAML representation of the template", func() {
			str, err := template.YAML()
			Expect(err).NotTo(HaveOccurred())
			Expect(str).ToNot(BeEmpty())
		})
		It("Markdown should return the Markdown representation of the template", func() {
			str := template.Markdown()
			Expect(str).ToNot(BeEmpty())
		})
	})

	Describe("Identity in serialized output", func() {
		It("JSON should include the name and location", func() {
			str, err := template.JSON()
			Expect(err).NotTo(HaveOccurred())

			var out map[string]interface{}
			Expect(json.Unmarshal([]byte(str), &out)).To(Succeed())
			Expect(out).To(HaveKeyWithValue("name", "flowfile"))
			Expect(out).To(HaveKeyWithValue("assignedName", "flowfile"))
			Expect(out).To(HaveKeyWithValue("location", "flowfile.flow.tmpl"))
		})

		It("YAML should include the name and location", func() {
			str, err := template.YAML()
			Expect(err).NotTo(HaveOccurred())
			Expect(str).To(ContainSubstring("name: flowfile"))
			Expect(str).To(ContainSubstring("location: flowfile.flow.tmpl"))
		})

		It("should round-trip through JSON without losing its identity", func() {
			str, err := template.JSON()
			Expect(err).NotTo(HaveOccurred())

			decoded := &executable.Template{}
			Expect(json.Unmarshal([]byte(str), decoded)).To(Succeed())
			Expect(decoded.Name()).To(Equal("flowfile"))
			Expect(decoded.Location()).To(Equal("flowfile.flow.tmpl"))
			Expect(decoded.Template).To(Equal(template.Template))
		})

		It("should not panic when the context was never set", func() {
			bare := &executable.Template{Template: "namespace: test\n"}
			Expect(bare.Name()).To(BeEmpty())
			Expect(bare.Location()).To(BeEmpty())

			str, err := bare.JSON()
			Expect(err).NotTo(HaveOccurred())

			var out map[string]interface{}
			Expect(json.Unmarshal([]byte(str), &out)).To(Succeed())
			Expect(out).To(HaveKeyWithValue("name", ""))
			Expect(out).To(HaveKeyWithValue("location", ""))
		})
	})
})

var _ = Describe("TemplateList", func() {
	var (
		templates executable.TemplateList
	)

	BeforeEach(func() {
		templates = []*executable.Template{
			{
				Artifacts: []executable.Artifact{
					{SrcName: "main.go"},
					{SrcName: "go.mod"},
				},
				Form: executable.FormFields{
					&executable.Field{
						Key:     "testKey",
						Prompt:  "testPrompt",
						Default: "testDefault",
					},
				},
				Template: `namespace: test
description: {{ .testKey }}
tags: [test]
`,
			},
			{
				Template: `namespace: test2
description: test2
tags: [test2]
`,
			},
		}
		templates[0].SetContext("flowfile", "flowfile.flow.tmpl")
		templates[1].SetContext("flowfile2", "flowfile2.flow.tmpl")
	})

	Describe("Format Methods", func() {
		It("JSON should return the JSON representation of the templates", func() {
			str, err := templates.JSON()
			Expect(err).NotTo(HaveOccurred())
			Expect(str).ToNot(BeEmpty())
		})
		It("YAML should return the YAML representation of the templates", func() {
			str, err := templates.YAML()
			Expect(err).NotTo(HaveOccurred())
			Expect(str).ToNot(BeEmpty())
		})
		It("Items should return the tuikit item representation of the templates", func() {
			items := templates.Items()
			Expect(items).To(HaveLen(2))
		})
		It("JSON should carry the identity of every element in the list", func() {
			str, err := templates.JSON()
			Expect(err).NotTo(HaveOccurred())

			var out []map[string]interface{}
			Expect(json.Unmarshal([]byte(str), &out)).To(Succeed())
			Expect(out).To(HaveLen(2))
			Expect(out[0]).To(HaveKeyWithValue("name", "flowfile"))
			Expect(out[0]).To(HaveKeyWithValue("location", "flowfile.flow.tmpl"))
			Expect(out[1]).To(HaveKeyWithValue("name", "flowfile2"))
			Expect(out[1]).To(HaveKeyWithValue("location", "flowfile2.flow.tmpl"))
		})
	})

	Describe("Find", func() {
		It("should find the correct template", func() {
			Expect(templates.Find("flowfile2")).ToNot(BeNil())
		})

		It("should return nil when the template is not found", func() {
			Expect(templates.Find("flowfile3")).To(BeNil())
		})
	})
})

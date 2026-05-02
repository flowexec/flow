package schema_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/tools/docsgen/schema"
)

func TestSchema(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Schema Suite")
}

var _ = Describe("MergeSchemas", func() {
	var (
		dst       *schema.JSONSchema
		dstFile   schema.FileName
		schemaMap map[schema.FileName]*schema.JSONSchema
	)

	BeforeEach(func() {
		dst = &schema.JSONSchema{
			Definitions: map[schema.FieldKey]*schema.JSONSchema{
				"MyEnum": {Type: "string", Enum: []string{"a", "b", "c"}},
			},
			Properties: map[schema.FieldKey]*schema.JSONSchema{
				"prop1": {Ref: "#/definitions/MyEnum"},
				"prop2": {Ref: "../alfa/schema.yaml#/definitions/MyString"},
				"prop3": {Ref: "../charlie/schema.yaml#/"},
				"prop4": {Ref: "../bravo/other_schema.yaml#/definitions/MyString"},
			},
		}
		//nolint:exhaustive
		schemaMap = map[schema.FileName]*schema.JSONSchema{
			"alfa/schema.yaml": {
				Definitions: map[schema.FieldKey]*schema.JSONSchema{
					"MyString": {Type: "string"},
					"MyBool":   {Type: "boolean"},
				},
				Required: []string{"MyString"},
			},
			"bravo/schema.yaml": dst,
			"bravo/other_schema.yaml": {
				Definitions: map[schema.FieldKey]*schema.JSONSchema{
					"MyString": {Type: "string"},
				},
			},
			"charlie/schema.yaml": {
				Definitions: map[schema.FieldKey]*schema.JSONSchema{
					"MyInt": {Type: "integer"},
				},
				Properties: map[schema.FieldKey]*schema.JSONSchema{
					"prop1": {Ref: "#/definitions/MyInt"},
				},
			},
		}
		dstFile = "bravo/schema.yaml"
	})

	Context("when the source is a local ref", func() {
		var src *schema.JSONSchema
		BeforeEach(func() {
			src = dst.Properties["prop1"]
		})

		It("should keep the defined schema", func() {
			schema.MergeSchemas(dst, src, dstFile, schemaMap)
			Expect(dst.Definitions).To(HaveKey(schema.FieldKey("MyEnum")))
		})
	})

	Context("when the source is an external ref", func() {
		var src *schema.JSONSchema
		BeforeEach(func() {
			src = dst.Properties["prop2"]
		})

		It("should merge the external schema", func() {
			schema.MergeSchemas(dst, src, dstFile, schemaMap)
			Expect(dst.Definitions).To(HaveKey(schema.FieldKey("MyEnum")))
			// the key should include the external file tile
			Expect(dst.Definitions).To(HaveKey(schema.FieldKey("AlfaMyString")))
			Expect(src.Ref).To(Equal(schema.Ref("#/definitions/AlfaMyString")))
		})
	})

	Context("when the source is an external ref with a path to the root", func() {
		var src *schema.JSONSchema
		BeforeEach(func() {
			src = dst.Properties["prop3"]
		})

		It("should merge the external schema", func() {
			schema.MergeSchemas(dst, src, dstFile, schemaMap)
			Expect(dst.Definitions).To(HaveKey(schema.FieldKey("MyEnum")))
			Expect(dst.Definitions).To(HaveKey(schema.FieldKey("Charlie")))
			Expect(dst.Definitions).To(HaveKey(schema.FieldKey("CharlieMyInt")))
			Expect(src.Ref).To(Equal(schema.Ref("#/definitions/Charlie")))
		})
	})

	Context("when the source is an external ref with a path to a definition", func() {
		var src *schema.JSONSchema
		BeforeEach(func() {
			src = dst.Properties["prop4"]
		})

		It("should merge the external schema", func() {
			schema.MergeSchemas(dst, src, dstFile, schemaMap)
			Expect(dst.Definitions).To(HaveKey(schema.FieldKey("MyEnum")))
			Expect(dst.Definitions).To(HaveKey(schema.FieldKey("OtherMyString")))
			Expect(src.Ref).To(Equal(schema.Ref("#/definitions/OtherMyString")))
		})
	})
})

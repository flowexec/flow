package executable_test

import (
	"fmt"
	"slices"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/types/common"
	"github.com/flowexec/flow/types/executable"
)

func TestExecutable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Executable Type Suite")
}

var _ = Describe("Executable", func() {
	const (
		testWsName        = "workspace"
		testNsName        = "namespace"
		testWorkspacePath = "workspace-path"
		testExecCfgPath   = "config-path"
	)
	var exec *executable.Executable

	BeforeEach(func() {
		exec = &executable.Executable{
			Verb:        "run",
			Name:        "test",
			Aliases:     []string{"t"},
			Description: "test description",
			Exec: &executable.ExecExecutableType{
				Cmd: "echo hello",
			},
		}
		exec.SetDefaults()
		exec.SetContext(testWsName, testWorkspacePath, testNsName, testExecCfgPath)
		Expect(exec.Validate()).To(Succeed())
	})

	Describe("Format Methods", func() {
		It("JSON should return the JSON representation of the executable", func() {
			str, err := exec.JSON()
			Expect(err).NotTo(HaveOccurred())
			Expect(str).ToNot(BeEmpty())
		})
		It("YAML should return the YAML representation of the executable", func() {
			str, err := exec.YAML()
			Expect(err).NotTo(HaveOccurred())
			Expect(str).ToNot(BeEmpty())
		})
		It("Markdown should return the Markdown representation of the executable", func() {
			str := exec.Markdown()
			Expect(str).ToNot(BeEmpty())
		})
	})

	Describe("Metadata Methods", func() {
		It("Ref should return the reference of the executable", func() {
			Expect(exec.Ref().String()).
				To(Equal(fmt.Sprintf("%s %s/%s:%s", exec.Verb, testWsName, testNsName, exec.Name)))
		})
		It("Ref should return the Ref of the executable", func() {
			Expect(exec.ID()).
				To(Equal(fmt.Sprintf("%s/%s:%s", testWsName, testNsName, exec.Name)))
		})
		It("WorkspacePath should return the workspace path of the executable", func() {
			Expect(exec.WorkspacePath()).To(Equal(testWorkspacePath))
		})
		It("FlowFilePath should return the exec config path of the executable", func() {
			Expect(exec.FlowFilePath()).To(Equal(testExecCfgPath))
		})
	})

	Describe("AliasesIDs", func() {
		It("should return the correct aliases IDs", func() {
			exec.Aliases = []string{"alias1", "alias2"}
			aliasesIDs := exec.AliasesIDs()
			Expect(aliasesIDs).To(ConsistOf(
				fmt.Sprintf("%s/%s:alias1", testWsName, testNsName),
				fmt.Sprintf("%s/%s:alias2", testWsName, testNsName),
			))
		})

		It("should return nil if there are no aliases", func() {
			exec.Aliases = nil
			aliasesIDs := exec.AliasesIDs()
			Expect(aliasesIDs).To(BeNil())
		})
	})

	Describe("NameEquals", func() {
		It("should return the expected value", func() {
			By("having a matching name")
			Expect(exec.NameEquals(exec.Name)).To(BeTrue())

			By("having a matching alias")
			Expect(exec.NameEquals(exec.Aliases[0])).To(BeTrue())

			By("not having a matching name or alias")
			Expect(exec.NameEquals("nonexistent")).To(BeFalse())
		})
	})

	Describe("MergeTags", func() {
		It("should merge the given tags with the executable's existing tags", func() {
			exec.MergeTags(common.Tags{"tag1", "tag2"})
			Expect(exec.Tags).To(ConsistOf("tag1", "tag2"))
		})

		It("should remove duplicate tags", func() {
			exec.MergeTags(common.Tags{"tag1", "tag1"})
			compact := slices.Compact(exec.Tags)
			Expect(compact).To(HaveLen(len(exec.Tags)))
		})
	})

	DescribeTable("IsVisibleFromWorkspace", func(visibility *common.Visibility, wsMatch, expected bool) {
		v := executable.ExecutableVisibility(*visibility)
		exec.Visibility = &v
		if wsMatch {
			Expect(exec.IsVisibleFromWorkspace(testWsName, common.VisibilityPrivate)).To(Equal(expected))
		} else {
			Expect(exec.IsVisibleFromWorkspace("another-ws", common.VisibilityPrivate)).To(Equal(expected))
		}
	},
		Entry("public from ws", common.VisibilityPublic.NewPointer(), true, true),
		Entry("public from another ws", common.VisibilityPublic.NewPointer(), false, true),
		Entry("private from ws", common.VisibilityPrivate.NewPointer(), true, true),
		Entry("private from another ws", common.VisibilityPrivate.NewPointer(), false, false),
		Entry("internal from ws", common.VisibilityInternal.NewPointer(), true, false),
		Entry("internal from another ws", common.VisibilityInternal.NewPointer(), false, false),
		Entry("hidden from ws", common.VisibilityHidden.NewPointer(), true, false),
		Entry("hidden from another ws", common.VisibilityHidden.NewPointer(), false, false),
	)

	DescribeTable("IsExecutableFromWorkspace", func(visibility *common.Visibility, wsMatch, expected bool) {
		v := executable.ExecutableVisibility(*visibility)
		exec.Visibility = &v
		if wsMatch {
			Expect(exec.IsExecutableFromWorkspace(testWsName)).To(Equal(expected))
		} else {
			Expect(exec.IsExecutableFromWorkspace("another-ws")).To(Equal(expected))
		}
	},
		Entry("public from ws", common.VisibilityPublic.NewPointer(), true, true),
		Entry("public from another ws", common.VisibilityPublic.NewPointer(), false, true),
		Entry("private from ws", common.VisibilityPrivate.NewPointer(), true, true),
		Entry("private from another ws", common.VisibilityPrivate.NewPointer(), false, false),
		Entry("internal from ws", common.VisibilityInternal.NewPointer(), true, true),
		Entry("internal from another ws", common.VisibilityInternal.NewPointer(), false, false),
		Entry("hidden from ws", common.VisibilityHidden.NewPointer(), true, false),
		Entry("hidden from another ws", common.VisibilityHidden.NewPointer(), false, false),
	)
})

var _ = Describe("ExecutableList", func() {
	const (
		exec1Ws = "ws1"
		exec2Ws = "ws2"
		exec1Ns = "ns1"
		exec2Ns = "ns2"
	)
	var (
		exec1 *executable.Executable
		exec2 *executable.Executable
		execs executable.ExecutableList
	)

	BeforeEach(func() {
		exec1 = &executable.Executable{
			Verb: "run",
			Name: "test1",
			Exec: &executable.ExecExecutableType{
				Cmd: "echo hello",
			},
		}
		exec1.SetDefaults()
		exec1.SetContext(exec1Ws, "workspace-path", exec1Ns, "exec-definition-path")
		exec2 = &executable.Executable{
			Verb: "start",
			Name: "test2",
			Exec: &executable.ExecExecutableType{
				Cmd: "echo hello",
			},
		}
		exec2.SetDefaults()
		exec2.SetContext(exec2Ws, "workspace-path", exec2Ns, "exec-definition-path")
		execs = executable.ExecutableList{exec1, exec2}
	})

	Describe("Format Methods", func() {
		It("JSON should return the JSON representation of the executables", func() {
			str, err := execs.JSON()
			Expect(err).NotTo(HaveOccurred())
			Expect(str).ToNot(BeEmpty())
		})
		It("YAML should return the YAML representation of the executables", func() {
			str, err := execs.YAML()
			Expect(err).NotTo(HaveOccurred())
			Expect(str).ToNot(BeEmpty())
		})
		It("Items should return the Markdown representation of the executables", func() {
			items := execs.Items()
			// TODO: test the markdown content
			Expect(items).To(HaveLen(2))
		})
	})

	Describe("FilterByNamespace", func() {
		It("should return only executables with the given namespace", func() {
			filtered := execs.FilterByNamespace(exec1Ns)
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].Name).To(Equal(exec1.Name))
		})
	})

	Describe("FilterByTag", func() {
		BeforeEach(func() {
			exec1.Tags = executable.ExecutableTags{"tag1", "tag2"}
			exec2.Tags = executable.ExecutableTags{"tag3", "tag4"}
			execs = executable.ExecutableList{exec1, exec2}
		})

		It("should return executables with the given tag", func() {
			filtered := execs.FilterByTags(common.Tags{"tag2"})
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].Name).To(Equal(exec1.Name))
		})

		It("should return no executables if the tag is not found", func() {
			filtered := execs.FilterByTags(common.Tags{"tag5"})
			Expect(filtered).To(BeEmpty())
		})

		It("should return executables with multiple tags", func() {
			filtered := execs.FilterByTags(common.Tags{"tag2", "tag3"})
			Expect(filtered).To(HaveLen(2))
			Expect(filtered[0].Name).To(Equal(exec1.Name))
			Expect(filtered[1].Name).To(Equal(exec2.Name))
		})
	})

	Describe("FilterByVerb", func() {
		BeforeEach(func() {
			exec1.Verb = "run"
			exec2.Verb = "launch"
			execs = executable.ExecutableList{exec1, exec2}
		})

		It("should return executables with the given verb", func() {
			filtered := execs.FilterByVerb("run")
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].Name).To(Equal(exec1.Name))
		})

		It("should return no executables if the verb is not found", func() {
			filtered := execs.FilterByVerb("uninstall")
			Expect(filtered).To(BeEmpty())
		})
	})

	Describe("FilterByWorkspace", func() {
		It("should return executables with the given workspace", func() {
			filtered := execs.FilterByWorkspace(exec1Ws)
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].Name).To(Equal(exec1.Name))
		})
	})

	Describe("FilterBySubstring", func() {
		It("should return executables when the reference matches", func() {
			exec1.Name = "abcdefgh"
			filtered := execs.FilterBySubstring("def")
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].Name).To(Equal(exec1.Name))
		})

		It("should return executables when the description matches", func() {
			exec1.Description = "abcdefgh"
			filtered := execs.FilterBySubstring("def")
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].Name).To(Equal(exec1.Name))
		})

		It("should return executables when one of the aliases matches", func() {
			exec1.Aliases = append(exec1.Aliases, "ijklmn")
			exec1.Aliases = append(exec1.Aliases, "opqrst")
			filtered := execs.FilterBySubstring("pqr")
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].Name).To(Equal(exec1.Name))
		})
	})

	Describe("FindByVerbAndID", func() {
		It("should return the executable with the given verb and id", func() {
			exec, err := execs.FindByVerbAndID(exec1.Verb, exec1.ID())
			Expect(err).NotTo(HaveOccurred())
			Expect(exec).To(Equal(exec1))
		})

		It("should return nil if there is no match", func() {
			exec, err := execs.FindByVerbAndID("nonexistent", "nonexistent")
			Expect(err).To(HaveOccurred())
			Expect(exec).To(BeNil())
		})
	})
})

var _ = DescribeTable("HasFlowFileExt", func(file string, expected bool) {
	Expect(executable.HasFlowFileExt(file)).To(Equal(expected))
},
	Entry("ends with .flow", "development.flow", true),
	Entry("ends with .flow.yaml", "development.flow.yaml", true),
	Entry("ends with .flow.yml", "development.flow.yml", true),
	Entry("ends with .flow + something else", "development.flow.txt", false),
	Entry("ends with something else", "development.txt", false),
)

var _ = DescribeTable("HasFlowFileTemplateExt", func(file string, expected bool) {
	Expect(executable.HasFlowFileTemplateExt(file)).To(Equal(expected))
},
	Entry("ends with .flow.tmpl", "development.flow.tmpl", true),
	Entry("ends with .flow.tmpl.yaml", "development.flow.tmpl.yaml", true),
	Entry("ends with .flow.tmpl.yml", "development.flow.tmpl.yml", true),
	Entry("ends with .flow", "development.flow", false),
	Entry("ends with .flow.tmpl + something else", "development.flow.tmpl.txt", false),
	Entry("ends with something else", "development.flow.txt", false),
)

var _ = Describe("Executable Visibility", func() {
	It("should show public executables with any filter", func() {
		ws := "ws1"
		vis := common.VisibilityPublic
		ex := &executable.Executable{Visibility: (*executable.ExecutableVisibility)(&vis)}
		ex.SetContext(ws, "path", "", "file.flow")
		Expect(ex.IsVisibleFromWorkspace("ws2", common.VisibilityPublic)).To(BeTrue())
		Expect(ex.IsVisibleFromWorkspace("ws2", common.VisibilityPrivate)).To(BeTrue())
		Expect(ex.IsVisibleFromWorkspace("ws2", common.VisibilityInternal)).To(BeTrue())
	})

	It("should respect workspace scoping for private", func() {
		ws := "ws1"
		vis := common.VisibilityPrivate
		ex := &executable.Executable{Visibility: (*executable.ExecutableVisibility)(&vis)}
		ex.SetContext(ws, "path", "", "file.flow")
		Expect(ex.IsVisibleFromWorkspace(ws, common.VisibilityPrivate)).To(BeTrue())
		Expect(ex.IsVisibleFromWorkspace("ws2", common.VisibilityPrivate)).To(BeFalse())
	})

	It("should hide private with public filter", func() {
		ws := "ws1"
		vis := common.VisibilityPrivate
		ex := &executable.Executable{Visibility: (*executable.ExecutableVisibility)(&vis)}
		ex.SetContext(ws, "path", "", "file.flow")
		Expect(ex.IsVisibleFromWorkspace(ws, common.VisibilityPublic)).To(BeFalse())
	})

	It("should hide internal without internal+ filter", func() {
		ws := "ws1"
		vis := common.VisibilityInternal
		ex := &executable.Executable{Visibility: (*executable.ExecutableVisibility)(&vis)}
		ex.SetContext(ws, "path", "", "file.flow")
		Expect(ex.IsVisibleFromWorkspace(ws, common.VisibilityPrivate)).To(BeFalse())
		Expect(ex.IsVisibleFromWorkspace(ws, common.VisibilityInternal)).To(BeTrue())
	})

	It("should respect workspace scoping for internal", func() {
		ws := "ws1"
		vis := common.VisibilityInternal
		ex := &executable.Executable{Visibility: (*executable.ExecutableVisibility)(&vis)}
		ex.SetContext(ws, "path", "", "file.flow")
		Expect(ex.IsVisibleFromWorkspace("ws2", common.VisibilityInternal)).To(BeFalse())
	})

	It("should only show hidden with hidden filter", func() {
		ws := "ws1"
		vis := common.VisibilityHidden
		ex := &executable.Executable{Visibility: (*executable.ExecutableVisibility)(&vis)}
		ex.SetContext(ws, "path", "", "file.flow")
		Expect(ex.IsVisibleFromWorkspace(ws, common.VisibilityInternal)).To(BeFalse())
		Expect(ex.IsVisibleFromWorkspace(ws, common.VisibilityHidden)).To(BeTrue())
	})
})

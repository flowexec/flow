package env_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/internal/utils/env"
	"github.com/flowexec/flow/types/config"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

func TestEnv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Env Utility Suite")
}

var _ = Describe("Env", func() {
	var (
		ctrl       *gomock.Controller
		mockLogger *mocks.MockLogger
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockLogger = mocks.NewMockLogger(ctrl)
		logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("SetEnv", func() {
		When("EnvKey is specified", func() {
			It("should set environment variables correctly from params", func() {
				exec := &executable.ExecutableEnvironment{
					Params: []executable.Parameter{
						{EnvKey: "TEST_TEXT", Text: "test"},
						{EnvKey: "TEST_PROMPT", Prompt: "Enter value"},
						{EnvKey: "TEST_SECRET", SecretRef: "message"},
					},
				}
				promptedEnv := map[string]string{
					"TEST_PROMPT": "my value",
				}
				err := env.SetEnv("demo", exec, []string{}, promptedEnv)
				Expect(err).ToNot(HaveOccurred())
				val, exists := os.LookupEnv("TEST_TEXT")
				Expect(exists).To(BeTrue())
				Expect(val).To(Equal("test"))
				val, exists = os.LookupEnv("TEST_SECRET")
				Expect(exists).To(BeTrue())
				Expect(val).To(ContainSubstring("Thanks for trying flow!"))
				val, exists = os.LookupEnv("TEST_PROMPT")
				Expect(exists).To(BeTrue())
				Expect(val).To(Equal("my value"))
			})

			It("should set environment variables correctly from args", func() {
				pos := 1
				exec := &executable.ExecutableEnvironment{
					Args: []executable.Argument{
						{EnvKey: "TEST_POS", Pos: &pos},
						{EnvKey: "TEST_FLAG", Flag: "flag"},
					},
				}
				promptedEnv := make(map[string]string)
				err := env.SetEnv("", exec, []string{"test", "flag=value"}, promptedEnv)
				Expect(err).ToNot(HaveOccurred())
				val, exists := os.LookupEnv("TEST_POS")
				Expect(exists).To(BeTrue())
				Expect(val).To(Equal("test"))
				val, exists = os.LookupEnv("TEST_FLAG")
				Expect(exists).To(BeTrue())
				Expect(val).To(Equal("value"))
			})

			It("should use the input value when there is also an arg and param of the same key", func() {
				exec := &executable.ExecutableEnvironment{
					Params: []executable.Parameter{{EnvKey: "TEST_KEY", Text: "param"}},
					Args:   []executable.Argument{{EnvKey: "TEST_KEY", Flag: "flag"}},
				}
				promptedEnv := map[string]string{"TEST_KEY": "input"}
				err := env.SetEnv("", exec, []string{"flag=flag"}, promptedEnv)
				Expect(err).ToNot(HaveOccurred())
				val, exists := os.LookupEnv("TEST_KEY")
				Expect(exists).To(BeTrue())
				Expect(val).To(Equal("input"))
			})

			It("uses the args value when there is also a param of the same key", func() {
				exec := &executable.ExecutableEnvironment{
					Params: []executable.Parameter{{EnvKey: "TEST_KEY", Text: "param"}},
					Args:   []executable.Argument{{EnvKey: "TEST_KEY", Flag: "flag"}},
				}
				promptedEnv := map[string]string{"TEST_KEY": "input"}
				err := env.SetEnv("", exec, []string{"flag=flag"}, promptedEnv)
				Expect(err).ToNot(HaveOccurred())
				val, exists := os.LookupEnv("TEST_KEY")
				Expect(exists).To(BeTrue())
				Expect(val).To(Equal("input"))
			})
		})
	})

	Describe("CreateTempEnvFiles", func() {
		It("should create temporary files for parameters with OutputFile", func() {
			pos := 1
			exec := &executable.ExecutableEnvironment{
				Params: []executable.Parameter{
					{Text: "paramval", OutputFile: "//test_param_output.txt"},
				},
				Args: []executable.Argument{
					{Pos: &pos, OutputFile: "//test_arg_output.txt"},
				},
			}
			tmpDir := GinkgoTB().TempDir()
			cb, err := env.CreateTempEnvFiles("", "", tmpDir, exec, []string{"argval"}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cb).ToNot(BeNil())

			paramContent, err := os.ReadFile(filepath.Join(tmpDir, "test_param_output.txt"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(paramContent)).To(Equal("paramval"))

			argContent, err := os.ReadFile(filepath.Join(tmpDir, "test_arg_output.txt"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(argContent)).To(Equal("argval"))

			err = cb(nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = os.Stat(filepath.Join(tmpDir, "test_param_output.txt"))
			Expect(os.IsNotExist(err)).To(BeTrue())
			_, err = os.Stat(filepath.Join(tmpDir, "test_arg_output.txt"))
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})

	Describe("BuildArgsEnvMap", func() {
		It("should correctly parse flag arguments", func() {
			args := executable.ArgumentList{{EnvKey: "flag1", Flag: "flag1"}, {EnvKey: "flag2", Flag: "flag2"}}
			inputVals := []string{"flag1=value1", "flag2=value2"}
			envMap, err := env.BuildArgsEnvMap(args, inputVals, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(envMap).To(Equal(map[string]string{"flag1": "value1", "flag2": "value2"}))
		})

		It("should correctly parse positional arguments", func() {
			p1 := 1
			p2 := 2
			args := executable.ArgumentList{{EnvKey: "pos1", Pos: &p1}, {EnvKey: "pos2", Pos: &p2}}
			inputVals := []string{"pos1", "pos2"}
			envMap, err := env.BuildArgsEnvMap(args, inputVals, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(envMap).To(Equal(map[string]string{"pos1": "pos1", "pos2": "pos2"}))
		})

		It("should correctly parse mixed arguments", func() {
			p1 := 1
			args := executable.ArgumentList{{EnvKey: "flag1", Flag: "flag1"}, {EnvKey: "pos1", Pos: &p1}}
			inputVals := []string{"flag1=value1", "pos1"}
			envMap, err := env.BuildArgsEnvMap(args, inputVals, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(envMap).To(Equal(map[string]string{"flag1": "value1", "pos1": "pos1"}))
		})

		It("should correctly parse flag arguments with equal sign in value", func() {
			args := executable.ArgumentList{{EnvKey: "flag1", Flag: "flag1"}}
			inputVals := []string{"flag1=value1=value2"}
			envMap, err := env.BuildArgsEnvMap(args, inputVals, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(envMap).To(Equal(map[string]string{"flag1": "value1=value2"}))
		})
	})

	Describe("ResolveParameterValue", func() {
		It("should return empty string when all parameter fields are empty", func() {
			param := executable.Parameter{}
			promptedEnv := make(map[string]string)
			val, err := env.ResolveParameterValue("", param, promptedEnv)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(""))
		})

		It("should return text when text field is not empty", func() {
			param := executable.Parameter{
				Text: "test",
			}
			promptedEnv := make(map[string]string)
			val, err := env.ResolveParameterValue("", param, promptedEnv)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("test"))
		})

		It("should return error when prompt field is not empty but env key is not in promptedEnv", func() {
			param := executable.Parameter{
				Prompt: "test",
				EnvKey: "TEST_KEY",
			}
			promptedEnv := make(map[string]string)
			_, err := env.ResolveParameterValue("", param, promptedEnv)
			Expect(err).To(HaveOccurred())
		})

		It("should return value from promptedEnv when prompt field is not empty and env key is in promptedEnv", func() {
			param := executable.Parameter{
				Prompt: "test",
				EnvKey: "TEST_KEY",
			}
			promptedEnv := map[string]string{
				"TEST_KEY": "test",
			}
			val, err := env.ResolveParameterValue("", param, promptedEnv)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("test"))
		})

		// TODO: Add test cases for SecretRef
	})

	Describe("EnvMapToEnvList", func() {
		It("should convert the map to list correctly", func() {
			envMap := map[string]string{"TEST_KEY": "test", "TEST_KEY_2": "test2"}
			envList := env.EnvMapToEnvList(envMap)
			Expect(envList).To(Equal([]string{"TEST_KEY=test", "TEST_KEY_2=test2"}))
		})
	})

	Describe("EnvListToEnvMap", func() {
		It("should convert the list to map correctly", func() {
			envList := []string{"TEST_KEY=test", "TEST_KEY_2=test2"}
			envMap := env.EnvListToEnvMap(envList)
			Expect(envMap).To(Equal(map[string]string{"TEST_KEY": "test", "TEST_KEY_2": "test2"}))
		})

		It("should handle malformed entries gracefully", func() {
			envList := []string{"TEST_KEY=test", "INVALID_ENTRY"}
			envMap := env.EnvListToEnvMap(envList)
			Expect(envMap).To(Equal(map[string]string{"TEST_KEY": "test"}))
		})
	})

	Describe("BuildEnvMap", func() {
		It("should build environment map correctly", func() {
			exec := &executable.ExecutableEnvironment{
				Params: []executable.Parameter{
					{
						EnvKey: "TEST_KEY",
						Text:   "test",
					},
					{
						EnvKey: "TEST_KEY_2",
						Text:   "test2",
					},
				},
				Args: []executable.Argument{
					{
						EnvKey: "TEST_KEY_3",
						Flag:   "flag",
					},
				},
			}
			inputEnv := make(map[string]string)
			defaultEnv := make(map[string]string)
			envMap, err := env.BuildEnvMap("", exec, []string{"flag=test3"}, inputEnv, defaultEnv)
			Expect(err).ToNot(HaveOccurred())
			Expect(envMap).To(Equal(map[string]string{"TEST_KEY": "test", "TEST_KEY_2": "test2", "TEST_KEY_3": "test3"}))
		})
	})

	Describe("DefaultEnv", func() {
		It("should return default environment correctly", func() {
			wsName := "test-workspace"
			wsLocation := "test-location"
			nsName := "test-namespace"
			execName := "test-executable"
			execDefinitionLocation := "test-definition-location"

			ws := workspace.Workspace{}
			ws.SetContext(wsName, wsLocation)
			ctx := &context.Context{
				CurrentWorkspace: &ws,
				Config: &config.Config{
					CurrentNamespace: nsName,
				},
			}
			exec := executable.Executable{Name: execName}
			exec.SetContext(wsName, wsLocation, nsName, execDefinitionLocation)
			envMap := env.DefaultEnv(ctx, &exec)
			Expect(envMap["FLOW_RUNNER"]).To(Equal("true"))
			Expect(envMap["FLOW_CURRENT_WORKSPACE"]).To(Equal(wsName))
			Expect(envMap["FLOW_CURRENT_NAMESPACE"]).To(Equal(nsName))
			Expect(envMap["FLOW_EXECUTABLE_NAME"]).To(Equal(execName))
			Expect(envMap["DISABLE_FLOW_INTERACTIVE"]).To(Equal("true"))
			// TODO: Add more assertions for other keys in the environment map
		})
	})
})

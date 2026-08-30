package request_test

import (
	stdCtx "context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/v2/internal/runner"
	"github.com/flowexec/flow/v2/internal/runner/engine/mocks"
	"github.com/flowexec/flow/v2/internal/runner/request"
	testUtils "github.com/flowexec/flow/v2/tests/utils"
	"github.com/flowexec/flow/v2/types/executable"
)

func TestRequest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Request Suite")
}

var _ = Describe("Request Runner", func() {
	var (
		requestRnr runner.Runner
		ctx        *testUtils.ContextWithMocks
		mockEngine *mocks.MockEngine
	)

	BeforeEach(func() {
		ctx = testUtils.NewContextWithMocks(stdCtx.Background(), GinkgoTB())
		requestRnr = request.NewRunner()
		ctrl := gomock.NewController(GinkgoT())
		mockEngine = mocks.NewMockEngine(ctrl)
	})

	Context("Name", func() {
		It("should return the correct requestRnr name", func() {
			Expect(requestRnr.Name()).To(Equal("request"))
		})
	})

	Context("IsCompatible", func() {
		It("should return false when executable is nil", func() {
			Expect(requestRnr.IsCompatible(nil)).To(BeFalse())
		})

		It("should return false when executable type is nil", func() {
			executable := &executable.Executable{}
			Expect(requestRnr.IsCompatible(executable)).To(BeFalse())
		})

		It("should return true when executable type is serial", func() {
			executable := &executable.Executable{
				Request: &executable.RequestExecutableType{},
			}
			Expect(requestRnr.IsCompatible(executable)).To(BeTrue())
		})
	})

	Describe("Exec", func() {
		var testServer *httptest.Server
		BeforeEach(func() {
			testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"message": "GET request successful"}`))
				case http.MethodPost:
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"key": "value"}`))
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}))
		})

		Describe("body", func() {
			var received string
			var bodyServer *httptest.Server

			BeforeEach(func() {
				received = ""
				bodyServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					received = string(b)
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"ok": true}`))
				}))
			})

			AfterEach(func() { bodyServer.Close() })

			sendBody := func(body string, envMap map[string]string) {
				exec := &executable.Executable{
					Request: &executable.RequestExecutableType{
						URL:    bodyServer.URL,
						Method: executable.RequestExecutableTypeMethodPOST,
						Body:   body,
					},
				}
				ctx.Logger.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(1)
				Expect(requestRnr.Exec(ctx.Ctx, exec, mockEngine, envMap, nil)).To(Succeed())
			}

			It("should send a JSON object body as written", func() {
				sendBody(
					"{\n  \"environment\": \"$ENVIRONMENT\",\n  \"version\": \"$VERSION\"\n}",
					map[string]string{"ENVIRONMENT": "prod", "VERSION": "1.0"},
				)
				Expect(received).To(MatchJSON(`{"environment": "prod", "version": "1.0"}`))
			})

			It("should send a nested JSON body as written", func() {
				sendBody(`{"messages": [{"role": "user"}]}`, map[string]string{})
				Expect(received).To(MatchJSON(`{"messages": [{"role": "user"}]}`))
			})

			It("should evaluate a non-JSON body as an expression", func() {
				sendBody(`'{"model":' + toJSON(env["MODEL"]) + '}'`, map[string]string{"MODEL": `a "quoted" name`})
				Expect(received).To(MatchJSON(`{"model": "a \"quoted\" name"}`))
			})

			It("should send a JSON array body as written", func() {
				sendBody(`[1, 2, 3]`, map[string]string{})
				Expect(received).To(Equal(`[1, 2, 3]`))
			})
		})

		It("should send a GET request and log the response", func() {
			exec := &executable.Executable{
				Request: &executable.RequestExecutableType{
					URL:         testServer.URL,
					Method:      executable.RequestExecutableTypeMethodGET,
					LogResponse: true,
				},
			}

			ctx.Logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			err := requestRnr.Exec(ctx.Ctx, exec, mockEngine, make(map[string]string), nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should send a POST request with a body and log the response", func() {
			exec := &executable.Executable{
				Request: &executable.RequestExecutableType{
					URL:         testServer.URL,
					Method:      executable.RequestExecutableTypeMethodPOST,
					Body:        `{"key": "value"}`,
					LogResponse: true,
				},
			}

			ctx.Logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Regex("value")).Times(1)
			err := requestRnr.Exec(ctx.Ctx, exec, mockEngine, make(map[string]string), nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should save the response to a file", func() {
			exec := &executable.Executable{
				Request: &executable.RequestExecutableType{
					URL:    testServer.URL,
					Method: executable.RequestExecutableTypeMethodGET,
					ResponseFile: &executable.RequestResponseFile{
						Filename: "response.json",
						Dir:      executable.Directory("//"),
						SaveAs:   executable.RequestResponseFileSaveAsJson,
					},
				},
			}
			exec.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), ctx.Ctx.CurrentWorkspace.Location(), "", "")

			ctx.Logger.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(2)
			err := requestRnr.Exec(ctx.Ctx, exec, mockEngine, make(map[string]string), nil)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(filepath.Clean(filepath.Join(ctx.Ctx.CurrentWorkspace.Location(), "response.json")))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should transform the response when specified", func() {
			exec := &executable.Executable{
				Request: &executable.RequestExecutableType{
					URL:               testServer.URL,
					Method:            executable.RequestExecutableTypeMethodGET,
					TransformResponse: `upper(body)`,
					LogResponse:       true,
				},
			}

			ctx.Logger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Regex("SUCCESSFUL")).Times(1)
			err := requestRnr.Exec(ctx.Ctx, exec, mockEngine, make(map[string]string), nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

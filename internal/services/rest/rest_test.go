package rest_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/internal/services/rest"
)

func TestRest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rest Suite")
}

var _ = Describe("Rest", func() {
	var testServer *httptest.Server

	BeforeEach(func() {
		testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()

			if code := query.Get("code"); code != "" {
				if statusCode, err := strconv.Atoi(code); err == nil {
					w.WriteHeader(statusCode)
					return
				}
			}
			if sleep := query.Get("sleep"); sleep != "" {
				if duration, err := time.ParseDuration(sleep); err == nil {
					time.Sleep(duration)
				}
			}
			if query.Get("print-headers") == "true" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{"headers": {"%s": "%s"}}`,
					"Test-Header", r.Header.Get("Test-Header"))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`success`))
		}))
	})

	Context("SendRequest", func() {
		It("should return error when invalid URL is provided", func() {
			req := &rest.Request{
				URL:     "invalid_url",
				Method:  "GET",
				Timeout: 30 * time.Second,
			}
			_, err := rest.SendRequest(req, []int{http.StatusOK})
			Expect(err).To(HaveOccurred())
		})

		It("should return error when unexpected status code is received", func() {
			req := &rest.Request{
				URL:     testServer.URL + "?code=500",
				Method:  "GET",
				Timeout: 30 * time.Second,
			}
			_, err := rest.SendRequest(req, []int{http.StatusOK})
			Expect(err).To(Equal(rest.ErrUnexpectedStatusCode))
		})

		It("should return the correct body when a valid request is made", func() {
			req := &rest.Request{
				URL:     testServer.URL,
				Method:  "GET",
				Timeout: 30 * time.Second,
			}
			resp, err := rest.SendRequest(req, []int{http.StatusOK})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Body).To(Equal("success"))
		})

		It("should timeout when the request takes longer than the specified timeout", func() {
			req := &rest.Request{
				URL:     testServer.URL + "?sleep=2s",
				Method:  "GET",
				Timeout: 1 * time.Second,
			}
			_, err := rest.SendRequest(req, []int{http.StatusOK})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Client.Timeout exceeded while awaiting headers"))
		})

		It("should return the correct headers when a valid request is made", func() {
			req := &rest.Request{
				URL:     testServer.URL + "?print-headers=true",
				Method:  "GET",
				Headers: map[string]string{"Test-Header": "Test-Value"},
				Timeout: 30 * time.Second,
			}
			resp, err := rest.SendRequest(req, []int{http.StatusOK})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Body).To(ContainSubstring("\"Test-Header\": \"Test-Value\""))
		})
	})
})

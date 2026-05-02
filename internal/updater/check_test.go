package updater_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/v2/internal/updater"
	storemocks "github.com/flowexec/flow/v2/pkg/store/mocks"
)

var _ = Describe("IsNewer", func() {
	DescribeTable("version comparisons",
		func(current, latest string, want bool) {
			Expect(updater.IsNewer(current, latest)).To(Equal(want))
		},
		Entry("newer patch", "1.0.0", "1.0.1", true),
		Entry("same version", "1.0.0", "1.0.0", false),
		Entry("current is newer", "2.0.0", "1.9.9", false),
		Entry("v-prefixed tag", "1.0.0", "v1.0.1", true),
		Entry("invalid version strings", "bad", "also-bad", false),
	)
})

var _ = Describe("IsDisabled", func() {
	const noCheckEnvVar = "FLOW_NO_UPDATE_CHECK"
	var original string

	BeforeEach(func() { original = os.Getenv(noCheckEnvVar) })
	AfterEach(func() { _ = os.Setenv(noCheckEnvVar, original) })

	It("returns false when env var is not set", func() {
		_ = os.Unsetenv(noCheckEnvVar)
		Expect(updater.IsDisabled()).To(BeFalse())
	})

	It("returns true for a non-empty value", func() {
		_ = os.Setenv(noCheckEnvVar, "1")
		Expect(updater.IsDisabled()).To(BeTrue())
	})

	It("returns false for '0'", func() {
		_ = os.Setenv(noCheckEnvVar, "0")
		Expect(updater.IsDisabled()).To(BeFalse())
	})
})

var _ = Describe("CachedUpdateNotice", func() {
	var (
		ctrl    *gomock.Controller
		mockDS  *storemocks.MockDataStore
		origVer func() string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDS = storemocks.NewMockDataStore(ctrl)
		origVer = *updater.CurrentSemVer
		*updater.CurrentSemVer = func() string { return "1.0.0" }
		_ = os.Unsetenv("FLOW_NO_UPDATE_CHECK")
	})

	AfterEach(func() { *updater.CurrentSemVer = origVer })

	It("returns empty string when cache is empty or errors", func() {
		mockDS.EXPECT().GetCacheEntry(updater.CacheKey).Return(nil, errors.New("db error"))
		Expect(updater.CachedUpdateNotice(mockDS)).To(BeEmpty())
	})

	It("returns empty string when cached version is not newer", func() {
		data, _ := json.Marshal(updater.ReleaseInfo{TagName: "v1.0.0"})
		mockDS.EXPECT().GetCacheEntry(updater.CacheKey).Return(data, nil)
		Expect(updater.CachedUpdateNotice(mockDS)).To(BeEmpty())
	})

	It("returns a notice containing both versions when an update is available", func() {
		data, _ := json.Marshal(updater.ReleaseInfo{TagName: "v1.1.0"})
		mockDS.EXPECT().GetCacheEntry(updater.CacheKey).Return(data, nil)
		notice := updater.CachedUpdateNotice(mockDS)
		Expect(notice).To(ContainSubstring("v1.1.0"))
		Expect(notice).To(ContainSubstring("1.0.0"))
		Expect(notice).To(ContainSubstring("flow cli update"))
	})

	It("returns empty string when disabled or version is unknown", func() {
		*updater.CurrentSemVer = func() string { return "" }
		Expect(updater.CachedUpdateNotice(mockDS)).To(BeEmpty())
	})
})

var _ = Describe("RefreshCache", func() {
	var (
		ctrl   *gomock.Controller
		mockDS *storemocks.MockDataStore
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDS = storemocks.NewMockDataStore(ctrl)
	})

	It("stores marshaled release info and sets CheckedAt", func() {
		before := time.Now()
		var stored []byte
		mockDS.EXPECT().SetCacheEntry(updater.CacheKey, gomock.Any()).DoAndReturn(
			func(_ string, data []byte) error {
				stored = data
				return nil
			},
		)

		Expect(updater.RefreshCache(mockDS, &updater.ReleaseInfo{TagName: "v2.0.0"})).To(Succeed())

		var decoded updater.ReleaseInfo
		Expect(json.Unmarshal(stored, &decoded)).To(Succeed())
		Expect(decoded.TagName).To(Equal("v2.0.0"))
		Expect(decoded.CheckedAt).To(BeTemporally(">=", before))
	})

	It("propagates SetCacheEntry errors", func() {
		mockDS.EXPECT().SetCacheEntry(updater.CacheKey, gomock.Any()).Return(errors.New("write error"))
		Expect(updater.RefreshCache(mockDS, &updater.ReleaseInfo{TagName: "v2.0.0"})).To(HaveOccurred())
	})
})

var _ = Describe("CheckInBackground", func() {
	var (
		ctrl    *gomock.Controller
		mockDS  *storemocks.MockDataStore
		origVer func() string
		origURL string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDS = storemocks.NewMockDataStore(ctrl)
		origVer = *updater.CurrentSemVer
		origURL = *updater.GithubReleaseURL
		*updater.CurrentSemVer = func() string { return "1.0.0" }
		_ = os.Unsetenv("FLOW_NO_UPDATE_CHECK")
	})

	AfterEach(func() {
		*updater.CurrentSemVer = origVer
		*updater.GithubReleaseURL = origURL
	})

	It("is a no-op when enabled is false", func() {
		updater.CheckInBackground(mockDS, false) // no mock expectations; any call would fail
	})

	It("is a no-op when version is unknown", func() {
		*updater.CurrentSemVer = func() string { return "" }
		updater.CheckInBackground(mockDS, true)
	})

	It("is a no-op when cache is still fresh", func() {
		fresh, _ := json.Marshal(updater.ReleaseInfo{
			TagName:   "v1.1.0",
			CheckedAt: time.Now().Add(-1 * time.Hour),
		})
		mockDS.EXPECT().GetCacheEntry(updater.CacheKey).Return(fresh, nil)
		updater.CheckInBackground(mockDS, true)
	})

	It("fires a background fetch and updates cache when stale", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(updater.ReleaseInfo{TagName: "v2.0.0"})
		}))
		defer server.Close()
		*updater.GithubReleaseURL = server.URL

		stale, _ := json.Marshal(updater.ReleaseInfo{
			TagName:   "v1.1.0",
			CheckedAt: time.Now().Add(-25 * time.Hour),
		})
		mockDS.EXPECT().GetCacheEntry(updater.CacheKey).Return(stale, nil)

		setCalled := make(chan struct{})
		mockDS.EXPECT().SetCacheEntry(updater.CacheKey, gomock.Any()).DoAndReturn(
			func(_ string, _ []byte) error { close(setCalled); return nil },
		)

		updater.CheckInBackground(mockDS, true)
		Eventually(setCalled, 3*time.Second).Should(BeClosed())
	})
})

var _ = Describe("LatestRelease", func() {
	var origURL string

	BeforeEach(func() { origURL = *updater.GithubReleaseURL })
	AfterEach(func() { *updater.GithubReleaseURL = origURL })

	It("returns release info on a 200 response and sends required headers", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.Header.Get("Accept")).To(Equal("application/vnd.github+json"))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(updater.ReleaseInfo{TagName: "v3.0.0"})
		}))
		defer server.Close()
		*updater.GithubReleaseURL = server.URL

		info, err := updater.LatestRelease()
		Expect(err).NotTo(HaveOccurred())
		Expect(info.TagName).To(Equal("v3.0.0"))
	})

	It("returns an error on a non-200 response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()
		*updater.GithubReleaseURL = server.URL

		_, err := updater.LatestRelease()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("500"))
	})
})

var _ = Describe("ReleaseByTag", func() {
	var origBase string

	BeforeEach(func() { origBase = *updater.GithubReleaseTagBaseURL })
	AfterEach(func() { *updater.GithubReleaseTagBaseURL = origBase })

	It("returns release info and normalises a missing v prefix", func() {
		var capturedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(updater.ReleaseInfo{TagName: "v2.1.0"})
		}))
		defer server.Close()
		*updater.GithubReleaseTagBaseURL = server.URL

		info, err := updater.ReleaseByTag("2.1.0") // no v prefix
		Expect(err).NotTo(HaveOccurred())
		Expect(info.TagName).To(Equal("v2.1.0"))
		Expect(capturedPath).To(ContainSubstring("v2.1.0"))
	})

	It("returns an error on a 404 response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()
		*updater.GithubReleaseTagBaseURL = server.URL

		_, err := updater.ReleaseByTag("v9.9.9")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("404"))
	})
})

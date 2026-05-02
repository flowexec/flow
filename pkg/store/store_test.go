package store_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/pkg/store"
)

func TestStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DataStore Suite")
}

var _ = Describe("BoltDataStore", func() {
	var ds store.DataStore
	var err error

	BeforeEach(func() {
		path := filepath.Join(GinkgoT().TempDir(), fmt.Sprintf("test_%s.db", GinkgoT().Name()))
		ds, err = store.NewDataStore(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(ds).NotTo(BeNil())
	})

	AfterEach(func() {
		Expect(ds.Close()).To(Succeed())
	})

	Describe("Cache operations", func() {
		It("should set and get a cache entry", func() {
			Expect(ds.SetCacheEntry("key", []byte("value"))).To(Succeed())

			val, err := ds.GetCacheEntry("key")
			Expect(err).NotTo(HaveOccurred())
			Expect(val).To(Equal([]byte("value")))
		})

		It("should return nil for a missing cache entry", func() {
			val, err := ds.GetCacheEntry("missing")
			Expect(err).NotTo(HaveOccurred())
			Expect(val).To(BeNil())
		})

		It("should overwrite an existing cache entry", func() {
			Expect(ds.SetCacheEntry("key", []byte("v1"))).To(Succeed())
			Expect(ds.SetCacheEntry("key", []byte("v2"))).To(Succeed())

			val, err := ds.GetCacheEntry("key")
			Expect(err).NotTo(HaveOccurred())
			Expect(val).To(Equal([]byte("v2")))
		})

		It("should delete a cache entry", func() {
			Expect(ds.SetCacheEntry("key", []byte("value"))).To(Succeed())
			Expect(ds.DeleteCacheEntry("key")).To(Succeed())

			val, err := ds.GetCacheEntry("key")
			Expect(err).NotTo(HaveOccurred())
			Expect(val).To(BeNil())
		})

		It("should not error when deleting a missing cache entry", func() {
			Expect(ds.DeleteCacheEntry("missing")).To(Succeed())
		})
	})

	Describe("Execution history operations", func() {
		var ref = "ws/ns:exec"

		It("should record and retrieve an execution", func() {
			rec := store.ExecutionRecord{
				Ref:          ref,
				StartedAt:    time.Now().UTC().Truncate(time.Millisecond),
				Duration:     500 * time.Millisecond,
				ExitCode:     0,
				LogArchiveID: "archive-123",
			}
			Expect(ds.RecordExecution(rec)).To(Succeed())

			history, err := ds.GetExecutionHistory(ref, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(history).To(HaveLen(1))
			Expect(history[0].Ref).To(Equal(ref))
			Expect(history[0].ExitCode).To(Equal(0))
			Expect(history[0].LogArchiveID).To(Equal("archive-123"))
		})

		It("should return empty history for an unknown ref", func() {
			history, err := ds.GetExecutionHistory("unknown/ns:exec", 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(history).To(BeEmpty())
		})

		It("should respect the limit parameter", func() {
			for i := range 5 {
				Expect(ds.RecordExecution(store.ExecutionRecord{
					Ref:      ref,
					ExitCode: i,
				})).To(Succeed())
			}

			history, err := ds.GetExecutionHistory(ref, 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(history).To(HaveLen(3))
			// Most recent 3 entries (exit codes 2, 3, 4)
			Expect(history[0].ExitCode).To(Equal(2))
			Expect(history[2].ExitCode).To(Equal(4))
		})

		It("should record execution failure with error message", func() {
			rec := store.ExecutionRecord{
				Ref:      ref,
				ExitCode: 1,
				Error:    "command not found",
			}
			Expect(ds.RecordExecution(rec)).To(Succeed())

			history, err := ds.GetExecutionHistory(ref, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(history).To(HaveLen(1))
			Expect(history[0].Error).To(Equal("command not found"))
		})

		It("should delete all history for a ref", func() {
			Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref})).To(Succeed())
			Expect(ds.DeleteExecutionHistory(ref)).To(Succeed())

			history, err := ds.GetExecutionHistory(ref, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(history).To(BeEmpty())
		})

		It("should not error when deleting history for an unknown ref", func() {
			Expect(ds.DeleteExecutionHistory("unknown/ns:exec")).To(Succeed())
		})

		It("should maintain separate history per ref", func() {
			ref2 := "ws/ns:other"
			Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref, ExitCode: 0})).To(Succeed())
			Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref2, ExitCode: 1})).To(Succeed())

			h1, err := ds.GetExecutionHistory(ref, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(h1).To(HaveLen(1))
			Expect(h1[0].ExitCode).To(Equal(0))

			h2, err := ds.GetExecutionHistory(ref2, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(h2).To(HaveLen(1))
			Expect(h2[0].ExitCode).To(Equal(1))
		})
	})
})

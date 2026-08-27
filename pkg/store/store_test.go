package store_test

import (
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
		// TempDir is already unique per spec, so the file needs no disambiguating
		// suffix - and deriving one from the spec name broke on Windows, where a
		// name containing "->" yields an illegal filename (ERROR_INVALID_NAME).
		path := filepath.Join(GinkgoT().TempDir(), "test.db")
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

		It("should bulk read cache entries, omitting missing keys", func() {
			Expect(ds.SetCacheEntry("a", []byte("1"))).To(Succeed())
			Expect(ds.SetCacheEntry("b", []byte("2"))).To(Succeed())

			entries, err := ds.GetCacheEntries([]string{"a", "b", "missing"})
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(2))
			Expect(entries["a"]).To(Equal([]byte("1")))
			Expect(entries["b"]).To(Equal([]byte("2")))
			Expect(entries).NotTo(HaveKey("missing"))
		})

		It("should return an empty map when bulk reading with no cache bucket", func() {
			entries, err := ds.GetCacheEntries([]string{"a"})
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(BeEmpty())
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

		It("should upsert records that share an ID (running -> terminal)", func() {
			start := time.Now().UTC().Truncate(time.Millisecond)
			Expect(ds.RecordExecution(store.ExecutionRecord{
				ID:        "run-1",
				Ref:       ref,
				StartedAt: start,
				Status:    store.RunRunning,
				PID:       4242,
			})).To(Succeed())

			// Same ID upserts in place rather than appending a second record.
			done := start.Add(time.Second)
			Expect(ds.RecordExecution(store.ExecutionRecord{
				ID:          "run-1",
				Ref:         ref,
				StartedAt:   start,
				CompletedAt: &done,
				Duration:    time.Second,
				Status:      store.RunCompleted,
			})).To(Succeed())

			history, err := ds.GetExecutionHistory(ref, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(history).To(HaveLen(1))
			Expect(history[0].Status).To(Equal(store.RunCompleted))
			Expect(history[0].CompletedAt).NotTo(BeNil())
		})

		It("should append records without an ID (legacy behavior)", func() {
			Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref})).To(Succeed())
			Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref})).To(Succeed())

			history, err := ds.GetExecutionHistory(ref, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(history).To(HaveLen(2))
		})

		It("should bulk read history for multiple refs, omitting refs with no history", func() {
			ref2 := "ws/ns:other"
			Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref, ExitCode: 0})).To(Succeed())
			Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref2, ExitCode: 1})).To(Succeed())

			byRef, err := ds.GetExecutionHistories([]string{ref, ref2, "ws/ns:absent"}, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(byRef).To(HaveLen(2))
			Expect(byRef[ref]).To(HaveLen(1))
			Expect(byRef[ref2]).To(HaveLen(1))
			Expect(byRef).NotTo(HaveKey("ws/ns:absent"))
		})

		It("should apply the per-ref limit in bulk history reads", func() {
			for i := range 5 {
				Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref, ExitCode: i})).To(Succeed())
			}

			byRef, err := ds.GetExecutionHistories([]string{ref}, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(byRef[ref]).To(HaveLen(2))
			// Most recent 2 entries (exit codes 3, 4).
			Expect(byRef[ref][0].ExitCode).To(Equal(3))
			Expect(byRef[ref][1].ExitCode).To(Equal(4))
		})

		It("should read all execution history across refs in one call", func() {
			ref2 := "ws/ns:other"
			Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref, ExitCode: 0})).To(Succeed())
			Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref, ExitCode: 1})).To(Succeed())
			Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref2, ExitCode: 2})).To(Succeed())

			byRef, err := ds.GetAllExecutionHistory(0)
			Expect(err).NotTo(HaveOccurred())
			Expect(byRef).To(HaveLen(2))
			Expect(byRef[ref]).To(HaveLen(2))
			Expect(byRef[ref2]).To(HaveLen(1))
		})

		It("should apply the per-ref limit in GetAllExecutionHistory", func() {
			for i := range 4 {
				Expect(ds.RecordExecution(store.ExecutionRecord{Ref: ref, ExitCode: i})).To(Succeed())
			}

			byRef, err := ds.GetAllExecutionHistory(1)
			Expect(err).NotTo(HaveOccurred())
			Expect(byRef[ref]).To(HaveLen(1))
			Expect(byRef[ref][0].ExitCode).To(Equal(3))
		})

		It("should return an empty map from GetAllExecutionHistory when no history exists", func() {
			byRef, err := ds.GetAllExecutionHistory(0)
			Expect(err).NotTo(HaveOccurred())
			Expect(byRef).To(BeEmpty())
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

	Describe("Background run operations", func() {
		It("should bulk read background runs by ID, omitting missing IDs", func() {
			Expect(ds.SaveBackgroundRun(store.BackgroundRun{ID: "r1", Status: store.BackgroundRunning})).To(Succeed())
			Expect(ds.SaveBackgroundRun(store.BackgroundRun{ID: "r2", Status: store.BackgroundCompleted})).To(Succeed())

			runs, err := ds.GetBackgroundRuns([]string{"r1", "r2", "missing"})
			Expect(err).NotTo(HaveOccurred())
			Expect(runs).To(HaveLen(2))
			Expect(runs["r1"].Status).To(Equal(store.BackgroundRunning))
			Expect(runs["r2"].Status).To(Equal(store.BackgroundCompleted))
			Expect(runs).NotTo(HaveKey("missing"))
		})

		It("should return an empty map when bulk reading with no background bucket", func() {
			runs, err := ds.GetBackgroundRuns([]string{"r1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(runs).To(BeEmpty())
		})
	})
})

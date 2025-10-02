package engine_test

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/internal/runner/engine"
)

func TestEngine_Execute(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Execute Engine Suite")
}

var _ = Describe("e.Execute", func() {
	var (
		eng    engine.Engine
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		eng = engine.NewExecEngine()
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
	})

	Context("Parallel execution", func() {
		It("should execute execs in parallel", func() {
			execs := []engine.Exec{
				{ID: "exec1", Function: func() error { time.Sleep(100 * time.Millisecond); return nil }},
				{ID: "exec2", Function: func() error { return nil }},
			}

			start := time.Now()
			ff := false
			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Parallel), engine.WithFailFast(&ff))
			duration := time.Since(start)

			Expect(summary.Results).To(HaveLen(2))
			Expect(summary.Results[0].Error).NotTo(HaveOccurred())
			Expect(summary.Results[1].Error).NotTo(HaveOccurred())
			Expect(duration).To(BeNumerically("<", 200*time.Millisecond))
		})

		It("should handle exec failures with fail fast", func() {
			execs := []engine.Exec{
				{ID: "exec1", Function: func() error { return errors.New("error") }},
				{ID: "exec2", Function: func() error { time.Sleep(100 * time.Millisecond); return nil }},
			}

			ff := true
			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Parallel), engine.WithFailFast(&ff))

			Expect(summary.Results).To(HaveLen(2))
			Expect(summary.Results[0].Error).To(HaveOccurred())
			Expect(summary.Results[1].Error).ToNot(HaveOccurred())
			Expect(summary.HasErrors()).To(BeTrue())
		})

		It("should limit the number of concurrent execs", func() {
			execs := []engine.Exec{
				{ID: "exec1", Function: func() error { time.Sleep(100 * time.Millisecond); return nil }},
				{ID: "exec2", Function: func() error { time.Sleep(100 * time.Millisecond); return nil }},
				{ID: "exec3", Function: func() error { time.Sleep(100 * time.Millisecond); return nil }},
				{ID: "exec4", Function: func() error { time.Sleep(100 * time.Millisecond); return nil }},
				{ID: "exec5", Function: func() error { time.Sleep(100 * time.Millisecond); return nil }},
			}

			start := time.Now()
			ff := false
			summary := eng.Execute(ctx, execs,
				engine.WithMode(engine.Parallel), engine.WithFailFast(&ff), engine.WithMaxThreads(2))
			duration := time.Since(start)

			Expect(summary.Results).To(HaveLen(5))
			Expect(summary.Results[0].Error).NotTo(HaveOccurred())
			Expect(summary.Results[1].Error).NotTo(HaveOccurred())
			Expect(summary.Results[2].Error).NotTo(HaveOccurred())
			Expect(summary.Results[3].Error).NotTo(HaveOccurred())
			Expect(summary.Results[4].Error).NotTo(HaveOccurred())
			Expect(duration).To(BeNumerically(">=", 250*time.Millisecond))
		})

		It("should skip exec when condition returns false", func() {
			execs := []engine.Exec{
				{
					ID:       "exec1",
					Function: func() error { return nil },
				},
				{
					ID:        "exec2",
					Function:  func() error { return nil },
					Condition: func() (bool, error) { return false, nil },
				},
				{
					ID:       "exec3",
					Function: func() error { return nil },
				},
			}

			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Parallel))

			// Results array has fixed size, but skipped execs have zero-value Result
			Expect(summary.Results).To(HaveLen(3))
			Expect(summary.Results[0].Error).NotTo(HaveOccurred())
			Expect(summary.Results[0].ID).To(Equal("exec1"))
			// exec2 was skipped but still has entry (zero-value)
			Expect(summary.Results[1].ID).To(Equal(""))
			Expect(summary.Results[2].Error).NotTo(HaveOccurred())
			Expect(summary.Results[2].ID).To(Equal("exec3"))
		})

		It("should execute when condition returns true", func() {
			execs := []engine.Exec{
				{
					ID:        "exec1",
					Function:  func() error { return nil },
					Condition: func() (bool, error) { return true, nil },
				},
				{
					ID:        "exec2",
					Function:  func() error { return nil },
					Condition: func() (bool, error) { return true, nil },
				},
			}

			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Parallel))

			Expect(summary.Results).To(HaveLen(2))
			Expect(summary.Results[0].Error).NotTo(HaveOccurred())
			Expect(summary.Results[1].Error).NotTo(HaveOccurred())
		})

		It("should fail when condition evaluation fails", func() {
			execs := []engine.Exec{
				{
					ID:       "exec1",
					Function: func() error { return nil },
				},
				{
					ID:        "exec2",
					Function:  func() error { return nil },
					Condition: func() (bool, error) { return false, errors.New("condition error") },
				},
			}

			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Parallel))

			Expect(summary.Results).To(HaveLen(2))
			Expect(summary.Results[0].Error).NotTo(HaveOccurred())
			Expect(summary.Results[1].Error).To(HaveOccurred())
			Expect(summary.Results[1].Error.Error()).To(ContainSubstring("condition evaluation failed"))
			Expect(summary.HasErrors()).To(BeTrue())
		})

		It("should stop on condition error with fail fast", func() {
			execs := []engine.Exec{
				{
					ID:        "exec1",
					Function:  func() error { time.Sleep(100 * time.Millisecond); return nil },
					Condition: func() (bool, error) { return false, errors.New("condition error") },
				},
				{
					ID:       "exec2",
					Function: func() error { time.Sleep(100 * time.Millisecond); return nil },
				},
			}

			ff := true
			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Parallel), engine.WithFailFast(&ff))

			Expect(summary.Results).To(HaveLen(2))
			Expect(summary.HasErrors()).To(BeTrue())
		})
	})

	Context("Serial execution", func() {
		It("should execute execs serially", func() {
			execs := []engine.Exec{
				{ID: "exec1", Function: func() error { time.Sleep(100 * time.Millisecond); return nil }},
				{ID: "exec2", Function: func() error { time.Sleep(110 * time.Millisecond); return nil }},
			}

			start := time.Now()
			ff := false
			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Serial), engine.WithFailFast(&ff))
			duration := time.Since(start)

			Expect(summary.Results).To(HaveLen(2))
			Expect(summary.Results[0].Error).NotTo(HaveOccurred())
			Expect(summary.Results[1].Error).NotTo(HaveOccurred())
			Expect(duration).To(BeNumerically(">=", 200*time.Millisecond))
		})

		It("should handle exec failures with fail fast", func() {
			execs := []engine.Exec{
				{ID: "exec1", Function: func() error { return errors.New("error") }},
				{ID: "exec2", Function: func() error { return nil }},
			}

			ff := true
			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Serial), engine.WithFailFast(&ff))

			Expect(summary.Results).To(HaveLen(1))
			Expect(summary.Results[0].Error).To(HaveOccurred())
			Expect(summary.HasErrors()).To(BeTrue())
		})

		It("should skip exec when condition returns false", func() {
			executed := []string{}
			execs := []engine.Exec{
				{
					ID:       "exec1",
					Function: func() error { executed = append(executed, "exec1"); return nil },
				},
				{
					ID:        "exec2",
					Function:  func() error { executed = append(executed, "exec2"); return nil },
					Condition: func() (bool, error) { return false, nil },
				},
				{
					ID:       "exec3",
					Function: func() error { executed = append(executed, "exec3"); return nil },
				},
			}

			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Serial))

			Expect(summary.Results).To(HaveLen(2)) // Only exec1 and exec3
			Expect(executed).To(Equal([]string{"exec1", "exec3"}))
			Expect(summary.HasErrors()).To(BeFalse())
		})

		It("should execute when condition returns true", func() {
			executed := []string{}
			execs := []engine.Exec{
				{
					ID:        "exec1",
					Function:  func() error { executed = append(executed, "exec1"); return nil },
					Condition: func() (bool, error) { return true, nil },
				},
				{
					ID:        "exec2",
					Function:  func() error { executed = append(executed, "exec2"); return nil },
					Condition: func() (bool, error) { return true, nil },
				},
			}

			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Serial))

			Expect(summary.Results).To(HaveLen(2))
			Expect(executed).To(Equal([]string{"exec1", "exec2"}))
			Expect(summary.HasErrors()).To(BeFalse())
		})

		It("should evaluate conditions between executions", func() {
			sharedState := ""
			execs := []engine.Exec{
				{
					ID:       "exec1",
					Function: func() error { sharedState = "updated"; return nil },
				},
				{
					ID:       "exec2",
					Function: func() error { return nil },
					Condition: func() (bool, error) {
						// Condition can see update from exec1
						return sharedState == "updated", nil
					},
				},
			}

			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Serial))

			Expect(summary.Results).To(HaveLen(2))
			Expect(summary.HasErrors()).To(BeFalse())
		})

		It("should fail when condition evaluation fails", func() {
			execs := []engine.Exec{
				{
					ID:       "exec1",
					Function: func() error { return nil },
				},
				{
					ID:        "exec2",
					Function:  func() error { return nil },
					Condition: func() (bool, error) { return false, errors.New("condition error") },
				},
			}

			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Serial))

			Expect(summary.Results).To(HaveLen(2))
			Expect(summary.Results[0].Error).NotTo(HaveOccurred())
			Expect(summary.Results[1].Error).To(HaveOccurred())
			Expect(summary.Results[1].Error.Error()).To(ContainSubstring("condition evaluation failed"))
			Expect(summary.HasErrors()).To(BeTrue())
		})

		It("should stop on condition error with fail fast", func() {
			execs := []engine.Exec{
				{
					ID:        "exec1",
					Function:  func() error { return nil },
					Condition: func() (bool, error) { return false, errors.New("condition error") },
				},
				{
					ID:       "exec2",
					Function: func() error { return nil },
				},
			}

			ff := true
			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Serial), engine.WithFailFast(&ff))

			Expect(summary.Results).To(HaveLen(1))
			Expect(summary.Results[0].Error).To(HaveOccurred())
			Expect(summary.HasErrors()).To(BeTrue())
		})

		It("should continue on condition error without fail fast", func() {
			execs := []engine.Exec{
				{
					ID:        "exec1",
					Function:  func() error { return nil },
					Condition: func() (bool, error) { return false, errors.New("condition error") },
				},
				{
					ID:       "exec2",
					Function: func() error { return nil },
				},
			}

			ff := false
			summary := eng.Execute(ctx, execs, engine.WithMode(engine.Serial), engine.WithFailFast(&ff))

			Expect(summary.Results).To(HaveLen(2))
			Expect(summary.Results[0].Error).To(HaveOccurred())
			Expect(summary.Results[1].Error).NotTo(HaveOccurred())
			Expect(summary.HasErrors()).To(BeTrue())
		})
	})
})

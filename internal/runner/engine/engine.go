package engine

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/flowexec/flow/internal/runner/engine/retry"
)

//go:generate mockgen -destination=mocks/mock_engine.go -package=mocks github.com/flowexec/flow/internal/runner/engine Engine

type Result struct {
	ID      string
	Error   error
	Retries int
}

type ResultSummary struct {
	Results []Result
}

func (rs ResultSummary) HasErrors() bool {
	for _, r := range rs.Results {
		if r.Error != nil {
			return true
		}
	}
	return false
}

func (rs ResultSummary) String() string {
	var res string
	if rs.HasErrors() {
		res += "execution error encountered\n"
	}
	for _, r := range rs.Results {
		if r.Error == nil {
			continue
		}
		res += fmt.Sprintf("\n- Executable: %s\n  Error: %v", r.ID, r.Error)
		if r.Retries > 0 {
			res += fmt.Sprintf("\n  Retries: %d\n", r.Retries)
		}
	}
	return res
}

type Exec struct {
	ID         string
	Function   func() error
	Condition  func() (bool, error) // Optional condition evaluated before execution
	MaxRetries int
}

type ExecutionMode int

const (
	Parallel ExecutionMode = iota
	Serial
)

type Options struct {
	MaxThreads    int
	ExecutionMode ExecutionMode
	FailFast      *bool
}

type OptionFunc func(*Options)

type Engine interface {
	Execute(ctx context.Context, execs []Exec, opts ...OptionFunc) ResultSummary
}

type execEngine struct{}

func NewExecEngine() Engine {
	return &execEngine{}
}

func WithMaxThreads(maxThreads int) OptionFunc {
	return func(o *Options) {
		o.MaxThreads = maxThreads
	}
}

func WithFailFast(failFast *bool) OptionFunc {
	return func(o *Options) {
		o.FailFast = failFast
	}
}

func WithMode(mode ExecutionMode) OptionFunc {
	return func(o *Options) {
		o.ExecutionMode = mode
	}
}

func (e *execEngine) Execute(ctx context.Context, execs []Exec, opts ...OptionFunc) ResultSummary {
	options := Options{MaxThreads: 0, ExecutionMode: Serial}
	for _, opt := range opts {
		opt(&options)
	}
	var results []Result
	switch options.ExecutionMode {
	case Parallel:
		results = e.executeParallel(ctx, execs, options)
	case Serial:
		results = e.executeSerial(ctx, execs, options)
	default:
		results = []Result{{Error: fmt.Errorf("invalid execution mode")}}
	}
	return ResultSummary{Results: results}
}

func (e *execEngine) executeParallel(ctx context.Context, execs []Exec, opts Options) []Result {
	results := make([]Result, len(execs))

	groupCtx, groupCancel := context.WithCancel(ctx)
	defer groupCancel()
	group, _ := errgroup.WithContext(groupCtx)
	limit := opts.MaxThreads
	if limit == 0 {
		limit = len(execs)
	}
	group.SetLimit(limit)

	for i, exec := range execs {
		runExec := func() error {
			// Evaluate condition before execution
			if exec.Condition != nil {
				shouldRun, err := exec.Condition()
				if err != nil {
					results[i] = Result{
						ID:    exec.ID,
						Error: fmt.Errorf("condition evaluation failed: %w", err),
					}
					ff := opts.FailFast == nil || *opts.FailFast
					if ff {
						return err
					}
					return nil
				}
				if !shouldRun {
					// Skip this execution - leave result empty/zero value
					return nil
				}
			}

			rh := retry.NewRetryHandler(exec.MaxRetries, 0)
			err := rh.Execute(exec.Function)
			results[i] = Result{
				ID:      exec.ID,
				Error:   err,
				Retries: rh.GetStats().Attempts - 1,
			}
			ff := opts.FailFast == nil || *opts.FailFast
			if err != nil && ff {
				return err
			}
			return nil
		}
		group.Go(runExec)
	}

	if err := group.Wait(); err != nil {
		if len(results) > 0 {
			return results
		}
		return []Result{{Error: err}}
	}
	return results
}

func (e *execEngine) executeSerial(ctx context.Context, execs []Exec, opts Options) []Result {
	results := make([]Result, 0, len(execs))
	for _, exec := range execs {
		select {
		case <-ctx.Done():
			results = append(results, Result{
				ID:    exec.ID,
				Error: ctx.Err(),
			})
			return results
		default:
			// Evaluate condition before execution
			if exec.Condition != nil {
				shouldRun, err := exec.Condition()
				if err != nil {
					results = append(results, Result{
						ID:    exec.ID,
						Error: fmt.Errorf("condition evaluation failed: %w", err),
					})
					ff := opts.FailFast == nil || *opts.FailFast
					if ff {
						return results
					}
					continue
				}
				if !shouldRun {
					// Skip this execution - don't add to results
					continue
				}
			}

			rh := retry.NewRetryHandler(exec.MaxRetries, 0)
			err := rh.Execute(exec.Function)
			results = append(results, Result{
				ID:      exec.ID,
				Error:   err,
				Retries: rh.GetStats().Attempts - 1,
			})

			ff := opts.FailFast == nil || *opts.FailFast
			if err != nil && ff {
				return results
			}
		}
	}

	return results
}

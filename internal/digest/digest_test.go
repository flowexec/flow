package digest_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tuikitIO "github.com/flowexec/tuikit/io"

	"github.com/flowexec/flow/v2/internal/digest"
	flowerrors "github.com/flowexec/flow/v2/pkg/errors"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/types/executable"
)

var now = time.Date(2026, 11, 21, 10, 30, 0, 0, time.UTC)

type fixture struct {
	t       *testing.T
	ds      store.DataStore
	logsDir string
	wsDir   string
	execs   *fakeCache
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	dir := t.TempDir()
	ds, err := store.NewDataStore(filepath.Join(dir, "store.db"))
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { _ = ds.Close() })
	wsDir := filepath.Join(dir, "ws")
	if err := os.MkdirAll(wsDir, 0o750); err != nil {
		t.Fatal(err)
	}
	return &fixture{
		t: t, ds: ds, logsDir: filepath.Join(dir, "logs"), wsDir: wsDir,
		execs: &fakeCache{byRef: map[executable.Ref]*executable.Executable{}},
	}
}

func (f *fixture) deps() digest.Deps {
	return digest.Deps{
		Store: f.ds, LogsDir: f.logsDir, Executables: f.execs, WorkspaceRoot: f.wsDir,
		CurrentWorkspace: "ws", Now: func() time.Time { return now },
	}
}

// addRun records a run and, when log is non-empty, writes its archive file.
func (f *fixture) addRun(id, ref string, exit int, ago time.Duration, log string) {
	f.t.Helper()
	rec := store.ExecutionRecord{
		ID: id, Ref: ref, ExitCode: exit, StartedAt: now.Add(-ago), Duration: 1500 * time.Millisecond,
		Status: store.RunCompleted,
	}
	if exit != 0 {
		rec.Status = store.RunFailed
		rec.Error = fmt.Sprintf("exit status %d", exit)
	}
	if log != "" {
		file := tuikitIO.NewArchiveLogFile(f.logsDir, id)
		if _, err := file.WriteString(log); err != nil {
			f.t.Fatal(err)
		}
		_ = file.Close()
		rec.LogArchiveID = file.Name()
	}
	if err := f.ds.RecordExecution(rec); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) writeFile(rel, content string) string {
	f.t.Helper()
	path := filepath.Join(f.wsDir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		f.t.Fatal(err)
	}
	return path
}

type fakeCache struct {
	byRef   map[executable.Ref]*executable.Executable
	updates int
}

func (c *fakeCache) Update() error { c.updates++; return nil }
func (c *fakeCache) GetExecutableByRef(ref executable.Ref) (*executable.Executable, error) {
	if e, ok := c.byRef[ref]; ok {
		return e, nil
	}
	return nil, flowerrors.NewExecutableNotFoundError(ref.String())
}
func (c *fakeCache) GetExecutableList() (executable.ExecutableList, error) { return nil, nil }
func (c *fakeCache) AliasRefs() (map[executable.Ref]executable.Ref, error) {
	return map[executable.Ref]executable.Ref{}, nil
}

func lines(n int, prefix string) string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("%s line %d", prefix, i+1)
	}
	return strings.Join(out, "\n") + "\n"
}

func resolve(t *testing.T, f *fixture, opts digest.Options, sources ...string) *digest.Document {
	t.Helper()
	doc, err := digest.Resolve(context.Background(), sources, opts, f.deps())
	if err != nil {
		t.Fatalf("Resolve(%v): %v", sources, err)
	}
	return doc
}

func TestRunSection(t *testing.T) {
	f := newFixture(t)
	f.addRun("run-1", "test ws/unit", 1, 21*time.Minute, "setup\n--- FAIL: TestMerge\nFAIL\n")

	doc := resolve(t, f, digest.Options{}, "run:run-1")
	s := doc.Sections[0]
	if s.Type != digest.TypeRun || s.Label != "Run run-1" {
		t.Fatalf("unexpected section header: %+v", s)
	}
	if s.Metadata["status"] != "failed" || s.Metadata["exitCode"] != 1 || s.Metadata["execRef"] != "test ws/unit" {
		t.Fatalf("unexpected metadata: %v", s.Metadata)
	}
	if s.Content != "setup\n--- FAIL: TestMerge\nFAIL" || s.Truncated {
		t.Fatalf("unexpected content %q (truncated=%v)", s.Content, s.Truncated)
	}
	text := doc.Text()
	for _, want := range []string{
		"## Run run-1", "- Status: failed (exit 1)", "### Output\n\n```text\nsetup\n--- FAIL: TestMerge\nFAIL\n```",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("text output missing %q:\n%s", want, text)
		}
	}
}

func TestRunSection_MaxLinesKeepsTail(t *testing.T) {
	f := newFixture(t)
	f.addRun("run-1", "test ws/unit", 1, time.Minute, lines(50, "log"))

	doc := resolve(t, f, digest.Options{MaxLines: 3}, "run:run-1")
	s := doc.Sections[0]
	if s.Content != "log line 48\nlog line 49\nlog line 50" || !s.Truncated {
		t.Fatalf("expected the last 3 lines, truncated; got %q (truncated=%v)", s.Content, s.Truncated)
	}
	if !strings.Contains(doc.Text(), "### Output (last 3 of 50 lines)") {
		t.Fatalf("text heading should name the line cap:\n%s", doc.Text())
	}
}

func TestRunSection_NotFound(t *testing.T) {
	f := newFixture(t)
	_, err := digest.Resolve(context.Background(), []string{"run:missing"}, digest.Options{}, f.deps())
	var nf flowerrors.RunNotFoundError
	if !errors.As(err, &nf) || nf.Code() != flowerrors.ErrCodeNotFound {
		t.Fatalf("expected RunNotFoundError, got %v", err)
	}
}

func TestRunsSection_FiltersAndNotesRepeatedFailures(t *testing.T) {
	f := newFixture(t)
	f.addRun("a", "test ws/unit", 1, 1*time.Hour, "")
	f.addRun("b", "test ws/unit", 1, 2*time.Hour, "")
	f.addRun("c", "test ws/unit", 0, 3*time.Hour, "")
	f.addRun("d", "build ws/binary", 1, 4*time.Hour, "")
	f.addRun("e", "test ws/unit", 1, 10*24*time.Hour, "")

	doc := resolve(t, f, digest.Options{}, "runs:unit status:failed ran:7d")
	s := doc.Sections[0]
	rows, ok := s.Metadata["runs"].([]digest.RunSummary)
	if !ok || len(rows) != 2 || rows[0].ID != "a" || rows[1].ID != "b" {
		t.Fatalf("expected failed unit runs a, b within 7d; got %+v", s.Metadata["runs"])
	}
	if !strings.Contains(s.Content, "| a | test ws/unit | failed | 1 |") {
		t.Fatalf("table missing row for a:\n%s", s.Content)
	}
	// c (passed) is outside status:failed but still counts toward the ref's recent history, and e
	// is outside the 7d window.
	if !strings.Contains(s.Content, "### Repeated failures\n\n- `test ws/unit` — failed 2 of its last 3 runs within 7d, "+
		"including the latest") {
		t.Fatalf("expected a repeated-failure note:\n%s", s.Content)
	}
	notes, ok := s.Metadata["repeatedFailures"].([]digest.RepeatedFailure)
	if !ok || len(notes) != 1 || notes[0].Failed != 2 || notes[0].Runs != 3 {
		t.Fatalf("unexpected structured notes: %+v", s.Metadata["repeatedFailures"])
	}
	if strings.Contains(s.Content, "line") {
		t.Fatalf("runs summary must not include log content:\n%s", s.Content)
	}
}

func TestRunsSection_Limit(t *testing.T) {
	f := newFixture(t)
	for i := range 5 {
		f.addRun(fmt.Sprintf("r%d", i), "test ws/unit", 0, time.Duration(i+1)*time.Minute, "")
	}
	doc := resolve(t, f, digest.Options{}, "runs:limit:2")
	s := doc.Sections[0]
	if s.Metadata["count"] != 2 || s.Metadata["matched"] != 5 {
		t.Fatalf("expected 2 of 5 runs, got %v of %v", s.Metadata["count"], s.Metadata["matched"])
	}
	if !strings.Contains(doc.Text(), "- Matched: 5 (showing the newest 2)") {
		t.Fatalf("text should say the table is cut:\n%s", doc.Text())
	}
}

func TestRunsSection_NoNoteForRecoveredRefs(t *testing.T) {
	f := newFixture(t)
	f.addRun("ok", "test ws/unit", 0, 1*time.Hour, "")
	f.addRun("bad1", "test ws/unit", 1, 2*time.Hour, "")
	f.addRun("bad2", "test ws/unit", 1, 3*time.Hour, "")

	doc := resolve(t, f, digest.Options{}, "runs:status:failed")
	if _, ok := doc.Sections[0].Metadata["repeatedFailures"]; ok {
		t.Fatalf("a ref whose latest run passed shouldn't be flagged: %v", doc.Sections[0].Metadata)
	}
}

func TestRunsSection_RejectsUnknownFilterKeys(t *testing.T) {
	f := newFixture(t)
	_, err := digest.Resolve(context.Background(), []string{"runs:stauts:failed"}, digest.Options{}, f.deps())
	if err == nil || !strings.Contains(err.Error(), `unknown filter "stauts"`) {
		t.Fatalf("expected an unknown-filter error, got %v", err)
	}
	// A full ref's colon is not a key.
	resolve(t, f, digest.Options{}, "runs:ws/ns:unit")
}

func TestRunSection_IDPrefix(t *testing.T) {
	f := newFixture(t)
	f.addRun("abcdef12-0000", "test ws/unit", 0, time.Minute, "")
	f.addRun("abcdef99-0000", "test ws/unit", 0, time.Minute, "")

	doc := resolve(t, f, digest.Options{}, "run:abcdef12")
	if doc.Sections[0].Metadata["id"] != "abcdef12-0000" {
		t.Fatalf("prefix should resolve to the full ID: %v", doc.Sections[0].Metadata)
	}
	_, err := digest.Resolve(context.Background(), []string{"run:abcdef"}, digest.Options{}, f.deps())
	if err == nil || !strings.Contains(err.Error(), "matches 2 runs") {
		t.Fatalf("expected an ambiguous-prefix error, got %v", err)
	}
	_, err = digest.Resolve(context.Background(), []string{"run:abc"}, digest.Options{}, f.deps())
	var nf flowerrors.RunNotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("a prefix under 6 characters should not match: %v", err)
	}
}

func TestRunsSection_ErrorSnippetKeepsTheCause(t *testing.T) {
	f := newFixture(t)
	rec := store.ExecutionRecord{
		ID: "nested", Ref: "validate ws/", StartedAt: now, Status: store.RunFailed, ExitCode: 1,
		Error: "serial execution failed: lint ws/: parallel execution failed: lint ws/go: " +
			"parallel execution failed: lint ws/go: command exited with non-zero status exit status 1",
	}
	if err := f.ds.RecordExecution(rec); err != nil {
		t.Fatal(err)
	}
	doc := resolve(t, f, digest.Options{}, "runs:")
	if !strings.Contains(doc.Sections[0].Content, "| lint ws/: lint ws/go: exit status 1 |") {
		t.Fatalf("error cell should drop wrappers and repeats:\n%s", doc.Sections[0].Content)
	}
}

func TestText_FencesEmbeddedMarkdown(t *testing.T) {
	f := newFixture(t)
	f.writeFile("README.md", "## Not a digest section\n```go\nx := 1\n```")
	text := resolve(t, f, digest.Options{}, "file:README.md").Text()
	if !strings.Contains(text, "````md\n## Not a digest section") {
		t.Fatalf("file content should be fenced with a fence longer than any inside it:\n%s", text)
	}
}

func TestExecSection(t *testing.T) {
	f := newFixture(t)
	flowFile := f.writeFile(".execs/test.flow", "executables: []\n")
	f.writeFile(".execs/scripts/unit.sh", "#!/bin/sh\ngo test ./...\n")
	e := &executable.Executable{
		Verb: "test", Name: "unit", Tags: []string{"go", "unit"},
		Exec: &executable.ExecExecutableType{File: "scripts/unit.sh"},
	}
	e.SetContext("ws", f.wsDir, "", flowFile)
	ref := executable.NewRef("ws/unit", "test")
	f.execs.byRef[ref] = e
	f.addRun("r1", ref.String(), 1, 21*time.Minute, "")
	f.addRun("r2", ref.String(), 0, 6*24*time.Hour, "")

	for _, src := range []string{"exec:test/unit", "exec:test unit", "exec:test ws/unit"} {
		doc := resolve(t, f, digest.Options{}, src)
		s := doc.Sections[0]
		if s.Label != "Executable: test ws/unit" {
			t.Fatalf("%s: unexpected label %q", src, s.Label)
		}
		for _, want := range []string{
			"### Definition\n\n```yaml\nexec:\n  file: scripts/unit.sh\nname: unit\n",
			"### Script: .execs/scripts/unit.sh\n\n```sh\n#!/bin/sh\ngo test ./...\n```",
			"### Recent runs", "| r1 | failed | 1 | 1.5s | 21 min ago |", "| r2 | completed | 0 | 1.5s | 6d ago |",
		} {
			if !strings.Contains(s.Content, want) {
				t.Errorf("%s: content missing %q:\n%s", src, want, s.Content)
			}
		}
	}

	doc := resolve(t, f, digest.Options{NoHistory: true}, "exec:test/unit")
	if strings.Contains(doc.Sections[0].Content, "Recent runs") {
		t.Fatalf("--no-history should omit recent runs")
	}
}

func TestExecSection_SyncsCacheOnMiss(t *testing.T) {
	f := newFixture(t)
	_, err := digest.Resolve(context.Background(), []string{"exec:build/missing"}, digest.Options{}, f.deps())
	var nf flowerrors.ExecutableNotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("expected ExecutableNotFoundError, got %v", err)
	}
	if f.execs.updates != 1 {
		t.Fatalf("expected one cache sync on a miss, got %d", f.execs.updates)
	}
}

func TestFileSection(t *testing.T) {
	f := newFixture(t)
	f.writeFile("config.yaml", "a: 1\nb: 2\n")

	doc := resolve(t, f, digest.Options{}, "file:config.yaml")
	s := doc.Sections[0]
	if s.Content != "a: 1\nb: 2" || s.Truncated || s.Metadata["path"] != "config.yaml" {
		t.Fatalf("unexpected file section: %+v", s)
	}
}

func TestFileSection_RejectsPathsOutsideWorkspace(t *testing.T) {
	f := newFixture(t)
	outside := filepath.Join(filepath.Dir(f.wsDir), "secret.txt")
	if err := os.WriteFile(outside, []byte("nope"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(f.wsDir, "link.txt")); err != nil {
		t.Fatal(err)
	}

	for _, src := range []string{"file:../secret.txt", "file:" + outside, "file:link.txt"} {
		_, err := digest.Resolve(context.Background(), []string{src}, digest.Options{}, f.deps())
		var ue flowerrors.UsageError
		if !errors.As(err, &ue) || !strings.Contains(err.Error(), "outside the workspace") {
			t.Errorf("%s: expected an outside-workspace usage error, got %v", src, err)
		}
	}

	_, err := digest.Resolve(context.Background(), []string{"file:missing.txt"}, digest.Options{}, f.deps())
	var nf flowerrors.FileNotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("expected FileNotFoundError, got %v", err)
	}
}

func TestUnknownAndMalformedSources(t *testing.T) {
	f := newFixture(t)
	_, err := digest.Resolve(context.Background(),
		[]string{"bogus:x", "noprefix", "runs:status:maybe", "file:ok"}, digest.Options{}, f.deps())
	var ue flowerrors.UsageError
	if !errors.As(err, &ue) || ue.Code() != flowerrors.ErrCodeInvalidInput {
		t.Fatalf("expected a usage error, got %v", err)
	}
	for _, want := range []string{
		`"bogus:x": unknown type "bogus"`, `"noprefix": expected type:value`, `invalid status "maybe"`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %q: %v", want, err)
		}
	}
	if strings.Contains(err.Error(), "file:ok") {
		t.Errorf("valid sources should not be reported: %v", err)
	}
}

func TestBudget_ComposesSectionsAndCapsLogsFirst(t *testing.T) {
	f := newFixture(t)
	f.addRun("run-1", "test ws/unit", 1, time.Minute, lines(300, "log"))
	f.writeFile("small.txt", "small file")
	f.writeFile("medium.txt", strings.Repeat("0123456789\n", 200))

	const budget = 3000
	doc := resolve(t, f, digest.Options{MaxChars: budget}, "file:small.txt", "run:run-1", "file:medium.txt")
	if len(doc.Sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(doc.Sections))
	}
	small, run, medium := doc.Sections[0], doc.Sections[1], doc.Sections[2]
	if small.Truncated || medium.Truncated {
		t.Fatalf("files that fit should be kept whole while logs shrink: small=%v medium=%v",
			small.Truncated, medium.Truncated)
	}
	if !run.Truncated || !strings.HasSuffix(run.Content, "log line 300") ||
		!strings.HasPrefix(run.Content, "log line ") {
		t.Fatalf("the log should keep whole lines from its tail: %q", run.Content)
	}
	if doc.TotalChars > budget || doc.MaxChars != budget {
		t.Fatalf("total %d exceeds budget %d", doc.TotalChars, budget)
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"sections"`, `"type":"run"`, `"truncated":true`, `"totalChars"`, `"maxChars":3000`} {
		if !strings.Contains(string(data), key) {
			t.Errorf("JSON missing %s", key)
		}
	}
}

func TestBudget_TruncatesFilesFromTheEnd(t *testing.T) {
	f := newFixture(t)
	f.writeFile("big.txt", strings.Repeat("0123456789\n", 200))

	doc := resolve(t, f, digest.Options{MaxChars: 100}, "file:big.txt")
	s := doc.Sections[0]
	if !s.Truncated || !strings.HasPrefix(s.Content, "0123456789\n") || strings.HasSuffix(s.Content, "\n") {
		t.Fatalf("expected whole lines from the head: %q", s.Content)
	}
	if doc.TotalChars > 100 {
		t.Fatalf("total %d exceeds budget", doc.TotalChars)
	}
}

func TestBudget_MetadataSurvivesAZeroContentBudget(t *testing.T) {
	f := newFixture(t)
	f.addRun("run-1", "test ws/unit", 1, time.Minute, lines(10, "log"))
	doc := resolve(t, f, digest.Options{MaxChars: 1}, "run:run-1")
	s := doc.Sections[0]
	if s.Metadata["status"] != "failed" || !s.Truncated {
		t.Fatalf("metadata should remain and content be marked truncated: %+v", s)
	}
}

func TestBudget_CompositeSectionsStayWithinBudget(t *testing.T) {
	f := newFixture(t)
	flowFile := f.writeFile(".execs/test.flow", "executables: []\n")
	f.writeFile(".execs/big.sh", strings.Repeat("echo line\n", 500))
	e := &executable.Executable{Verb: "test", Name: "big", Exec: &executable.ExecExecutableType{File: "big.sh"}}
	e.SetContext("ws", f.wsDir, "", flowFile)
	f.execs.byRef[executable.NewRef("ws/big", "test")] = e

	const budget = 800
	doc := resolve(t, f, digest.Options{MaxChars: budget}, "exec:test/big")
	if doc.TotalChars > budget {
		t.Fatalf("total %d exceeds budget %d:\n%s", doc.TotalChars, budget, doc.Sections[0].Content)
	}
	if !strings.Contains(doc.Sections[0].Content, "### Script: .execs/big.sh (first ") {
		t.Fatalf("the cut script heading should say how much was kept:\n%s", doc.Sections[0].Content)
	}
}

func TestRunsSection_RunningRunHasNoDurationOrExit(t *testing.T) {
	f := newFixture(t)
	rec := store.ExecutionRecord{ID: "live", Ref: "test ws/unit", StartedAt: now, Status: store.RunRunning}
	if err := f.ds.RecordExecution(rec); err != nil {
		t.Fatal(err)
	}
	doc := resolve(t, f, digest.Options{}, "runs:")
	if !strings.Contains(doc.Sections[0].Content, "| live | test ws/unit | running | - | - |") {
		t.Fatalf("a running run shouldn't show a final duration or exit code:\n%s", doc.Sections[0].Content)
	}
}

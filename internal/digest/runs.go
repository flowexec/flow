package digest

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flowexec/flow/v2/internal/io/logs"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/types/executable"
)

// RunSummary is one row of a run table.
type RunSummary struct {
	ID        string `json:"id"                 yaml:"id"`
	Ref       string `json:"execRef"            yaml:"execRef"`
	Status    string `json:"status"             yaml:"status"`
	ExitCode  int    `json:"exitCode"           yaml:"exitCode"`
	Duration  string `json:"duration,omitempty" yaml:"duration,omitempty"`
	StartedAt string `json:"startedAt"          yaml:"startedAt"`
	Error     string `json:"error,omitempty"    yaml:"error,omitempty"`
}

type runsQuery struct {
	ref    string
	status string
	since  time.Duration
	// sinceText is the window as written (7d), for notes.
	sinceText string
	limit     int
}

// parseRunsQuery reads `key:value` terms (ref, status, since/ran, limit). Bare words are part of
// the ref match, so `runs:"test unit status:failed"` matches failed runs of refs containing
// "test unit".
func parseRunsQuery(value string) (runsQuery, error) {
	q := runsQuery{limit: defaultRunsLimit}
	var refTerms []string
	for _, tok := range strings.Fields(value) {
		key, val, ok := strings.Cut(tok, ":")
		if !ok {
			refTerms = append(refTerms, tok)
			continue
		}
		switch strings.ToLower(key) {
		case "ref":
			refTerms = append(refTerms, val)
		case "status":
			status, err := normalizeStatus(val)
			if err != nil {
				return q, err
			}
			q.status = status
		case "since", "ran":
			d, err := logs.ParseAge(val)
			if err != nil || d <= 0 {
				return q, fmt.Errorf("invalid %s %q; use a duration like 90m, 24h or 7d", key, val)
			}
			q.since, q.sinceText = d, val
		case "limit":
			n, err := strconv.Atoi(val)
			if err != nil || n <= 0 {
				return q, fmt.Errorf("invalid limit %q; use a positive number", val)
			}
			q.limit = n
		default:
			// A colon inside a full ref (ws/ns:name) is not a key; anything else is a typo, which
			// would otherwise silently match nothing.
			if !strings.Contains(key, "/") {
				return q, fmt.Errorf("unknown filter %q (filters: ref, status, ran or since, limit; "+
					"write ref:<ns>:<name> to match a namespaced ref)", key)
			}
			refTerms = append(refTerms, tok)
		}
	}
	q.ref = strings.Join(refTerms, " ")
	return q, nil
}

func normalizeStatus(s string) (string, error) {
	switch strings.ToLower(s) {
	case "passed", "pass", "success", "ok", "completed":
		return string(store.RunCompleted), nil
	case "failed", "fail", "failure", "error":
		return string(store.RunFailed), nil
	case "cancelled", "canceled":
		return string(store.RunCancelled), nil
	case "running":
		return string(store.RunRunning), nil
	}
	return "", fmt.Errorf("invalid status %q; use passed, failed, cancelled or running", s)
}

func (r *resolver) runSection(id string) (*Section, error) {
	rec, err := logs.FindRecord(r.deps.Store, r.deps.LogsDir, id)
	if err != nil {
		return nil, err
	}

	s := &Section{Type: TypeRun, Ref: id, Label: "Run " + id}
	status := string(logs.CanonicalStatus(*rec))
	s.Metadata = map[string]any{
		"id":        rec.ID,
		"execRef":   rec.Ref,
		"status":    status,
		"exitCode":  rec.ExitCode,
		"startedAt": rec.StartedAt.UTC().Format(time.RFC3339),
	}
	setIf(s.Metadata, "duration", runDuration(*rec))
	setIf(s.Metadata, "completedAt", formatTime(rec.CompletedAt))
	setIf(s.Metadata, "workspace", refWorkspace(rec.Ref))
	setIf(s.Metadata, "command", rec.Command)
	setIf(s.Metadata, "label", rec.Label)
	setIf(s.Metadata, "error", rec.Error)
	setIf(s.Metadata, "workingDir", rec.WorkingDir)

	s.addField("Ref", rec.Ref)
	s.addField("Label", rec.Label)
	s.addField("Status", statusText(status, rec.ExitCode))
	s.addField("Duration", runDuration(*rec))
	s.addField("Started", rec.StartedAt.UTC().Format(time.RFC3339))
	s.addField("Workspace", refWorkspace(rec.Ref))
	s.addField("Command", rec.Command)
	s.addField("Error", rec.Error)

	if rec.LogEntry == nil {
		s.Metadata["log"] = "unavailable"
		return s, nil
	}
	raw, err := rec.LogEntry.Read()
	if err != nil {
		// A pruned or unreadable log still leaves the run's metadata worth returning.
		s.Metadata["log"] = "unreadable"
		return s, nil //nolint:nilerr // see above
	}
	res, err := logs.ExtractContent(logs.HideStreamField(raw), logs.ContentOptions{Tail: r.opts.MaxLines})
	if err != nil {
		return nil, err
	}
	s.Metadata["logLines"] = res.TotalLines
	s.parts = append(s.parts, &part{
		heading: "Output", body: res.Content, tier: tierLog, keepTail: true, limit: maxLogChars,
		verbatim: true, lang: "text", totalLines: res.TotalLines, truncated: res.Truncated,
	})
	return s, nil
}

func (r *resolver) runsSection(value string, q runsQuery) (*Section, error) {
	records, err := logs.LoadRecords(r.deps.Store, r.deps.LogsDir)
	if err != nil {
		return nil, fmt.Errorf("reading run history: %w", err)
	}
	var window logs.RecordFilter
	if q.since > 0 {
		window.Since = r.deps.Now().Add(-q.since)
	}
	records = logs.FilterRecords(records, window)
	if q.ref != "" {
		want := strings.ToLower(q.ref)
		var matched []logs.UnifiedRecord
		for _, rec := range records {
			if strings.Contains(strings.ToLower(rec.Ref), want) {
				matched = append(matched, rec)
			}
		}
		records = matched
	}
	// Failure streaks are judged on each ref's recent history, before status and limit narrow
	// what the table shows; otherwise status:failed would report every ref as always failing.
	failing := repeatedFailures(records)

	records = logs.FilterRecords(records, logs.RecordFilter{Status: q.status})
	matched := len(records)
	if len(records) > q.limit {
		records = records[:q.limit]
	}

	label := "Runs"
	if value != "" {
		label = "Runs: " + value
	}
	s := &Section{Type: TypeRuns, Ref: value, Label: label}
	rows := make([]RunSummary, 0, len(records))
	shown := map[string]bool{}
	for _, rec := range records {
		rows = append(rows, summarize(rec))
		shown[rec.Ref] = true
	}
	var notes []RepeatedFailure
	for _, f := range failing {
		if shown[f.Ref] {
			notes = append(notes, f)
		}
	}

	s.Metadata = map[string]any{"matched": matched, "count": len(rows), "runs": rows}
	if len(notes) > 0 {
		s.Metadata["repeatedFailures"] = notes
	}
	if matched > len(rows) {
		s.addField("Matched", fmt.Sprintf("%d (showing the newest %d)", matched, len(rows)))
	} else {
		s.addField("Matched", strconv.Itoa(matched))
	}

	body := "No matching runs."
	if len(rows) > 0 {
		body = r.runTable(rows, true)
	}
	s.parts = append(s.parts, &part{body: body, tier: tierFixed})
	if len(notes) > 0 {
		s.parts = append(s.parts, &part{
			heading: "Repeated failures", body: failureNotes(notes, q.sinceText), tier: tierFixed,
		})
	}
	return s, nil
}

// RepeatedFailure flags a ref whose latest run failed and that failed more than once recently.
type RepeatedFailure struct {
	Ref    string `json:"execRef" yaml:"execRef"`
	Failed int    `json:"failed"  yaml:"failed"`
	Runs   int    `json:"runs"    yaml:"runs"`
}

// repeatedFailures finds refs that are failing now (latest run failed) and failed at least twice
// in the given history, which must be newest first. Most failures first.
func repeatedFailures(records []logs.UnifiedRecord) []RepeatedFailure {
	byRef := map[string]*RepeatedFailure{}
	latestFailed := map[string]bool{}
	var order []string
	for _, rec := range records {
		failed := logs.CanonicalStatus(rec) == store.RunFailed
		t, ok := byRef[rec.Ref]
		if !ok {
			t = &RepeatedFailure{Ref: rec.Ref}
			byRef[rec.Ref] = t
			latestFailed[rec.Ref] = failed
			order = append(order, rec.Ref)
		}
		t.Runs++
		if failed {
			t.Failed++
		}
	}
	var out []RepeatedFailure
	for _, ref := range order {
		if t := byRef[ref]; latestFailed[ref] && t.Failed >= 2 {
			out = append(out, *t)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Failed > out[j].Failed })
	return out
}

func failureNotes(notes []RepeatedFailure, since string) string {
	window := ""
	if since != "" {
		window = " within " + since
	}
	lines := make([]string, 0, len(notes))
	for _, n := range notes {
		if n.Failed == n.Runs {
			lines = append(lines, fmt.Sprintf("- `%s` — failed all of its last %d runs%s", n.Ref, n.Runs, window))
		} else {
			lines = append(lines, fmt.Sprintf("- `%s` — failed %d of its last %d runs%s, including the latest",
				n.Ref, n.Failed, n.Runs, window))
		}
	}
	return strings.Join(lines, "\n")
}

func summarize(rec logs.UnifiedRecord) RunSummary {
	return RunSummary{
		ID:        rec.ID,
		Ref:       rec.Ref,
		Status:    string(logs.CanonicalStatus(rec)),
		ExitCode:  rec.ExitCode,
		Duration:  runDuration(rec),
		StartedAt: rec.StartedAt.UTC().Format(time.RFC3339),
		Error:     errorSnippet(rec.Error, 80),
	}
}

func (r *resolver) runTable(rows []RunSummary, withRef bool) string {
	header := []string{"ID", "Status", "Exit", "Duration", "When"}
	if withRef {
		header = []string{"ID", "Ref", "Status", "Exit", "Duration", "When", "Error"}
	}
	table := make([][]string, 0, len(rows))
	for _, row := range rows {
		started, _ := time.Parse(time.RFC3339, row.StartedAt)
		when := relativeTime(r.deps.Now(), started)
		exit := strconv.Itoa(row.ExitCode)
		if row.Status == string(store.RunRunning) {
			exit = "-"
		}
		id := shortID(row.ID)
		if withRef {
			table = append(table, []string{id, row.Ref, row.Status, exit, row.Duration, when, row.Error})
		} else {
			table = append(table, []string{id, row.Status, exit, row.Duration, when})
		}
	}
	return markdownTable(header, table)
}

func statusText(status string, exitCode int) string {
	if status == string(store.RunFailed) {
		return fmt.Sprintf("%s (exit %d)", status, exitCode)
	}
	return status
}

func refWorkspace(ref string) string {
	_, id, ok := strings.Cut(ref, " ")
	if !ok {
		return ""
	}
	ws, _, _, err := executable.ParseExecutableID(id)
	if err != nil || ws == executable.WildcardWorkspace {
		return ""
	}
	return ws
}

func setIf(m map[string]any, key, value string) {
	if value != "" {
		m[key] = value
	}
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// runDuration is empty while a run is in flight, since its recorded duration is not final.
func runDuration(rec logs.UnifiedRecord) string {
	if logs.CanonicalStatus(rec) == store.RunRunning {
		return ""
	}
	return formatDuration(rec.Duration)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(100 * time.Millisecond).String()
}

func relativeTime(now, t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// shortID abbreviates a run ID for tables; run:<prefix> resolves it back.
func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// errorWrappers are the prefixes composite runners add around a step's error. They carry no
// information the ref column doesn't, and would push the actual cause out of a short cell.
var errorWrappers = []string{
	"serial execution failed: ", "parallel execution failed: ", "command exited with non-zero status ",
}

// errorSnippet shortens a run error to its first line, keeping its end, where the cause is.
func errorSnippet(s string, n int) string {
	s, _, _ = strings.Cut(strings.TrimSpace(s), "\n")
	for _, w := range errorWrappers {
		s = strings.ReplaceAll(s, w, "")
	}
	// Nested runners repeat the failing step's ref once per level.
	segments := strings.Split(s, ": ")
	kept := segments[:0]
	for i, seg := range segments {
		if i == 0 || seg != segments[i-1] {
			kept = append(kept, seg)
		}
	}
	s = strings.Join(kept, ": ")
	if runeLen(s) <= n {
		return s
	}
	r := []rune(s)
	return "…" + string(r[len(r)-(n-1):])
}

func markdownTable(header []string, rows [][]string) string {
	esc := func(c string) string {
		if c == "" {
			return "-"
		}
		return strings.ReplaceAll(c, "|", `\|`)
	}
	var b strings.Builder
	b.WriteString("| " + strings.Join(header, " | ") + " |\n")
	seps := make([]string, len(header))
	for i := range seps {
		seps[i] = "---"
	}
	b.WriteString("| " + strings.Join(seps, " | ") + " |")
	for _, row := range rows {
		cells := make([]string, len(row))
		for i, c := range row {
			cells[i] = esc(c)
		}
		b.WriteString("\n| " + strings.Join(cells, " | ") + " |")
	}
	return b.String()
}

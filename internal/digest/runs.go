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
	ID        string `json:"id"              yaml:"id"`
	Ref       string `json:"execRef"         yaml:"execRef"`
	Status    string `json:"status"          yaml:"status"`
	ExitCode  int    `json:"exitCode"        yaml:"exitCode"`
	Duration  string `json:"duration"        yaml:"duration"`
	StartedAt string `json:"startedAt"       yaml:"startedAt"`
	Error     string `json:"error,omitempty" yaml:"error,omitempty"`
}

type runsQuery struct {
	ref    string
	status string
	since  time.Duration
	limit  int
}

// parseRunsQuery reads `key:value` terms (ref, status, since/ran, limit). Any other word is part
// of the ref match, so `runs:"test unit status:failed"` matches failed runs of refs containing
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
			q.since = d
		case "limit":
			n, err := strconv.Atoi(val)
			if err != nil || n <= 0 {
				return q, fmt.Errorf("invalid limit %q; use a positive number", val)
			}
			q.limit = n
		default:
			// A colon inside a ref (ws/ns:name) is not a key.
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
		"duration":  formatDuration(rec.Duration),
		"startedAt": rec.StartedAt.UTC().Format(time.RFC3339),
	}
	setIf(s.Metadata, "completedAt", formatTime(rec.CompletedAt))
	setIf(s.Metadata, "workspace", refWorkspace(rec.Ref))
	setIf(s.Metadata, "command", rec.Command)
	setIf(s.Metadata, "label", rec.Label)
	setIf(s.Metadata, "error", rec.Error)
	setIf(s.Metadata, "workingDir", rec.WorkingDir)

	s.addField("Ref", rec.Ref)
	s.addField("Label", rec.Label)
	s.addField("Status", statusText(status, rec.ExitCode))
	s.addField("Duration", formatDuration(rec.Duration))
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
	s.textHeading = "Output"
	if res.Truncated {
		s.textHeading = fmt.Sprintf("Output (last %d lines)", res.ReturnedLines)
	}
	s.parts = append(s.parts, &part{
		body: res.Content, tier: tierLog, keepTail: true, limit: maxLogChars, truncated: res.Truncated,
	})
	return s, nil
}

func (r *resolver) runsSection(value string, q runsQuery) (*Section, error) {
	records, err := logs.LoadRecords(r.deps.Store, r.deps.LogsDir)
	if err != nil {
		return nil, fmt.Errorf("reading run history: %w", err)
	}
	filter := logs.RecordFilter{Status: q.status}
	if q.since > 0 {
		filter.Since = r.deps.Now().Add(-q.since)
	}
	records = logs.FilterRecords(records, filter)
	if q.ref != "" {
		want := strings.ToLower(q.ref)
		matched := records[:0]
		for _, rec := range records {
			if strings.Contains(strings.ToLower(rec.Ref), want) {
				matched = append(matched, rec)
			}
		}
		records = matched
	}
	if len(records) > q.limit {
		records = records[:q.limit]
	}

	label := "Runs"
	if value != "" {
		label = "Runs: " + value
	}
	s := &Section{Type: TypeRuns, Ref: value, Label: label}
	rows := make([]RunSummary, 0, len(records))
	for _, rec := range records {
		rows = append(rows, summarize(rec))
	}
	patterns := failurePatterns(records)
	s.Metadata = map[string]any{"count": len(rows), "runs": rows}
	if len(patterns) > 0 {
		s.Metadata["patterns"] = patterns
	}
	s.addField("Matched", strconv.Itoa(len(rows)))

	body := "No matching runs."
	if len(rows) > 0 {
		body = r.runTable(rows, true)
	}
	if len(patterns) > 0 {
		body += "\n\n" + strings.Join(patterns, "\n")
	}
	s.parts = append(s.parts, &part{body: body, tier: tierFixed})
	return s, nil
}

// failurePatterns notes each ref that failed more than once in the results, most failures first.
func failurePatterns(records []logs.UnifiedRecord) []string {
	type tally struct {
		ref           string
		failed, total int
	}
	byRef := map[string]*tally{}
	var order []*tally
	for _, rec := range records {
		t, ok := byRef[rec.Ref]
		if !ok {
			t = &tally{ref: rec.Ref}
			byRef[rec.Ref] = t
			order = append(order, t)
		}
		t.total++
		if logs.CanonicalStatus(rec) == store.RunFailed {
			t.failed++
		}
	}
	sort.SliceStable(order, func(i, j int) bool { return order[i].failed > order[j].failed })
	var notes []string
	for _, t := range order {
		if t.failed < 2 {
			break
		}
		notes = append(notes, fmt.Sprintf("Note: %s failed %d of its %d runs shown.", t.ref, t.failed, t.total))
	}
	return notes
}

func summarize(rec logs.UnifiedRecord) RunSummary {
	return RunSummary{
		ID:        rec.ID,
		Ref:       rec.Ref,
		Status:    string(logs.CanonicalStatus(rec)),
		ExitCode:  rec.ExitCode,
		Duration:  formatDuration(rec.Duration),
		StartedAt: rec.StartedAt.UTC().Format(time.RFC3339),
		Error:     snippet(rec.Error, 80),
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
		if withRef {
			table = append(table, []string{row.ID, row.Ref, row.Status, exit, row.Duration, when, row.Error})
		} else {
			table = append(table, []string{row.ID, row.Status, exit, row.Duration, when})
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

func snippet(s string, n int) string {
	s, _, _ = strings.Cut(strings.TrimSpace(s), "\n")
	if runeLen(s) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
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

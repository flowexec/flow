// Package digest assembles project state named by source strings (runs, executables, files)
// into one labeled document. The CLI's `flow digest` command and the MCP `get_digest` tool both
// render what Resolve returns.
package digest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flowexec/flow/v2/pkg/cache"
	flowerrors "github.com/flowexec/flow/v2/pkg/errors"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/types/executable"
)

const (
	DefaultMaxChars = 40000
	DefaultMaxLines = 400

	maxLogChars      = 24000
	maxFileChars     = 16000
	defaultRunsLimit = 20
	recentRunsLimit  = 5
)

// Source types, written as the prefix of a source string (`run:<id>`).
const (
	TypeRun  = "run"
	TypeRuns = "runs"
	TypeExec = "exec"
	TypeFile = "file"
)

// Options bounds the size of the resolved document.
type Options struct {
	// MaxChars is the character budget shared by all sections' content. Metadata and fixed
	// content (definitions, tables) are always kept; log, file and script text shrink to fit.
	MaxChars int
	// MaxLines caps the log lines kept per run section.
	MaxLines int
	// NoHistory omits recent runs from exec sections.
	NoHistory bool
}

// Deps are the data sources Resolve reads from.
type Deps struct {
	Store       store.DataStore
	LogsDir     string
	Executables cache.ExecutableCache
	// WorkspaceRoot is where file sources resolve from; files outside it are rejected.
	WorkspaceRoot string
	// CurrentWorkspace and CurrentNamespace fill in exec sources that omit them.
	CurrentWorkspace string
	CurrentNamespace string
	// Now is the reference time for relative timestamps. Defaults to time.Now.
	Now func() time.Time
}

// Document is the resolved digest: one section per source, in the order given.
type Document struct {
	Sections   []Section `json:"sections"   yaml:"sections"`
	TotalChars int       `json:"totalChars" yaml:"totalChars"`
	MaxChars   int       `json:"maxChars"   yaml:"maxChars"`
}

// Section is the resolved state for one source.
type Section struct {
	Type      string         `json:"type"      yaml:"type"`
	Ref       string         `json:"ref"       yaml:"ref"`
	Label     string         `json:"label"     yaml:"label"`
	Metadata  map[string]any `json:"metadata"  yaml:"metadata"`
	Content   string         `json:"content"   yaml:"content"`
	Truncated bool           `json:"truncated" yaml:"truncated"`

	// fields is the metadata in display order for text output.
	fields []field
	// textHeading titles the content in text output only, so JSON content stays raw.
	textHeading string
	parts       []*part
}

type field struct{ key, value string }

// tier orders which content gives way to the budget: fixed content never shrinks, and logs
// shrink before files and scripts.
type tier int

const (
	tierFixed tier = iota
	tierFile
	tierLog
)

type part struct {
	heading   string
	body      string
	tier      tier
	keepTail  bool // logs keep their end, where failures surface; files keep their start
	limit     int
	truncated bool
}

type source struct {
	raw   string
	kind  string
	value string
	runs  runsQuery
	exec  executable.Ref
}

// Resolve resolves each source into a section and fits their content to opts.MaxChars. A
// malformed or unknown source fails the whole call, naming every bad source at once.
func Resolve(ctx context.Context, sources []string, opts Options, deps Deps) (*Document, error) {
	if opts.MaxChars <= 0 {
		opts.MaxChars = DefaultMaxChars
	}
	if opts.MaxLines <= 0 {
		opts.MaxLines = DefaultMaxLines
	}
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if len(sources) == 0 {
		return nil, flowerrors.NewUsageError("at least one source is required, e.g. run:<id>, runs:failed, " +
			"exec:<verb>/<id> or file:<path>")
	}

	parsed, err := parseSources(sources, deps)
	if err != nil {
		return nil, err
	}

	r := &resolver{opts: opts, deps: deps}
	doc := &Document{MaxChars: opts.MaxChars}
	sections := make([]*Section, 0, len(parsed))
	for _, src := range parsed {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		s, err := r.resolve(src)
		if err != nil {
			return nil, err
		}
		sections = append(sections, s)
	}

	applyBudget(sections, opts.MaxChars)
	for _, s := range sections {
		s.assemble()
		doc.TotalChars += runeLen(s.Content)
		doc.Sections = append(doc.Sections, *s)
	}
	return doc, nil
}

func parseSources(sources []string, deps Deps) ([]source, error) {
	var parsed []source
	var problems []string
	for _, raw := range sources {
		src, err := parseSource(raw, deps)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%q: %s", raw, err))
			continue
		}
		parsed = append(parsed, src)
	}
	if len(problems) > 0 {
		return nil, flowerrors.NewUsageError("invalid source(s): %s", strings.Join(problems, "; "))
	}
	return parsed, nil
}

func parseSource(raw string, deps Deps) (source, error) {
	kind, value, ok := strings.Cut(strings.TrimSpace(raw), ":")
	kind = strings.ToLower(strings.TrimSpace(kind))
	value = strings.TrimSpace(value)
	src := source{raw: raw, kind: kind, value: value}
	if !ok {
		return src, fmt.Errorf("expected type:value, where type is run, runs, exec or file")
	}

	var err error
	switch kind {
	case TypeRun, TypeFile:
		if value == "" {
			err = fmt.Errorf("%s source needs a value", kind)
		}
	case TypeRuns:
		src.runs, err = parseRunsQuery(value)
	case TypeExec:
		src.exec, err = parseExecRef(value, deps.CurrentWorkspace, deps.CurrentNamespace)
	default:
		err = fmt.Errorf("unknown type %q; expected run, runs, exec or file", kind)
	}
	return src, err
}

type resolver struct {
	opts Options
	deps Deps
}

func (r *resolver) resolve(src source) (*Section, error) {
	switch src.kind {
	case TypeRun:
		return r.runSection(src.value)
	case TypeRuns:
		return r.runsSection(src.value, src.runs)
	case TypeExec:
		return r.execSection(src.exec)
	case TypeFile:
		return r.fileSection(src.value)
	}
	return nil, fmt.Errorf("unhandled source type %q", src.kind)
}

func (s *Section) addField(key, value string) {
	if value == "" {
		return
	}
	s.fields = append(s.fields, field{key: key, value: value})
}

// assemble joins the (budgeted) parts into Content.
func (s *Section) assemble() {
	var blocks []string
	for _, p := range s.parts {
		if p.truncated {
			s.Truncated = true
		}
		switch {
		case p.heading != "" && p.body != "":
			blocks = append(blocks, "### "+p.heading+"\n\n"+p.body)
		case p.heading != "":
			blocks = append(blocks, "### "+p.heading)
		case p.body != "":
			blocks = append(blocks, p.body)
		}
	}
	s.Content = strings.Join(blocks, "\n\n")
}

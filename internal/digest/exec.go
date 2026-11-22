package digest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flowexec/flow/v2/internal/io/logs"
	flowerrors "github.com/flowexec/flow/v2/pkg/errors"
	"github.com/flowexec/flow/v2/types/executable"
)

// parseExecRef accepts what `flow browse` does, as `<verb> <id>` or `<verb>/<id>`, filling in the
// current workspace and namespace when the ID omits them.
func parseExecRef(value, curWs, curNs string) (executable.Ref, error) {
	if value == "" {
		return "", fmt.Errorf("exec source needs a verb and optional ID, e.g. exec:test/unit")
	}
	verbStr, id := value, ""
	if i := strings.IndexAny(value, " /"); i >= 0 {
		verbStr, id = value[:i], strings.TrimSpace(value[i+1:])
	}
	verb := executable.Verb(verbStr)
	if err := verb.Validate(); err != nil {
		return "", err
	}
	ws, ns, name, err := executable.ParseExecutableID(id)
	if err != nil {
		return "", err
	}
	if ws == executable.WildcardWorkspace || ws == executable.CurrentWorkspace {
		ws = curWs
	}
	if ns == executable.WildcardNamespace {
		ns = curNs
	}
	return executable.NewRef(executable.NewExecutableID(ws, ns, name), verb), nil
}

func (r *resolver) execSection(ref executable.Ref) (*Section, error) {
	e, err := r.lookupExecutable(ref)
	if err != nil {
		return nil, err
	}
	def, err := e.YAML()
	if err != nil {
		return nil, err
	}

	s := &Section{Type: TypeExec, Ref: ref.String(), Label: "Executable: " + e.Ref().String()}
	s.Metadata = map[string]any{
		"execRef":   e.Ref().String(),
		"verb":      string(e.Verb),
		"id":        e.ID(),
		"workspace": e.Workspace(),
	}
	setIf(s.Metadata, "namespace", e.Namespace())
	setIf(s.Metadata, "flowFile", relTo(e.WorkspacePath(), e.FlowFilePath()))
	if len(e.Tags) > 0 {
		s.Metadata["tags"] = []string(e.Tags)
	}

	s.addField("Verb", string(e.Verb))
	s.addField("Workspace", e.Workspace())
	if len(e.Tags) > 0 {
		s.addField("Tags", "["+strings.Join(e.Tags, ", ")+"]")
	}
	s.addField("Flow file", relTo(e.WorkspacePath(), e.FlowFilePath()))

	s.parts = append(s.parts, &part{heading: "Definition", body: strings.TrimSuffix(def, "\n"), tier: tierFixed})

	if path, ok := scriptPath(e); ok {
		if body, err := readText(path); err == nil {
			s.Metadata["script"] = path
			s.parts = append(s.parts, &part{
				heading: "Script: " + relTo(e.WorkspacePath(), path),
				body:    body, tier: tierFile, limit: maxFileChars,
			})
		}
	}

	if !r.opts.NoHistory {
		records, err := logs.LoadRecordsForRefs(r.deps.Store, r.deps.LogsDir, historyRefs(e), recentRunsLimit)
		if err != nil {
			return nil, fmt.Errorf("reading history for %s: %w", e.Ref(), err)
		}
		if len(records) > recentRunsLimit {
			records = records[:recentRunsLimit]
		}
		rows := make([]RunSummary, 0, len(records))
		for _, rec := range records {
			rows = append(rows, summarize(rec))
		}
		s.Metadata["recentRuns"] = rows
		body := "No recorded runs."
		if len(rows) > 0 {
			body = r.runTable(rows, false)
		}
		s.parts = append(s.parts, &part{
			heading: fmt.Sprintf("Recent runs (last %d)", recentRunsLimit), body: body, tier: tierFixed,
		})
	}
	return s, nil
}

// historyRefs lists every ref the executable's runs may be recorded under: runs keep the verb
// that was typed, so `flow exec x` and `flow run x` land under different refs.
func historyRefs(e *executable.Executable) []string {
	verbs := append([]executable.Verb{e.Verb}, executable.RelatedVerbs(e.Verb)...)
	verbs = append(verbs, e.VerbAliases...)
	seen := map[string]bool{}
	refs := make([]string, 0, len(verbs))
	for _, v := range verbs {
		ref := executable.NewRef(e.ID(), v).String()
		if !seen[ref] {
			seen[ref] = true
			refs = append(refs, ref)
		}
	}
	return refs
}

// lookupExecutable reads the cache, syncing it once on a miss, as `flow browse` does.
func (r *resolver) lookupExecutable(ref executable.Ref) (*executable.Executable, error) {
	if r.deps.Executables == nil {
		return nil, errors.New("executable cache is not available")
	}
	e, err := r.deps.Executables.GetExecutableByRef(ref)
	var notFound flowerrors.ExecutableNotFoundError
	if err != nil && errors.As(err, &notFound) {
		if err := r.deps.Executables.Update(); err != nil {
			return nil, err
		}
		e, err = r.deps.Executables.GetExecutableByRef(ref)
	}
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, flowerrors.NewExecutableNotFoundError(ref.String())
	}
	return e, nil
}

// scriptPath resolves an exec executable's `file` the way the exec runner does, for the cases
// that don't depend on runtime state (env expansion, temp dirs). The script must sit inside the
// executable's workspace.
func scriptPath(e *executable.Executable) (string, bool) {
	if e.Exec == nil || e.Exec.File == "" {
		return "", false
	}
	dir := string(e.Exec.Dir)
	base := filepath.Dir(e.FlowFilePath())
	switch {
	case dir == executable.TmpDirLabel, strings.Contains(dir, "$"), strings.Contains(e.Exec.File, "$"):
		return "", false
	case dir == "":
		dir = base
	case strings.HasPrefix(dir, "//"):
		dir = filepath.Join(e.WorkspacePath(), filepath.FromSlash(strings.TrimPrefix(dir, "//")))
	case strings.HasPrefix(dir, "~/"):
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		dir = filepath.Join(home, dir[2:])
	case !filepath.IsAbs(dir):
		dir = filepath.Join(base, dir)
	}
	path := e.Exec.File
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", false
	}
	root, err := filepath.EvalSymlinks(e.WorkspacePath())
	if err != nil || !within(root, resolved) {
		return "", false
	}
	return resolved, true
}

func relTo(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}

package digest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	flowerrors "github.com/flowexec/flow/v2/pkg/errors"
)

// maxFileRead bounds how much of a file is read; anything past it would be cut by the budget.
const maxFileRead = 1 << 20

var errBinary = errors.New("binary file")

func (r *resolver) fileSection(p string) (*Section, error) {
	root := r.deps.WorkspaceRoot
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		root = wd
	}
	rootReal, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolving workspace root %s: %w", root, err)
	}

	target := p
	if !filepath.IsAbs(target) {
		target = filepath.Join(root, target)
	}
	resolved, err := filepath.EvalSymlinks(target)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, flowerrors.NewFileNotFoundError(p)
	} else if err != nil {
		return nil, err
	}
	if !within(rootReal, resolved) {
		return nil, flowerrors.NewUsageError("file %s is outside the workspace root %s", p, root)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, flowerrors.NewUsageError("file %s is a directory", p)
	}

	rel := relTo(rootReal, resolved)
	s := &Section{Type: TypeFile, Ref: p, Label: "File: " + rel}
	s.Metadata = map[string]any{"path": rel, "sizeBytes": info.Size()}
	s.addField("Path", rel)
	s.addField("Size", fmt.Sprintf("%d bytes", info.Size()))

	body, err := readText(resolved)
	if errors.Is(err, errBinary) {
		s.Metadata["binary"] = true
		s.addField("Contents", "omitted (binary file)")
		return s, nil
	} else if err != nil {
		return nil, err
	}
	s.textHeading = "Contents"
	s.parts = append(s.parts, &part{
		body: body, tier: tierFile, limit: maxFileChars, truncated: info.Size() > maxFileRead,
	})
	return s, nil
}

// readText reads up to maxFileRead bytes of a text file.
func readText(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxFileRead))
	if err != nil {
		return "", err
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return "", errBinary
	}
	text := string(data)
	if !utf8.ValidString(text) {
		text = strings.ToValidUTF8(text, "�")
	}
	return strings.TrimSuffix(text, "\n"), nil
}

// within reports whether path is root or inside it. Both must already be symlink-resolved.
func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

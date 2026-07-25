package logs

import (
	"fmt"
	"regexp"
	"strings"
)

// ContentOptions controls how a record's log output is sliced for display.
// The operations apply in order — grep filter, then tail, then byte cap — and
// each limit keeps the END of the log, since failures surface there.
type ContentOptions struct {
	// Tail limits output to the last N lines (applied after Grep). 0 means no line limit.
	Tail int
	// Grep keeps only lines matching this regular expression. Empty means no filter.
	Grep string
	// MaxBytes caps the returned content to the last N bytes, trimmed to a line
	// boundary where possible. 0 means no byte cap.
	MaxBytes int
}

// ContentResult is the outcome of slicing a record's raw log output.
type ContentResult struct {
	Content       string
	TotalLines    int  // lines in the raw log (before any filtering)
	ReturnedLines int  // lines in the returned content
	Truncated     bool // returned content omits part of the log due to the tail/byte limits
}

// ExtractContent applies the slicing options to a raw log string. With zero-valued
// options it returns the content unchanged. It is safe to call on partial logs from
// still-running executions — it simply operates on whatever has been written so far.
func ExtractContent(raw string, opts ContentOptions) (ContentResult, error) {
	var res ContentResult
	if raw == "" {
		return res, nil
	}

	// Split into lines without a trailing empty element from a final newline.
	lines := strings.Split(strings.TrimSuffix(raw, "\n"), "\n")
	res.TotalLines = len(lines)

	if opts.Grep != "" {
		re, err := regexp.Compile(opts.Grep)
		if err != nil {
			return res, fmt.Errorf("invalid grep pattern %q: %w", opts.Grep, err)
		}
		matched := make([]string, 0, len(lines))
		for _, ln := range lines {
			if re.MatchString(ln) {
				matched = append(matched, ln)
			}
		}
		lines = matched
	}

	if opts.Tail > 0 && len(lines) > opts.Tail {
		lines = lines[len(lines)-opts.Tail:]
		res.Truncated = true
	}

	content := strings.Join(lines, "\n")

	if opts.MaxBytes > 0 && len(content) > opts.MaxBytes {
		// Keep the tail: trim from the front, advancing to the next line boundary
		// so the leading line isn't left mangled (unless there is no newline left).
		cut := len(content) - opts.MaxBytes
		if nl := strings.IndexByte(content[cut:], '\n'); nl >= 0 {
			cut += nl + 1
		}
		content = content[cut:]
		res.Truncated = true
	}

	res.Content = content
	if content != "" {
		res.ReturnedLines = strings.Count(content, "\n") + 1
	}
	return res, nil
}

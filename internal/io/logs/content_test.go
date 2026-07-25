package logs_test

import (
	"strings"
	"testing"

	"github.com/flowexec/flow/v2/internal/io/logs"
)

func TestExtractContent_EmptyReturnsZeroValue(t *testing.T) {
	res, err := logs.ExtractContent("", logs.ContentOptions{Tail: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "" || res.TotalLines != 0 || res.ReturnedLines != 0 || res.Truncated {
		t.Fatalf("expected zero-valued result, got %+v", res)
	}
}

func TestExtractContent_NoOptionsReturnsFull(t *testing.T) {
	raw := "a\nb\nc"
	res, err := logs.ExtractContent(raw, logs.ContentOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != raw {
		t.Fatalf("expected content %q, got %q", raw, res.Content)
	}
	if res.TotalLines != 3 || res.ReturnedLines != 3 || res.Truncated {
		t.Fatalf("unexpected result %+v", res)
	}
}

func TestExtractContent_TrailingNewlineNotCountedAsLine(t *testing.T) {
	res, err := logs.ExtractContent("a\nb\n", logs.ContentOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TotalLines != 2 {
		t.Fatalf("expected 2 lines, got %d", res.TotalLines)
	}
}

func TestExtractContent_TailKeepsLastLines(t *testing.T) {
	raw := "1\n2\n3\n4\n5"
	res, err := logs.ExtractContent(raw, logs.ContentOptions{Tail: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "4\n5" {
		t.Fatalf("expected last two lines, got %q", res.Content)
	}
	if !res.Truncated {
		t.Fatalf("expected truncated=true")
	}
	if res.TotalLines != 5 || res.ReturnedLines != 2 {
		t.Fatalf("unexpected counts %+v", res)
	}
}

func TestExtractContent_TailLargerThanLogNotTruncated(t *testing.T) {
	res, err := logs.ExtractContent("1\n2", logs.ContentOptions{Tail: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Truncated {
		t.Fatalf("expected not truncated when tail exceeds line count")
	}
	if res.Content != "1\n2" {
		t.Fatalf("expected full content, got %q", res.Content)
	}
}

func TestExtractContent_GrepFiltersLines(t *testing.T) {
	raw := "info: ok\nERROR: boom\ninfo: still ok\nERROR: kaboom"
	res, err := logs.ExtractContent(raw, logs.ContentOptions{Grep: "^ERROR"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "ERROR: boom\nERROR: kaboom" {
		t.Fatalf("unexpected grep result %q", res.Content)
	}
	// Grep alone is a filter, not truncation.
	if res.Truncated {
		t.Fatalf("expected grep-only result to not be marked truncated")
	}
	if res.TotalLines != 4 || res.ReturnedLines != 2 {
		t.Fatalf("unexpected counts %+v", res)
	}
}

func TestExtractContent_GrepThenTail(t *testing.T) {
	raw := "ERROR: 1\nok\nERROR: 2\nok\nERROR: 3"
	res, err := logs.ExtractContent(raw, logs.ContentOptions{Grep: "ERROR", Tail: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "ERROR: 2\nERROR: 3" {
		t.Fatalf("expected last two matching lines, got %q", res.Content)
	}
	if !res.Truncated {
		t.Fatalf("expected truncated after tailing matches")
	}
}

func TestExtractContent_InvalidGrepReturnsError(t *testing.T) {
	_, err := logs.ExtractContent("a\nb", logs.ContentOptions{Grep: "("})
	if err == nil {
		t.Fatalf("expected error for invalid regexp")
	}
}

func TestExtractContent_MaxBytesKeepsTailOnLineBoundary(t *testing.T) {
	// Lines of 4 bytes each ("aaa\n"). Cap to keep only the final line's worth.
	raw := "aaaa\nbbbb\ncccc"
	res, err := logs.ExtractContent(raw, logs.ContentOptions{MaxBytes: 6})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "cccc" {
		t.Fatalf("expected trailing line aligned to boundary, got %q", res.Content)
	}
	if !res.Truncated {
		t.Fatalf("expected truncated=true when byte cap trims content")
	}
	if strings.Contains(res.Content, "\n") && res.ReturnedLines != strings.Count(res.Content, "\n")+1 {
		t.Fatalf("returned line count mismatch: %+v", res)
	}
}

func TestExtractContent_MaxBytesKeepsPartialLineWhenNoBoundary(t *testing.T) {
	// A single very long line with no newline: keep the last MaxBytes bytes rather than nothing.
	raw := strings.Repeat("x", 100)
	res, err := logs.ExtractContent(raw, logs.ContentOptions{MaxBytes: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Content) != 10 {
		t.Fatalf("expected 10 trailing bytes, got %d", len(res.Content))
	}
	if !res.Truncated {
		t.Fatalf("expected truncated=true")
	}
}

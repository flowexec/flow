package common_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/flowexec/flow/internal/io/common"
)

func TestTagColor_Deterministic(t *testing.T) {
	a := common.TagColor("deployment")
	b := common.TagColor("deployment")
	if a != b {
		t.Fatalf("expected deterministic color for the same tag, got %v and %v", a, b)
	}
}

func TestTagColor_DistinctTagsUsuallyDiffer(t *testing.T) {
	// Different tags typically produce different colors; collisions are possible
	// but extremely unlikely across a handful of common values.
	tags := []string{"a", "build", "deploy", "test", "release", "infra", "docs"}
	seen := make(map[string]int, len(tags))
	for _, t := range tags {
		seen[fmt.Sprintf("%v", common.TagColor(t))]++
	}
	if len(seen) < len(tags)-1 {
		t.Fatalf("expected nearly-unique colors across %d tags, got %d distinct", len(tags), len(seen))
	}
}

func TestColorizeTags_EmptyReturnsEmpty(t *testing.T) {
	if got := common.ColorizeTags(nil); got != "" {
		t.Fatalf("expected empty string for nil tags, got %q", got)
	}
	if got := common.ColorizeTags([]string{}); got != "" {
		t.Fatalf("expected empty string for empty tags, got %q", got)
	}
}

func TestColorizeTags_SortsAndJoinsWithComma(t *testing.T) {
	// Output contains ANSI styling, but sorted tag names should still appear in order,
	// separated by ", ".
	got := common.ColorizeTags([]string{"cherry", "apple", "banana"})
	iApple := strings.Index(got, "apple")
	iBanana := strings.Index(got, "banana")
	iCherry := strings.Index(got, "cherry")
	if iApple < 0 || iBanana < 0 || iCherry < 0 {
		t.Fatalf("expected all tag names in output, got %q", got)
	}
	if iApple >= iBanana || iBanana >= iCherry {
		t.Fatalf("expected alphabetical order apple < banana < cherry in %q", got)
	}
	if strings.Count(got, ", ") != 2 {
		t.Fatalf("expected 2 comma separators between 3 tags, got %q", got)
	}
}

func TestColorizeTags_DoesNotMutateInput(t *testing.T) {
	input := []string{"zebra", "apple"}
	_ = common.ColorizeTags(input)
	if input[0] != "zebra" || input[1] != "apple" {
		t.Fatalf("ColorizeTags mutated input slice: %v", input)
	}
}

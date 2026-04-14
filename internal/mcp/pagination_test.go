//nolint:testpackage // tests unexported pagination helpers
package mcp

import (
	"testing"
)

//nolint:gocognit
func TestPaginate(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	t.Run("empty input", func(t *testing.T) {
		page, next, total, err := paginate([]int{}, "", 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(page) != 0 {
			t.Errorf("expected empty page, got %d items", len(page))
		}
		if next != "" {
			t.Errorf("expected empty next cursor, got %q", next)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
	})

	t.Run("first page with more pages", func(t *testing.T) {
		page, next, total, err := paginate(items, "", 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(page) != 3 {
			t.Errorf("expected page size 3, got %d", len(page))
		}
		if next == "" {
			t.Errorf("expected non-empty next cursor")
		}
		if total != 10 {
			t.Errorf("expected total 10, got %d", total)
		}
	})

	t.Run("middle page", func(t *testing.T) {
		// Decode: offset 3, next 6
		cursor := encodeCursor(3)
		page, next, _, err := paginate(items, cursor, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(page) != 3 {
			t.Errorf("expected page size 3, got %d", len(page))
		}
		if page[0] != 4 {
			t.Errorf("expected first item 4, got %d", page[0])
		}
		if next == "" {
			t.Errorf("expected non-empty next cursor")
		}
	})

	t.Run("last page no next cursor", func(t *testing.T) {
		cursor := encodeCursor(8)
		page, next, _, err := paginate(items, cursor, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(page) != 2 {
			t.Errorf("expected page size 2, got %d", len(page))
		}
		if next != "" {
			t.Errorf("expected empty next cursor, got %q", next)
		}
	})

	t.Run("cursor past end", func(t *testing.T) {
		cursor := encodeCursor(100)
		page, next, total, err := paginate(items, cursor, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(page) != 0 {
			t.Errorf("expected empty page, got %d items", len(page))
		}
		if next != "" {
			t.Errorf("expected empty next cursor, got %q", next)
		}
		if total != 10 {
			t.Errorf("expected total 10, got %d", total)
		}
	})

	t.Run("invalid cursor returns error", func(t *testing.T) {
		_, _, _, err := paginate(items, "not-base64!", 3)
		if err == nil {
			t.Errorf("expected error for invalid cursor")
		}
	})

	t.Run("zero page size uses default", func(t *testing.T) {
		page, _, _, err := paginate(items, "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// All 10 items fit within default page size of 25
		if len(page) != 10 {
			t.Errorf("expected all 10 items, got %d", len(page))
		}
	})
}

func TestCursorRoundTrip(t *testing.T) {
	for _, offset := range []int{0, 1, 100, 99999} {
		cursor := encodeCursor(offset)
		decoded, err := decodeCursor(cursor)
		if err != nil {
			t.Fatalf("failed to decode cursor for offset %d: %v", offset, err)
		}
		if decoded != offset {
			t.Errorf("cursor round-trip failed: expected %d, got %d", offset, decoded)
		}
	}
}

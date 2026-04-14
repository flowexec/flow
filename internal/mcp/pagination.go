package mcp

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

const defaultPageSize = 25

// paginate applies cursor-based pagination to a slice of items.
// The cursor is an opaque base64-encoded offset. An empty cursor starts from the beginning.
// Returns the page of items, the next cursor (empty if on the last page), and the total count.
func paginate[T any](items []T, cursor string, pageSize int) (page []T, nextCursor string, totalCount int, err error) {
	totalCount = len(items)
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	offset := 0
	if cursor != "" {
		offset, err = decodeCursor(cursor)
		if err != nil {
			return nil, "", totalCount, fmt.Errorf("invalid cursor: %w", err)
		}
	}

	if offset >= totalCount {
		return []T{}, "", totalCount, nil
	}

	end := offset + pageSize
	if end > totalCount {
		end = totalCount
	}

	page = items[offset:end]

	if end < totalCount {
		nextCursor = encodeCursor(end)
	}

	return page, nextCursor, totalCount, nil
}

func encodeCursor(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodeCursor(cursor string) (int, error) {
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

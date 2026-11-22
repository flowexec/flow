package digest

import (
	"sort"
	"strings"
	"unicode/utf8"
)

// applyBudget shrinks elastic content so all sections fit maxChars. Fixed content is counted
// first and never cut; files and scripts then share what is left, and logs take the remainder.
// Within a tier the budget is split fairly: small parts keep everything and the largest are cut.
func applyBudget(sections []*Section, maxChars int) {
	remaining := maxChars
	var files, logParts []*part
	for _, s := range sections {
		for _, p := range s.parts {
			switch p.tier {
			case tierFixed:
				remaining -= runeLen(p.body)
			case tierFile:
				files = append(files, p)
			case tierLog:
				logParts = append(logParts, p)
			}
		}
	}
	remaining = max(remaining, 0)
	remaining = fairShare(files, remaining)
	fairShare(logParts, remaining)
}

// fairShare cuts parts to fit budget, giving each the lesser of its demand and an even share of
// what is left, smallest demand first. It returns the unused budget.
func fairShare(parts []*part, budget int) int {
	demand := func(p *part) int {
		n := runeLen(p.body)
		if p.limit > 0 && n > p.limit {
			return p.limit
		}
		return n
	}
	sort.SliceStable(parts, func(i, j int) bool { return demand(parts[i]) < demand(parts[j]) })
	for i, p := range parts {
		share := budget / (len(parts) - i)
		p.cut(min(demand(p), share))
		budget -= runeLen(p.body)
	}
	return budget
}

// cut trims the part's body to at most n characters on a line boundary, keeping its tail or head.
func (p *part) cut(n int) {
	if runeLen(p.body) <= n {
		return
	}
	p.truncated = true
	if p.keepTail {
		p.body = keepTail(p.body, n)
	} else {
		p.body = keepHead(p.body, n)
	}
}

func keepTail(s string, n int) string {
	if n <= 0 {
		return ""
	}
	start := len(s)
	for i := 0; i < n && start > 0; i++ {
		_, size := utf8.DecodeLastRuneInString(s[:start])
		start -= size
	}
	out := s[start:]
	// Drop the partial first line unless it is the only one.
	if start > 0 && s[start-1] != '\n' {
		if nl := strings.IndexByte(out, '\n'); nl >= 0 {
			out = out[nl+1:]
		}
	}
	return out
}

func keepHead(s string, n int) string {
	if n <= 0 {
		return ""
	}
	end := 0
	for i := 0; i < n && end < len(s); i++ {
		_, size := utf8.DecodeRuneInString(s[end:])
		end += size
	}
	out := s[:end]
	if end < len(s) && s[end] != '\n' {
		if nl := strings.LastIndexByte(out, '\n'); nl >= 0 {
			out = out[:nl]
		}
	}
	return strings.TrimSuffix(out, "\n")
}

func runeLen(s string) int { return utf8.RuneCountInString(s) }

package common

import (
	"fmt"
	"hash/fnv"
	"image/color"
	"math"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

// TagColor derives a deterministic color from the tag text using FNV hashing
// mapped to the HSL hue wheel with fixed saturation/lightness.
func TagColor(tag string) color.Color {
	h := fnv.New32a()
	h.Write([]byte(tag))
	hue := float64(h.Sum32()%360) / 360.0
	r, g, b := hslToRGB(hue, 0.55, 0.65)
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}

// ColorizeTags sorts tags and renders each as a colored pill badge.
func ColorizeTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	sorted := make([]string, len(tags))
	copy(sorted, tags)
	slices.Sort(sorted)
	parts := make([]string, len(sorted))
	for i, t := range sorted {
		c := TagColor(t)
		parts[i] = lipgloss.NewStyle().Foreground(c).Render(t)
	}
	return strings.Join(parts, ", ")
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		v := uint8(math.Round(l * 255))
		return v, v, v
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	r := hueToRGB(p, q, h+1.0/3.0)
	g := hueToRGB(p, q, h)
	b := hueToRGB(p, q, h-1.0/3.0)
	return uint8(math.Round(r * 255)), uint8(math.Round(g * 255)), uint8(math.Round(b * 255))
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6
	default:
		return p
	}
}

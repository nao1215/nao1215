package main

import "strings"

// helvetica holds advance widths of printable ASCII (0x20-0x7E) in 1/1000 em,
// taken from the Helvetica AFM. GitHub renders the cards with the viewer's
// system sans-serif font, so these are an estimate, not an exact measurement.
var helvetica = [95]int{
	278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278, // ' ' - '/'
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, // 0-9
	278, 278, 584, 584, 584, 556, 1015, // : - @
	667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, // A-M
	722, 778, 667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, // N-Z
	278, 278, 278, 469, 556, 333, // [ - `
	556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, // a-m
	556, 556, 556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, // n-z
	334, 260, 334, 584, // { - ~
}

// widthSafety widens the Helvetica estimate. Segoe UI and Noto Sans, which
// Windows and Linux viewers get, run up to about 10% wider than Helvetica.
const widthSafety = 1.1

// textWidth estimates the rendered width of s in pixels at the given size.
// Characters outside ASCII (CJK, emoji) are counted as one full em.
func textWidth(s string, size float64, bold bool) float64 {
	total := 0
	for _, r := range s {
		if r >= 0x20 && r <= 0x7E {
			total += helvetica[r-0x20]
		} else {
			total += 1000
		}
	}
	w := float64(total) / 1000 * size * widthSafety
	if bold {
		w *= 1.06
	}
	return w
}

const ellipsis = "…"

// wrapText splits s into at most maxLines lines no wider than maxWidth. When
// the text does not fit, the last line ends with an ellipsis.
func wrapText(s string, size, maxWidth float64, maxLines int) []string {
	words := strings.Fields(s)
	if len(words) == 0 || maxLines <= 0 {
		return nil
	}

	var lines []string
	cur := words[0]
	for _, w := range words[1:] {
		if textWidth(cur+" "+w, size, false) <= maxWidth {
			cur += " " + w
			continue
		}
		lines = append(lines, cur)
		cur = w
	}
	lines = append(lines, cur)

	if len(lines) > maxLines {
		rest := strings.Join(lines[maxLines-1:], " ")
		lines = append(lines[:maxLines-1], rest)
	}
	for i, l := range lines {
		lines[i] = truncate(l, size, maxWidth)
	}
	return lines
}

// truncate shortens s with an ellipsis so it fits in maxWidth. It prefers to
// cut at a word boundary.
func truncate(s string, size, maxWidth float64) string {
	if textWidth(s, size, false) <= maxWidth {
		return s
	}
	runes := []rune(s)
	for n := len(runes) - 1; n > 0; n-- {
		head := strings.TrimRight(string(runes[:n]), " ")
		if textWidth(head+ellipsis, size, false) > maxWidth {
			continue
		}
		if sp := strings.LastIndex(head, " "); sp > len(head)/2 {
			head = head[:sp]
		}
		return strings.TrimRight(head, " ,.;:") + ellipsis
	}
	return ellipsis
}

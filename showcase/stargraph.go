package main

import (
	"fmt"
	"html"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	graphWidth  = 808
	graphHeight = 300
	plotLeft    = 56
	plotRight   = graphWidth - 112 // leaves a gutter for the end labels
	plotTop     = 76
	plotBottom  = graphHeight - 36
	// maxStarSeries is how many lines the graph can tell apart; it matches
	// the length of Theme.Series.
	maxStarSeries = 6
	labelGap      = 14.0 // minimum vertical distance between end labels
)

// StarSeries is one line of the star history graph.
type StarSeries struct {
	Name  string
	Dates []time.Time // oldest first
}

// renderStarGraph draws the cumulative stars of each repository over time as
// a standalone SVG document. now is the right edge of the time axis.
func renderStarGraph(series []StarSeries, now time.Time, t Theme) string {
	var b strings.Builder
	esc := html.EscapeString
	if len(series) > len(t.Series) {
		series = series[:len(t.Series)]
	}

	names := make([]string, len(series))
	for i, s := range series {
		names[i] = fmt.Sprintf("%s %s", s.Name, grouped(len(s.Dates)))
	}
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="Star history">`+"\n",
		graphWidth, graphHeight, graphWidth, graphHeight)
	fmt.Fprintf(&b, "<title>Star history: %s</title>\n", esc(strings.Join(names, ", ")))
	fmt.Fprintf(&b, `<style>text{font-family:%s}</style>`+"\n", fontStack)
	fmt.Fprintf(&b, `<rect x="0.75" y="0.75" width="%s" height="%s" rx="8" fill="%s" stroke="%s"/>`+"\n",
		num(graphWidth-1.5), num(graphHeight-1.5), t.Bg, t.Border)

	b.WriteString(icon(iconStar, padX, 17, 1, t.Muted))
	fmt.Fprintf(&b, `<text x="%d" y="30" font-size="%s" font-weight="600" fill="%s">Star history</text>`+"\n", padX+22, num(titleSize), t.Text)

	// Legend: color swatch, name, and current count of every line.
	lx := float64(padX)
	for i, s := range series {
		fmt.Fprintf(&b, `<rect x="%s" y="45" width="12" height="4" rx="2" fill="%s"/>`+"\n", num(lx), t.Series[i])
		lx += 17
		fmt.Fprintf(&b, `<text x="%s" y="51" font-size="%s" fill="%s">%s</text>`+"\n", num(lx), num(metaSize), t.Text, esc(s.Name))
		lx += textWidth(s.Name, metaSize, false) + 5
		count := grouped(len(s.Dates))
		fmt.Fprintf(&b, `<text x="%s" y="51" font-size="%s" fill="%s">%s</text>`+"\n", num(lx), num(metaSize), t.Muted, count)
		lx += textWidth(count, metaSize, false) + 18
	}

	var start time.Time
	most := 0
	for _, s := range series {
		if len(s.Dates) == 0 {
			continue
		}
		if start.IsZero() || s.Dates[0].Before(start) {
			start = s.Dates[0]
		}
		most = max(most, len(s.Dates))
	}
	if most == 0 {
		fmt.Fprintf(&b, `<text x="%d" y="%d" font-size="%s" text-anchor="middle" fill="%s">No stars yet</text>`+"\n",
			graphWidth/2, (plotTop+plotBottom)/2, num(metaSize), t.Muted)
		b.WriteString("</svg>\n")
		return b.String()
	}
	if !now.After(start) {
		now = start.Add(24 * time.Hour)
	}
	span := now.Sub(start).Seconds()
	xOf := func(tm time.Time) float64 {
		return plotLeft + tm.Sub(start).Seconds()/span*(plotRight-plotLeft)
	}
	step, yMax := niceScale(most)
	yOf := func(v int) float64 {
		return plotBottom - float64(v)/float64(yMax)*(plotBottom-plotTop)
	}

	// Recessive grid and y labels.
	for v := 0; v <= yMax; v += step {
		y := yOf(v)
		fmt.Fprintf(&b, `<line x1="%d" y1="%s" x2="%d" y2="%s" stroke="%s" stroke-width="1"/>`+"\n", plotLeft, num(y), plotRight, num(y), t.Border)
		fmt.Fprintf(&b, `<text x="%d" y="%s" font-size="11" text-anchor="end" fill="%s">%s</text>`+"\n", plotLeft-8, num(y+4), t.Muted, compact(v))
	}
	// One tick per year, thinned so the labels never crowd.
	years := yearTicks(start, now)
	every := (len(years) + 7) / 8
	for i, yr := range years {
		if i%every != 0 {
			continue
		}
		x := xOf(yr)
		fmt.Fprintf(&b, `<line x1="%s" y1="%d" x2="%s" y2="%d" stroke="%s" stroke-width="1"/>`+"\n", num(x), plotBottom, num(x), plotBottom+4, t.Border)
		fmt.Fprintf(&b, `<text x="%s" y="%d" font-size="11" text-anchor="middle" fill="%s">%d</text>`+"\n", num(x), plotBottom+18, t.Muted, yr.Year())
	}

	// Lines, smallest last so it is drawn on top of the crowd near zero.
	order := make([]int, 0, len(series))
	for i, s := range series {
		if len(s.Dates) > 0 {
			order = append(order, i)
		}
	}
	sort.SliceStable(order, func(a, c int) bool { return len(series[order[a]].Dates) > len(series[order[c]].Dates) })
	for _, i := range order {
		fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="%s" stroke-width="2" stroke-linejoin="round"/>`+"\n",
			stepPath(series[i].Dates, xOf, yOf, now), t.Series[i])
	}

	// End labels in the right gutter, pushed apart so they never overlap.
	ys := make([]float64, len(order))
	for k, i := range order {
		ys[k] = yOf(len(series[i].Dates))
	}
	labels := spreadLabels(ys, plotTop, plotBottom, labelGap)
	for k, i := range order {
		x := num(xOf(now))
		fmt.Fprintf(&b, `<circle cx="%s" cy="%s" r="4" fill="%s" stroke="%s" stroke-width="2"/>`+"\n", x, num(ys[k]), t.Series[i], t.Bg)
		fmt.Fprintf(&b, `<text x="%d" y="%s" font-size="11" fill="%s">%s</text>`+"\n",
			plotRight+10, num(labels[k]+4), t.Text, esc(truncate(series[i].Name, 11, graphWidth-plotRight-10-padX)))
	}

	b.WriteString("</svg>\n")
	return b.String()
}

// stepPath draws the cumulative count of dates as a step line, one vertex
// pair per day with stars, continuing flat to now.
func stepPath(dates []time.Time, xOf func(time.Time) float64, yOf func(int) float64, now time.Time) string {
	var p strings.Builder
	fmt.Fprintf(&p, "M%s %s", num(xOf(dates[0])), num(yOf(0)))
	for i := 0; i < len(dates); {
		day := dates[i].Truncate(24 * time.Hour)
		j := i
		for j < len(dates) && dates[j].Truncate(24*time.Hour).Equal(day) {
			j++
		}
		fmt.Fprintf(&p, "H%sV%s", num(xOf(dates[i])), num(yOf(j)))
		i = j
	}
	fmt.Fprintf(&p, "H%s", num(xOf(now)))
	return p.String()
}

// spreadLabels moves label positions as little as possible so that sorted
// neighbors are at least gap apart and all stay within [top, bottom]. ys is
// in any order; the result lines up with it.
func spreadLabels(ys []float64, top, bottom, gap float64) []float64 {
	idx := make([]int, len(ys))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, c int) bool { return ys[idx[a]] < ys[idx[c]] })
	pos := make([]float64, len(ys))
	for k, i := range idx {
		pos[i] = ys[i]
		if k > 0 {
			pos[i] = math.Max(pos[i], pos[idx[k-1]]+gap)
		}
	}
	// Pull back up from the bottom edge if the push went too far.
	for k := len(idx) - 1; k >= 0; k-- {
		i := idx[k]
		limit := bottom
		if k < len(idx)-1 {
			limit = pos[idx[k+1]] - gap
		}
		pos[i] = math.Max(math.Min(pos[i], limit), top)
	}
	return pos
}

// niceScale picks a 1-2-5 grid step giving about four intervals, and the axis
// maximum it implies.
func niceScale(max int) (step, top int) {
	raw := float64(max) / 4
	if raw < 1 {
		raw = 1
	}
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	for _, m := range []float64{1, 2, 5, 10} {
		if s := m * mag; s >= raw {
			step = int(s)
			break
		}
	}
	top = (max + step - 1) / step * step
	if top == max {
		top += step
	}
	return step, top
}

// yearTicks returns January 1 of every year strictly after start, up to now.
func yearTicks(start, now time.Time) []time.Time {
	var out []time.Time
	for y := start.Year() + 1; ; y++ {
		tm := time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
		if tm.After(now) {
			return out
		}
		out = append(out, tm)
	}
}

// grouped prints n with thousands separators.
func grouped(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

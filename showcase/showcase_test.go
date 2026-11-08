package main

import (
	"encoding/xml"
	"errors"
	"io"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestWrapText(t *testing.T) {
	t.Parallel()

	const width = 360.0
	long := strings.Repeat("Load CSV, TSV, JSON, and Parquet files into SQLite. ", 6)

	tests := []struct {
		name      string
		in        string
		wantLines int
		wantEllip bool
	}{
		{name: "empty", in: "   ", wantLines: 0},
		{name: "short", in: "Build Markdown from Go.", wantLines: 1},
		{name: "two lines", in: strings.Repeat("word ", 18), wantLines: 2},
		{name: "overflow", in: long, wantLines: 2, wantEllip: true},
		{name: "one huge word", in: strings.Repeat("x", 200), wantLines: 1, wantEllip: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lines := wrapText(tt.in, descSize, width, 2)
			if len(lines) != tt.wantLines {
				t.Fatalf("got %d lines %q, want %d", len(lines), lines, tt.wantLines)
			}
			for _, l := range lines {
				if w := textWidth(l, descSize, false); w > width {
					t.Errorf("line %q is %.1fpx, wider than %.0fpx", l, w, width)
				}
			}
			if got := len(lines) > 0 && strings.HasSuffix(lines[len(lines)-1], ellipsis); got != tt.wantEllip {
				t.Errorf("ellipsis = %v, want %v (%q)", got, tt.wantEllip, lines)
			}
		})
	}
}

func TestWrapTextKeepsEveryWordWhenItFits(t *testing.T) {
	t.Parallel()

	in := "Geographic math: haversine distance, bearing, geohash, polyline, Web Mercator, and bounding boxes."
	lines := wrapText(in, descSize, 360, 2)
	if got := strings.Join(lines, " "); got != in {
		t.Errorf("wrapped text lost words:\n got %q\nwant %q", got, in)
	}
}

func TestCompact(t *testing.T) {
	t.Parallel()

	for in, want := range map[int]string{0: "0", 999: "999", 1000: "1k", 1234: "1.2k", 9999: "10k", 12345: "12k"} {
		if got := compact(in); got != want {
			t.Errorf("compact(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderCardIsWellFormedXML(t *testing.T) {
	t.Parallel()

	d := CardData{
		Name:          `a<b>&"c"`,
		FullName:      `x/a<b>&"c"`,
		Description:   `Uses <script> & "quotes" in text`,
		Language:      "Go",
		LanguageColor: "#00ADD8",
		Stars:         12345,
		Forks:         7,
		Release:       "v1.0.0",
		Highlight:     true,
	}
	for _, theme := range themes {
		svg := renderCard(d, theme)
		dec := xml.NewDecoder(strings.NewReader(svg))
		for {
			_, err := dec.Token()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				t.Fatalf("%s card is not well-formed XML: %v\n%s", theme.Name, err, svg)
			}
		}
		if strings.Contains(svg, "<script>") {
			t.Errorf("%s card contains unescaped markup", theme.Name)
		}
		for _, want := range []string{"12k", "v1.0.0", "featured", theme.Bg} {
			if !strings.Contains(svg, want) {
				t.Errorf("%s card lacks %q", theme.Name, want)
			}
		}
	}
}

func TestRenderCardWithoutOptionalFields(t *testing.T) {
	t.Parallel()

	svg := renderCard(CardData{Name: "bare", FullName: "o/bare"}, themes[0])
	for _, unwanted := range []string{"featured", "<circle", iconTag} {
		if strings.Contains(svg, unwanted) {
			t.Errorf("card without data contains %q", unwanted)
		}
	}
}

func TestNewCardDataPrefersConfigDescription(t *testing.T) {
	t.Parallel()

	info := RepoInfo{Owner: "nao1215", Name: "gup", Description: "from github"}
	if got := newCardData(RepoSpec{Name: "gup", Description: "Run `go install`"}, info, "nao1215").Description; got != "Run go install" {
		t.Errorf("description = %q", got)
	}
	if got := newCardData(RepoSpec{Name: "gup"}, info, "nao1215").Description; got != "from github" {
		t.Errorf("fallback description = %q", got)
	}
	gozen := newCardData(RepoSpec{Name: "gup", Language: &Language{Name: "gozen", Color: "#ff5c8a"}}, RepoInfo{Owner: "nao1215", Name: "gup", Language: "Python", LanguageColor: "#3572A5"}, "nao1215")
	if gozen.Language != "gozen" || gozen.LanguageColor != "#ff5c8a" {
		t.Errorf("language override ignored: %q %q", gozen.Language, gozen.LanguageColor)
	}
	other := RepoInfo{Owner: "someone", Name: "tool"}
	if got := newCardData(RepoSpec{Name: "someone/tool"}, other, "nao1215").Name; got != "someone/tool" {
		t.Errorf("name of another owner's repository = %q", got)
	}
}

func TestReplaceBetweenMarkers(t *testing.T) {
	t.Parallel()

	doc := "intro\n\n" + markerStart + "\nold\n" + markerEnd + "\nfooter\n"
	got, err := replaceBetweenMarkers(doc, "new\n")
	if err != nil {
		t.Fatal(err)
	}
	want := "intro\n\n" + markerStart + "\n\nnew\n" + markerEnd + "\nfooter\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	again, err := replaceBetweenMarkers(got, "new\n")
	if err != nil || again != got {
		t.Errorf("second replacement changed the document: %q, %v", again, err)
	}

	for name, bad := range map[string]string{
		"no markers": "plain",
		"reversed":   markerEnd + markerStart,
		"duplicated": markerStart + markerEnd + markerStart + markerEnd,
	} {
		if _, err := replaceBetweenMarkers(bad, "x"); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestRenderShowcasePairsCards(t *testing.T) {
	t.Parallel()

	cfg := &Config{Owner: "o", Sections: []Section{
		{Summary: "Folded", Groups: []Group{{Title: "A", Repos: []RepoSpec{{Name: "a1"}, {Name: "a2"}, {Name: "a3"}}}}},
		{Groups: []Group{{Title: "B", Repos: []RepoSpec{{Name: "other/b1"}}}}},
	}}
	infos := map[string]RepoInfo{
		"a1": {Owner: "o", Name: "a1"}, "a2": {Owner: "o", Name: "a2"}, "a3": {Owner: "o", Name: "a3"},
		"other/b1": {Owner: "other", Name: "b1"},
	}
	out := renderShowcase(cfg, infos, "cards")

	for _, want := range []string{
		"<details>\n<summary>Folded</summary>",
		"</details>\n\n---\n\n### B",
		`srcset="cards/dark/o__a1.svg"`,
		`<img src="cards/light/other__b1.svg"`,
		"o__a1.svg\" width=\"400\" alt=\"o/a1\"></picture></a>\n<a",
		"o__a2.svg\" width=\"400\" alt=\"o/a2\"></picture></a>\n\n<a",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("showcase lacks %q:\n%s", want, out)
		}
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	ok := Config{Owner: "o", Sections: []Section{{Groups: []Group{{Title: "T", Repos: []RepoSpec{{Name: "a"}, {Name: "x/a"}}}}}}}
	if err := ok.validate(); err != nil {
		t.Errorf("valid config rejected: %v", err)
	}
	for name, cfg := range map[string]Config{
		"no owner":   {Sections: ok.Sections},
		"duplicate":  {Owner: "o", Sections: []Section{{Groups: []Group{{Title: "T", Repos: []RepoSpec{{Name: "a"}, {Name: "O/A"}}}}}}},
		"bad name":   {Owner: "o", Sections: []Section{{Groups: []Group{{Title: "T", Repos: []RepoSpec{{Name: "a/b/c"}}}}}}},
		"no title":   {Owner: "o", Sections: []Section{{Groups: []Group{{Repos: []RepoSpec{{Name: "a"}}}}}}},
		"bad color":  {Owner: "o", Sections: []Section{{Groups: []Group{{Title: "T", Repos: []RepoSpec{{Name: "a", Language: &Language{Name: "gozen", Color: "pink"}}}}}}}},
		"empty name": {Owner: "o", Sections: []Section{{Groups: []Group{{Title: "T", Repos: []RepoSpec{{Name: ""}}}}}}},
	} {
		if err := cfg.validate(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestShippedConfigIsValid(t *testing.T) {
	t.Parallel()

	if _, err := loadConfig("config.json"); err != nil {
		t.Fatal(err)
	}
}

func TestRenderStarGraph(t *testing.T) {
	t.Parallel()

	base := time.Date(2021, 3, 1, 12, 0, 0, 0, time.UTC)
	dates := func(n int, every time.Duration) []time.Time {
		var out []time.Time
		for i := range n {
			out = append(out, base.Add(time.Duration(i)*every))
		}
		return out
	}
	series := []StarSeries{
		{Name: "big", Dates: dates(1234, 36*time.Hour)},
		{Name: `a<b>&"c"`, Dates: dates(3, time.Hour)},
		{Name: "empty"},
		{Name: "small", Dates: dates(5, time.Hour)},
	}
	now := base.AddDate(5, 0, 0)
	for _, theme := range themes {
		for _, in := range [][]StarSeries{series, nil, series[2:3], series[1:2]} {
			svg := renderStarGraph(in, now, theme)
			dec := xml.NewDecoder(strings.NewReader(svg))
			for {
				_, err := dec.Token()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					t.Fatalf("%s graph is not well-formed XML: %v\n%s", theme.Name, err, svg)
				}
			}
			if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
				t.Errorf("%s graph with %d series has a bad coordinate", theme.Name, len(in))
			}
		}
	}
	svg := renderStarGraph(series, now, themes[0])
	for _, want := range []string{">1,234<", ">2022<", ">2026<", ">1.5k<", themes[0].Series[3]} {
		if !strings.Contains(svg, want) {
			t.Errorf("graph lacks %q", want)
		}
	}
}

func TestSpreadLabels(t *testing.T) {
	t.Parallel()

	got := spreadLabels([]float64{100, 100, 50, 101}, 40, 104, 14)
	sorted := append([]float64(nil), got...)
	sort.Float64s(sorted)
	for i := 1; i < len(sorted); i++ {
		if sorted[i]-sorted[i-1] < 14-1e-9 {
			t.Errorf("labels overlap: %v", got)
		}
	}
	for _, y := range got {
		if y < 40 || y > 104 {
			t.Errorf("label %v is outside the plot: %v", y, got)
		}
	}
	if got[2] != 50 {
		t.Errorf("a label with room moved: %v", got)
	}
}

func TestThemesHaveEnoughSeriesColors(t *testing.T) {
	t.Parallel()

	for _, theme := range themes {
		if len(theme.Series) != maxStarSeries {
			t.Errorf("%s theme has %d series colors, want %d", theme.Name, len(theme.Series), maxStarSeries)
		}
	}
}

func TestNiceScale(t *testing.T) {
	t.Parallel()

	for in, want := range map[int][2]int{1: {1, 2}, 7: {2, 8}, 1641: {500, 2000}, 2000: {500, 2500}} {
		step, top := niceScale(in)
		if step != want[0] || top != want[1] {
			t.Errorf("niceScale(%d) = %d, %d; want %v", in, step, top, want)
		}
	}
}

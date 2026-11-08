package main

import (
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"
)

const (
	cardWidth  = 400
	cardHeight = 120
	padX       = 20
	fontStack  = `-apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif`
	titleSize  = 15.0
	descSize   = 12.5
	metaSize   = 12.0
	// fallbackLanguageColor is used when GitHub reports no language color.
	fallbackLanguageColor = "#8b949e"
)

// Theme is the palette of one color scheme.
type Theme struct {
	Name      string
	Bg        string
	Border    string
	Title     string
	Text      string
	Muted     string
	Highlight string // border and badge color of highlighted cards
	// Series colors the lines of the star history graph, in config order.
	// Validated for colorblind separation against Bg as adjacent pairs.
	Series []string
}

var themes = []Theme{
	{Name: "light", Bg: "#ffffff", Border: "#d1d9e0", Title: "#0969da", Text: "#1f2328", Muted: "#59636e", Highlight: "#bf8700",
		Series: []string{"#2a78d6", "#eb6834", "#1baf7a", "#eda100", "#e87ba4", "#008300"}},
	{Name: "dark", Bg: "#151b23", Border: "#3d444d", Title: "#4493f8", Text: "#d1d7e0", Muted: "#9198a1", Highlight: "#d29922",
		Series: []string{"#3987e5", "#d95926", "#199e70", "#c98500", "#d55181", "#008300"}},
}

// Octicons (MIT License, https://github.com/primer/octicons), 16px.
const (
	iconRepo = "M2 2.5A2.5 2.5 0 0 1 4.5 0h8.75a.75.75 0 0 1 .75.75v12.5a.75.75 0 0 1-.75.75h-2.5a.75.75 0 0 1 0-1.5h1.75v-2h-8a1 1 0 0 0-.714 1.7.75.75 0 1 1-1.072 1.05A2.495 2.495 0 0 1 2 11.5Zm10.5-1h-8a1 1 0 0 0-1 1v6.708A2.486 2.486 0 0 1 4.5 9h8ZM5 12.25a.25.25 0 0 1 .25-.25h3.5a.25.25 0 0 1 .25.25v3.25a.25.25 0 0 1-.4.2l-1.45-1.087a.249.249 0 0 0-.3 0L5.4 15.7a.25.25 0 0 1-.4-.2Z"
	iconStar = "M8 .25a.75.75 0 0 1 .673.418l1.882 3.815 4.21.612a.75.75 0 0 1 .416 1.279l-3.046 2.97.719 4.192a.751.751 0 0 1-1.088.791L8 12.347l-3.766 1.98a.75.75 0 0 1-1.088-.79l.72-4.194L.818 6.374a.75.75 0 0 1 .416-1.28l4.21-.611L7.327.668A.75.75 0 0 1 8 .25Zm0 2.445L6.615 5.5a.75.75 0 0 1-.564.41l-3.097.45 2.24 2.184a.75.75 0 0 1 .216.664l-.528 3.084 2.769-1.456a.75.75 0 0 1 .698 0l2.77 1.456-.53-3.084a.75.75 0 0 1 .216-.664l2.24-2.183-3.096-.45a.75.75 0 0 1-.564-.41L8 2.694Z"
	iconFork = "M5 5.372v.878c0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75v-.878a2.25 2.25 0 1 1 1.5 0v.878a2.25 2.25 0 0 1-2.25 2.25h-1.5v2.128a2.251 2.251 0 1 1-1.5 0V8.5h-1.5A2.25 2.25 0 0 1 3.5 6.25v-.878a2.25 2.25 0 1 1 1.5 0ZM5 3.25a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Zm6.75.75a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Zm-3 8.75a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Z"
	iconTag  = "M1 7.775V2.75C1 1.784 1.784 1 2.75 1h5.025c.464 0 .91.184 1.238.513l6.25 6.25a1.75 1.75 0 0 1 0 2.474l-5.026 5.026a1.75 1.75 0 0 1-2.474 0l-6.25-6.25A1.752 1.752 0 0 1 1 7.775Zm1.5 0c0 .066.026.13.073.177l6.25 6.25a.25.25 0 0 0 .354 0l5.025-5.025a.25.25 0 0 0 0-.354l-6.25-6.25a.25.25 0 0 0-.177-.073H2.75a.25.25 0 0 0-.25.25ZM6 5a1 1 0 1 1 0 2 1 1 0 0 1 0-2Z"
)

// CardData is everything needed to draw one card.
type CardData struct {
	Name          string // shown as the title
	FullName      string // owner/name, used for the accessible label
	Description   string
	Language      string
	LanguageColor string
	Stars         int
	Forks         int
	Release       string
	Highlight     bool
}

func newCardData(spec RepoSpec, info RepoInfo, defaultOwner string) CardData {
	desc := spec.Description
	if desc == "" {
		desc = info.Description
	}
	name := info.Name
	if info.Owner != defaultOwner {
		name = info.Owner + "/" + info.Name
	}
	lang, langColor := info.Language, info.LanguageColor
	if spec.Language != nil {
		lang, langColor = spec.Language.Name, spec.Language.Color
	}
	return CardData{
		Name:          name,
		FullName:      info.Owner + "/" + info.Name,
		Description:   plainText(desc),
		Language:      lang,
		LanguageColor: langColor,
		Stars:         info.Stars,
		Forks:         info.Forks,
		Release:       info.LatestRelease,
		Highlight:     spec.Highlight,
	}
}

// plainText drops the Markdown code marks that the README descriptions use.
func plainText(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "`", "")), " ")
}

// renderCard draws the card as a standalone SVG document.
func renderCard(d CardData, t Theme) string {
	var b strings.Builder
	esc := html.EscapeString

	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`+"\n",
		cardWidth, cardHeight, cardWidth, cardHeight, esc(d.FullName))
	fmt.Fprintf(&b, "<title>%s</title>\n", esc(strings.TrimSpace(d.FullName+": "+d.Description)))
	fmt.Fprintf(&b, `<style>text{font-family:%s}</style>`+"\n", fontStack)

	border, borderWidth := t.Border, 1.0
	if d.Highlight {
		border, borderWidth = t.Highlight, 1.5
	}
	fmt.Fprintf(&b, `<rect x="0.75" y="0.75" width="%s" height="%s" rx="8" fill="%s" stroke="%s" stroke-width="%s"/>`+"\n",
		num(cardWidth-1.5), num(cardHeight-1.5), t.Bg, border, num(borderWidth))

	langColor := d.LanguageColor
	if langColor == "" {
		langColor = fallbackLanguageColor
	}
	// Title row: repository icon, name, and a badge for highlighted cards.
	b.WriteString(icon(iconRepo, padX, 17, 1, t.Muted))
	titleMax := float64(cardWidth - padX - padX - 22)
	if d.Highlight {
		titleMax -= 70
	}
	title := truncateBold(d.Name, titleSize, titleMax)
	fmt.Fprintf(&b, `<text x="%d" y="30" font-size="%s" font-weight="600" fill="%s">%s</text>`+"\n",
		padX+22, num(titleSize), t.Title, esc(title))
	if d.Highlight {
		x := cardWidth - padX - 62
		fmt.Fprintf(&b, `<rect x="%d" y="16" width="62" height="18" rx="9" fill="none" stroke="%s"/>`+"\n", x, t.Highlight)
		fmt.Fprintf(&b, `<text x="%d" y="29" font-size="11" font-weight="600" text-anchor="middle" fill="%s">featured</text>`+"\n", x+31, t.Highlight)
	}

	// Description, at most two lines.
	for i, line := range wrapText(d.Description, descSize, float64(cardWidth-2*padX), 2) {
		fmt.Fprintf(&b, `<text x="%d" y="%d" font-size="%s" fill="%s">%s</text>`+"\n",
			padX, 55+i*18, num(descSize), t.Text, esc(line))
	}

	// Footer: language, stars, forks, and the latest release on the right.
	const footerY = 101
	x := float64(padX)
	if d.Language != "" {
		fmt.Fprintf(&b, `<circle cx="%s" cy="%d" r="6" fill="%s"/>`+"\n", num(x+6), footerY-4, esc(langColor))
		x += 17
		fmt.Fprintf(&b, `<text x="%s" y="%d" font-size="%s" fill="%s">%s</text>`+"\n", num(x), footerY, num(metaSize), t.Muted, esc(d.Language))
		x += textWidth(d.Language, metaSize, false) + 18
	}
	for _, m := range []struct {
		path  string
		value int
	}{{iconStar, d.Stars}, {iconFork, d.Forks}} {
		b.WriteString(icon(m.path, x, footerY-12, 0.875, t.Muted))
		x += 18
		v := compact(m.value)
		fmt.Fprintf(&b, `<text x="%s" y="%d" font-size="%s" fill="%s">%s</text>`+"\n", num(x), footerY, num(metaSize), t.Muted, v)
		x += textWidth(v, metaSize, false) + 18
	}
	if d.Release != "" {
		space := float64(cardWidth-padX) - x - 18
		rel := truncate(d.Release, metaSize, space)
		if rel != ellipsis {
			w := textWidth(rel, metaSize, false)
			rx := float64(cardWidth-padX) - w
			b.WriteString(icon(iconTag, rx-18, footerY-12, 0.875, t.Muted))
			fmt.Fprintf(&b, `<text x="%s" y="%d" font-size="%s" fill="%s">%s</text>`+"\n", num(rx), footerY, num(metaSize), t.Muted, esc(rel))
		}
	}

	b.WriteString("</svg>\n")
	return b.String()
}

func icon(path string, x, y, scale float64, fill string) string {
	return fmt.Sprintf(`<path transform="translate(%s %s) scale(%s)" fill="%s" d="%s"/>`+"\n", num(x), num(y), num(scale), fill, path)
}

func truncateBold(s string, size, maxWidth float64) string {
	// The bold estimate is about 6% wider; shrink the budget instead of
	// duplicating truncate.
	return truncate(s, size, maxWidth/1.06)
}

// compact formats counts like GitHub does: 999, 1.2k, 12k.
func compact(n int) string {
	switch {
	case n < 1000:
		return strconv.Itoa(n)
	case n < 10000:
		s := strconv.FormatFloat(float64(n)/1000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "k"
	default:
		return strconv.Itoa(n/1000) + "k"
	}
}

// num prints a coordinate without trailing zeros.
func num(f float64) string {
	return strconv.FormatFloat(math.Round(f*100)/100, 'f', -1, 64)
}

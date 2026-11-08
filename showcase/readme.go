package main

import (
	"errors"
	"fmt"
	"html"
	"path"
	"strings"
)

const (
	markerStart = "<!-- showcase:start -->"
	markerEnd   = "<!-- showcase:end -->"
)

// cardFile is the file name of a repository's card, the same in every theme.
func cardFile(owner, name string) string {
	return owner + "__" + name + ".svg"
}

// renderShowcase builds the Markdown that goes between the markers.
// cardsDir is the card directory relative to the README.
func renderShowcase(cfg *Config, infos map[string]RepoInfo, cardsDir string) string {
	var b strings.Builder
	for i, s := range cfg.Sections {
		if i > 0 {
			b.WriteString("---\n\n")
		}
		if s.Summary != "" {
			fmt.Fprintf(&b, "<details>\n<summary>%s</summary>\n\n", html.EscapeString(s.Summary))
		}
		for _, g := range s.Groups {
			fmt.Fprintf(&b, "### %s\n\n", g.Title)
			for j, r := range g.Repos {
				info := infos[r.Name]
				b.WriteString(cardLink(info.Owner, info.Name, cardsDir))
				b.WriteString("\n")
				// Two cards per row; a blank line starts the next row.
				if j%2 == 1 || j == len(g.Repos)-1 {
					b.WriteString("\n")
				}
			}
		}
		if s.Summary != "" {
			b.WriteString("</details>\n\n")
		}
	}
	return b.String()
}

func cardLink(owner, name, cardsDir string) string {
	file := cardFile(owner, name)
	full := html.EscapeString(owner + "/" + name)
	return fmt.Sprintf(`<a href="https://github.com/%s"><picture><source media="(prefers-color-scheme: dark)" srcset="%s"><img src="%s" width="400" alt="%s"></picture></a>`,
		full, path.Join(cardsDir, "dark", file), path.Join(cardsDir, "light", file), full)
}

// replaceBetweenMarkers swaps the text between the start and end markers,
// keeping the markers themselves.
func replaceBetweenMarkers(doc, body string) (string, error) {
	start := strings.Index(doc, markerStart)
	end := strings.Index(doc, markerEnd)
	if start < 0 || end < 0 {
		return "", errors.New("README has no showcase markers")
	}
	if end < start {
		return "", errors.New("showcase end marker comes before the start marker")
	}
	if strings.Count(doc, markerStart) != 1 || strings.Count(doc, markerEnd) != 1 {
		return "", errors.New("README has more than one pair of showcase markers")
	}
	head := doc[:start+len(markerStart)]
	return head + "\n\n" + body + doc[end:], nil
}

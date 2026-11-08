// Command showcase draws a card for each repository listed in config.json and
// rewrites the showcase section of the profile README to show them.
//
// Run it from the showcase directory:
//
//	GITHUB_TOKEN=... go run .
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "showcase:", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "config.json", "repository list")
	readmePath := flag.String("readme", "../README.md", "README to rewrite")
	cardsDir := flag.String("cards", "cards", "directory to write cards into, relative to the README")
	flag.Parse()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token == "" {
		return errors.New("set GITHUB_TOKEN or GH_TOKEN")
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 30 * time.Second}
	infos, err := fetchRepos(ctx, client, token, cfg.Repos(), cfg.Owner)
	if err != nil {
		return err
	}

	readmeDir := filepath.Dir(*readmePath)
	outDir := filepath.Join(readmeDir, filepath.FromSlash(*cardsDir))
	if err := writeCards(cfg, infos, outDir); err != nil {
		return err
	}

	doc, err := os.ReadFile(*readmePath)
	if err != nil {
		return err
	}
	updated, err := replaceBetweenMarkers(string(doc), renderShowcase(cfg, infos, filepath.ToSlash(*cardsDir)))
	if err != nil {
		return err
	}
	for _, r := range cfg.Repos() {
		if infos[r.Name].Archived {
			fmt.Fprintf(os.Stderr, "showcase: warning: %s is archived\n", r.Name)
		}
	}
	return writeIfChanged(*readmePath, []byte(updated))
}

// writeCards writes every card in every theme and removes cards of
// repositories that are no longer listed.
func writeCards(cfg *Config, infos map[string]RepoInfo, outDir string) error {
	for _, t := range themes {
		dir := filepath.Join(outDir, t.Name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		keep := map[string]bool{}
		for _, r := range cfg.Repos() {
			info := infos[r.Name]
			file := cardFile(info.Owner, info.Name)
			keep[file] = true
			svg := renderCard(newCardData(r, info, cfg.Owner), t)
			if err := writeIfChanged(filepath.Join(dir, file), []byte(svg)); err != nil {
				return err
			}
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".svg") && !keep[e.Name()] {
				if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func writeIfChanged(path string, data []byte) error {
	if old, err := os.ReadFile(path); err == nil && string(old) == string(data) {
		return nil
	}
	return os.WriteFile(path, data, 0o644)
}

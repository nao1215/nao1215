package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

// fetchStarSeries returns, for each highlighted repository with at least
// cfg.StarGraphMinStars stars, in config order,
// when each of its stars was given, oldest first. Stars of users who later
// deleted their account are gone from the API, so a line can end a little
// below the count on the card.
func fetchStarSeries(ctx context.Context, client *http.Client, token string, cfg *Config, infos map[string]RepoInfo) ([]StarSeries, error) {
	var out []StarSeries
	for _, r := range cfg.Repos() {
		info := infos[r.Name]
		if !r.Highlight || info.Stars < cfg.StarGraphMinStars {
			continue
		}
		dates, err := fetchRepoStarDates(ctx, client, token, info.Owner, info.Name)
		if err != nil {
			return nil, fmt.Errorf("stargazers of %s/%s: %w", info.Owner, info.Name, err)
		}
		sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
		name := info.Name
		if info.Owner != cfg.Owner {
			name = info.Owner + "/" + info.Name
		}
		out = append(out, StarSeries{Name: name, Dates: dates})
	}
	return out, nil
}

// fetchRepoStarDates pages through the REST stargazers list. The GraphQL
// stargazers connection is not readable with the Actions GITHUB_TOKEN.
func fetchRepoStarDates(ctx context.Context, client *http.Client, token, owner, name string) ([]time.Time, error) {
	var out []time.Time
	for page := 1; ; page++ {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/stargazers?per_page=100&page=%d", owner, name, page)
		var stars []struct {
			StarredAt time.Time `json:"starred_at"`
		}
		if err := getWithRetry(ctx, client, token, url, &stars); err != nil {
			return nil, err
		}
		for _, st := range stars {
			out = append(out, st.StarredAt)
		}
		if len(stars) < 100 {
			return out, nil
		}
	}
}

func getWithRetry(ctx context.Context, client *http.Client, token, url string, out any) error {
	var lastErr error
	for attempt := range 3 {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 2 * time.Second):
			}
		}
		lastErr = get(ctx, client, token, url, out)
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

func get(ctx context.Context, client *http.Client, token, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.star+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "nao1215-showcase")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("github rest: %s: %s", res.Status, bytes.TrimSpace(data))
	}
	return json.Unmarshal(data, out)
}

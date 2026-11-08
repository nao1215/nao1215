package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
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

func fetchRepoStarDates(ctx context.Context, client *http.Client, token, owner, name string) ([]time.Time, error) {
	var out []time.Time
	cursor := ""
	for {
		after := ""
		if cursor != "" {
			after = fmt.Sprintf(", after: %q", cursor)
		}
		q := fmt.Sprintf(`query { repository(owner: %q, name: %q) { stargazers(first: 100%s) { edges { starredAt } pageInfo { hasNextPage endCursor } } } }`, owner, name, after)
		body, err := json.Marshal(map[string]string{"query": q})
		if err != nil {
			return nil, err
		}
		var resp stargazersResponse
		if err := postWithRetry(ctx, client, token, body, &resp); err != nil {
			return nil, err
		}
		if len(resp.Errors) > 0 {
			msgs := make([]string, 0, len(resp.Errors))
			for _, e := range resp.Errors {
				msgs = append(msgs, e.Message)
			}
			return nil, fmt.Errorf("github graphql: %s", strings.Join(msgs, "; "))
		}
		if resp.Data.Repository == nil {
			return nil, fmt.Errorf("repository not found")
		}
		sg := resp.Data.Repository.Stargazers
		for _, e := range sg.Edges {
			out = append(out, e.StarredAt)
		}
		if !sg.PageInfo.HasNextPage || sg.PageInfo.EndCursor == "" {
			return out, nil
		}
		cursor = sg.PageInfo.EndCursor
	}
}

type stargazersResponse struct {
	Data struct {
		Repository *struct {
			Stargazers struct {
				Edges []struct {
					StarredAt time.Time `json:"starredAt"`
				} `json:"edges"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"stargazers"`
		} `json:"repository"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

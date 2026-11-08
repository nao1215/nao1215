package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RepoInfo is what a card shows about a repository.
type RepoInfo struct {
	Owner         string
	Name          string
	Description   string
	Stars         int
	Forks         int
	Language      string
	LanguageColor string
	LatestRelease string
	Archived      bool
}

const graphqlEndpoint = "https://api.github.com/graphql"

// fetchRepos looks up every repository in a single GraphQL request.
func fetchRepos(ctx context.Context, client *http.Client, token string, specs []RepoSpec, defaultOwner string) (map[string]RepoInfo, error) {
	var q strings.Builder
	q.WriteString("query {")
	for i, s := range specs {
		owner, name := s.OwnerAndName(defaultOwner)
		fmt.Fprintf(&q, " r%d: repository(owner: %q, name: %q) { name owner { login } description stargazerCount forkCount isArchived primaryLanguage { name color } latestRelease { tagName } }", i, owner, name)
	}
	q.WriteString(" }")

	body, err := json.Marshal(map[string]string{"query": q.String()})
	if err != nil {
		return nil, err
	}

	var resp graphqlResponse
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

	out := make(map[string]RepoInfo, len(specs))
	for i, s := range specs {
		r, ok := resp.Data[fmt.Sprintf("r%d", i)]
		if !ok || r == nil {
			return nil, fmt.Errorf("repository %s not found", s.Name)
		}
		info := RepoInfo{
			Owner:       r.Owner.Login,
			Name:        r.Name,
			Description: r.Description,
			Stars:       r.StargazerCount,
			Forks:       r.ForkCount,
			Archived:    r.IsArchived,
		}
		if r.PrimaryLanguage != nil {
			info.Language = r.PrimaryLanguage.Name
			info.LanguageColor = r.PrimaryLanguage.Color
		}
		if r.LatestRelease != nil {
			info.LatestRelease = r.LatestRelease.TagName
		}
		out[s.Name] = info
	}
	return out, nil
}

type graphqlResponse struct {
	Data   map[string]*graphqlRepo `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type graphqlRepo struct {
	Name  string `json:"name"`
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`
	Description     string `json:"description"`
	StargazerCount  int    `json:"stargazerCount"`
	ForkCount       int    `json:"forkCount"`
	IsArchived      bool   `json:"isArchived"`
	PrimaryLanguage *struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	} `json:"primaryLanguage"`
	LatestRelease *struct {
		TagName string `json:"tagName"`
	} `json:"latestRelease"`
}

func postWithRetry(ctx context.Context, client *http.Client, token string, body []byte, out any) error {
	var lastErr error
	for attempt := range 3 {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 2 * time.Second):
			}
		}
		lastErr = post(ctx, client, token, body, out)
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

func post(ctx context.Context, client *http.Client, token string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlEndpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
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
		return fmt.Errorf("github graphql: %s: %s", res.Status, bytes.TrimSpace(data))
	}
	return json.Unmarshal(data, out)
}

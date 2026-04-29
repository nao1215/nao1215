#!/usr/bin/env bash

set -euo pipefail

readonly README_PATH="${1:-README.md}"

if [[ ! -f "$README_PATH" ]]; then
  echo "README file not found: $README_PATH" >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh command is required" >&2
  exit 1
fi

trim() {
  local value="$1"

  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

fetch_stars() {
  local owner="$1"
  local repo="$2"
  local attempt
  local stars

  for attempt in 1 2 3; do
    if stars="$(gh api "repos/${owner}/${repo}" --jq '.stargazers_count' 2>/dev/null)"; then
      printf '%s\n' "$stars"
      return 0
    fi

    sleep "$attempt"
  done

  echo "failed to fetch stars for ${owner}/${repo}" >&2
  return 1
}

tmp_file="$(mktemp)"
trap 'rm -f "$tmp_file"' EXIT
readonly REPO_LINK_PATTERN='^\[([^][]+)\]\(https://github\.com/([^/]+)/([^)/]+)\)$'

while IFS= read -r line; do
  if [[ "$line" == '| OSS | Language | Description |' ]]; then
    printf '| OSS | Language | Stars | Description |\n' >>"$tmp_file"
    continue
  fi

  if [[ "$line" == '| OSS | Language | Stars | Description |' ]]; then
    printf '%s\n' "$line" >>"$tmp_file"
    continue
  fi

  if [[ "$line" == '| :--- | :--- | :--- |' ]]; then
    printf '| :--- | :--- | ---: | :--- |\n' >>"$tmp_file"
    continue
  fi

  if [[ "$line" == '| :--- | :--- | ---: | :--- |' ]]; then
    printf '%s\n' "$line" >>"$tmp_file"
    continue
  fi

  if [[ "$line" != \|* ]]; then
    printf '%s\n' "$line" >>"$tmp_file"
    continue
  fi

  IFS='|' read -r _ raw_name raw_lang raw_col3 raw_col4 _ <<<"$line"

  name_cell="$(trim "$raw_name")"
  lang_cell="$(trim "$raw_lang")"
  col3_cell="$(trim "$raw_col3")"
  col4_cell="$(trim "$raw_col4")"

  if [[ ! $name_cell =~ $REPO_LINK_PATTERN ]]; then
    printf '%s\n' "$line" >>"$tmp_file"
    continue
  fi

  repo_name="${BASH_REMATCH[1]}"
  owner="${BASH_REMATCH[2]}"
  repo="${BASH_REMATCH[3]}"

  if [[ -n "$col4_cell" ]]; then
    description_cell="$col4_cell"
  else
    description_cell="$col3_cell"
  fi

  stars="$(fetch_stars "$owner" "$repo")"

  printf '| [%s](https://github.com/%s/%s) | %s | %s | %s |\n' \
    "$repo_name" \
    "$owner" \
    "$repo" \
    "$lang_cell" \
    "$stars" \
    "$description_cell" >>"$tmp_file"
done <"$README_PATH"

mv "$tmp_file" "$README_PATH"

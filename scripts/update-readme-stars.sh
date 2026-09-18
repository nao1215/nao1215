#!/usr/bin/env bash
# Rewrite the "Stars" column of every Markdown table in README.md whose first
# cell is a GitHub repository link. Other columns are left untouched, so tables
# may have any number of columns as long as the header contains "Stars".

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
    if stars="$(gh api "repos/${owner}/${repo}" --jq '.stargazers_count')"; then
      printf '%s\n' "$stars"
      return 0
    fi

    sleep "$attempt"
  done

  echo "failed to fetch stars for ${owner}/${repo}" >&2
  return 1
}

# Split a table row on unescaped "|" (a literal "\|" inside a cell is kept).
split_row() {
  local line="$1"
  local placeholder=$'\x1f'

  line="${line//\\|/$placeholder}"
  IFS='|' read -r -a CELLS <<<"$line"
  local i
  for i in "${!CELLS[@]}"; do
    CELLS[i]="${CELLS[i]//$placeholder/\\|}"
  done
}

tmp_file="$(mktemp)"
trap 'rm -f "$tmp_file"' EXIT
readonly REPO_LINK_PATTERN='^\[([^][]+)\]\(https://github\.com/([^/]+)/([^)/]+)\)$'

stars_index=-1
in_table=false

while IFS= read -r line; do
  if [[ "$line" != \|* ]]; then
    in_table=false
    stars_index=-1
    printf '%s\n' "$line" >>"$tmp_file"
    continue
  fi

  split_row "$line"

  # The first row of a table is its header: remember where "Stars" is.
  if [[ "$in_table" == false ]]; then
    in_table=true
    stars_index=-1
    for i in "${!CELLS[@]}"; do
      if [[ "$(trim "${CELLS[i]}")" == "Stars" ]]; then
        stars_index=$i
      fi
    done
    printf '%s\n' "$line" >>"$tmp_file"
    continue
  fi

  name_cell="$(trim "${CELLS[1]:-}")"
  if ((stars_index < 0)) || [[ ! $name_cell =~ $REPO_LINK_PATTERN ]]; then
    printf '%s\n' "$line" >>"$tmp_file"
    continue
  fi

  owner="${BASH_REMATCH[2]}"
  repo="${BASH_REMATCH[3]}"
  CELLS[stars_index]=" $(fetch_stars "$owner" "$repo") "

  out=""
  for ((i = 1; i < ${#CELLS[@]}; i++)); do
    out+="|${CELLS[i]}"
  done
  printf '%s|\n' "$out" >>"$tmp_file"
done <"$README_PATH"

mv "$tmp_file" "$README_PATH"

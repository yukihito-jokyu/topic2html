#!/usr/bin/env bash
set -euo pipefail

temporary_directory="$(mktemp -d "${TMPDIR:-/tmp}/topic2html-coverage.XXXXXX")"
trap 'rm -rf "$temporary_directory"' EXIT
merged_profile="$temporary_directory/coverage.out"
coverage_html="${COVERAGE_HTML:-coverage.html}"
printf 'mode: count\n' >"$merged_profile"

total_statements=0
covered_statements=0

while IFS= read -r package; do
	profile="$temporary_directory/$(printf '%s' "$package" | tr '/' '_').out"
	go test -coverpkg="$package" -coverprofile="$profile" "$package"
	tail -n +2 "$profile" >>"$merged_profile"

	coverage="$(awk '
		NR > 1 {
			total += $2
			if ($3 > 0) covered += $2
		}
		END { printf "%d %d", covered, total }
	' "$profile")"
	read -r covered statements <<<"$coverage"
	total_statements=$((total_statements + statements))
	covered_statements=$((covered_statements + covered))

	if (( statements == 0 )); then
		printf '%s: statementなし\n' "$package"
		continue
	fi

	percentage="$(awk -v covered="$covered" -v total="$statements" 'BEGIN { printf "%.1f", covered * 100 / total }')"
	printf '%s: %s%% (%d/%d statements)\n' "$package" "$percentage" "$covered" "$statements"
	if (( covered != statements )); then
		exit 1
	fi
done < <(go list -f '{{.ImportPath}}' ./...)

if (( total_statements == 0 )); then
	echo 'total: statementなし'
	exit 1
fi

percentage="$(awk -v covered="$covered_statements" -v total="$total_statements" 'BEGIN { printf "%.1f", covered * 100 / total }')"
printf 'total: %s%% (%d/%d statements)\n' "$percentage" "$covered_statements" "$total_statements"
if (( covered_statements != total_statements )); then
	exit 1
fi

go tool cover -html="$merged_profile" -o "$coverage_html"
printf 'coverage HTML: %s\n' "$coverage_html"

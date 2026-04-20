#!/usr/bin/env bash
set -euo pipefail

# Parse Go coverage.out and output coverage.json
# Outputs: line, function coverage stats per package and totals

COVERAGE_OUT="${1:-coverage.out}"
BRANCH="${GITHUB_REF_NAME:-local}"
COMMIT="${GITHUB_SHA:-unknown}"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)

if [[ ! -f "$COVERAGE_OUT" ]]; then
  echo "{\"error\": \"coverage.out not found\"}" >&2
  exit 1
fi

MODE=$(head -1 "$COVERAGE_OUT")
MODULE_PREFIX="github.com/poyrazk/thecloud/"

# Get total line coverage from go tool cover -func
TOTAL_LINE=$(go tool cover -func="$COVERAGE_OUT" 2>/dev/null | grep '^total:' | awk '{print $3}' | tr -d '%')

# Get total func coverage (average of all function coverage %)
# Parse each line: extract the last field (e.g. "20.0%"), strip %, add to sum
TOTAL_FUNC=$(go tool cover -func="$COVERAGE_OUT" 2>/dev/null | grep -v '^total:' | \
  sed 's/.*\t//' | tr -d '%' | \
  awk '{sum += $1; count++} END {if(count>0) printf "%.1f", sum/count; else print "0.0"}')

# Parse coverage.out for per-package line coverage
awk -v mod="$MODULE_PREFIX" '
  BEGIN { FS = "[ \t]+" }
  /^mode:/ { next }
  /^[ \t]*$/ { next }
  {
    file = $1
    sub(mod, "", file)
    # Strip /filename.go:line.col,line.col suffix
    if (match(file, "/[^/]+\\.go:[0-9]+\\.[0-9]+,[0-9]+\\.[0-9]+")) {
      file = substr(file, 1, RSTART - 1)
    }
    num_stmts = $2
    num_covered = $3

    if (!(file in pkg_cov)) {
      pkg_cov[file] = 0
      pkg_total[file] = 0
    }
    pkg_cov[file] += num_covered
    pkg_total[file] += num_stmts
  }
  END {
    for (pkg in pkg_cov) {
      total = pkg_total[pkg]
      cov = pkg_cov[pkg]
      if (total > 0) {
        pct = (cov * 100) / total
        printf "%s|%d|%d|%.2f\n", pkg, cov, total, pct
      }
    }
  }
' "$COVERAGE_OUT" | sort > /tmp/pkg_cov.txt

# Build JSON output
echo "{"
echo "  \"branch\": \"$BRANCH\","
echo "  \"commit\": \"$COMMIT\","
echo "  \"timestamp\": \"$TIMESTAMP\","
echo "  \"mode\": \"$MODE\","
echo "  \"total_line\": ${TOTAL_LINE:-0.0},"
echo "  \"total_func\": ${TOTAL_FUNC:-0.0},"
echo "  \"total_branch\": null,"
echo "  \"packages\": ["

first=true
while IFS='|' read -r pkg cov total pct; do
  if [[ "$first" == "true" ]]; then
    first=false
  else
    echo ","
  fi
  printf '    {"path": "./%s", "line": %s, "func": %s, "branch": null}' \
    "$pkg" "$pct" "${TOTAL_FUNC:-0.0}"
done < /tmp/pkg_cov.txt

echo ""
echo "  ]"
echo "}"

rm -f /tmp/pkg_cov.txt
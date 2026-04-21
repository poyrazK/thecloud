#!/usr/bin/env bash
set -euo pipefail

# Parse Go coverage.out and gobco.json for coverage stats
# Outputs: line, function, branch coverage stats per package and totals

COVERAGE_OUT="${1:-coverage.out}"
BRANCH="${GITHUB_REF_NAME:-local}"
COMMIT="${GITHUB_SHA:-unknown}"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
GOBCO_JSON="${2:-gobco.json}"

if [[ ! -f "$COVERAGE_OUT" ]]; then
  echo "{\"error\": \"coverage.out not found\"}" >&2
  exit 1
fi

MODE=$(head -1 "$COVERAGE_OUT")
MODULE_PREFIX="github.com/poyrazk/thecloud/"

# Get overall totals from go tool cover -func
TOTAL_LINE=$(go tool cover -func="$COVERAGE_OUT" 2>/dev/null | grep '^total:' | awk '{print $3}' | tr -d '%')
OVERALL_FUNC=$(go tool cover -func="$COVERAGE_OUT" 2>/dev/null | grep -v '^total:' | \
  sed 's/.*\t//' | tr -d '%' | \
  awk '{sum += $1; count++} END {if(count>0) printf "%.1f", sum/count; else print "0.0"}')

# Compute branch coverage from gobco JSON if available
OVERALL_BRANCH="null"
if [[ -f "$GOBCO_JSON" ]] && [[ -s "$GOBCO_JSON" ]]; then
  # Parse gobco JSON: compute coverage from TrueCount/FalseCount
  # Each condition: 2 outcomes (true, false)
  # Covered outcomes = count of conditions where both TrueCount>0 AND FalseCount>0 times 2
  #                   + count of conditions where only one is >0 times 1
  # Total outcomes = count of conditions * 2
  GOBCO_RESULT=$(cat "$GOBCO_JSON" | jq 'reduce .[] as $item (
    {"total": 0, "covered": 0};
    .total += 2 |
    if ($item.TrueCount > 0 and $item.FalseCount > 0) then
      .covered += 2
    elif ($item.TrueCount > 0 or $item.FalseCount > 0) then
      .covered += 1
    else
      .
    end
  ) | .coverage = (.covered / .total * 100 | floor) | {total, covered, coverage}' 2>/dev/null || echo "null")
  OVERALL_BRANCH=$(echo "$GOBCO_RESULT" | jq '.coverage' 2>/dev/null || echo "null")
fi

# Parse go tool cover output for per-package function coverage
go tool cover -func="$COVERAGE_OUT" 2>/dev/null | awk -v mod="$MODULE_PREFIX" '
  /^total:/ { next }
  {
    file = $1
    coverage = $NF
    gsub(/%/, "", coverage)
    sub(mod, "", file)
    goPos = index(file, ".go:")
    if (goPos > 0) pkg = substr(file, 1, goPos - 1)
    else pkg = file
    n = split(pkg, parts, "/")
    dir = ""
    for (i = 1; i < n; i++) {
      if (i > 1) dir = dir "/"
      dir = dir parts[i]
    }
    if (!(dir in f_sum)) { f_sum[dir] = 0; f_cnt[dir] = 0 }
    f_sum[dir] += coverage
    f_cnt[dir]++
  }
  END {
    for (d in f_sum) printf "%s|%.2f\n", d, f_sum[d]/f_cnt[d]
  }
' | sort > /tmp/pkg_func.txt

# Parse coverage.out for per-package line coverage
awk -v mod="$MODULE_PREFIX" '
  BEGIN { FS = "[ \t]+" }
  /^mode:/ { next }
  /^[ \t]*$/ { next }
  {
    file = $1
    sub(mod, "", file)
    if (match(file, "/[^/]+\\.go:[0-9]+\\.[0-9]+,[0-9]+\\.[0-9]+")) {
      file = substr(file, 1, RSTART - 1)
    }
    num_stmts = $2
    num_covered = $3

    if (!(file in l_cov)) {
      l_cov[file] = 0
      l_total[file] = 0
    }
    l_cov[file] += num_covered
    l_total[file] += num_stmts
  }
  END {
    for (pkg in l_cov) {
      total = l_total[pkg]
      cov = l_cov[pkg]
      if (total > 0) {
        pct = (cov * 100) / total
        printf "%s|%d|%d|%.2f\n", pkg, cov, total, pct
      }
    }
  }
' "$COVERAGE_OUT" | sort > /tmp/pkg_line.txt

# Build JSON output
echo "{"
echo "  \"branch\": \"$BRANCH\","
echo "  \"commit\": \"$COMMIT\","
echo "  \"timestamp\": \"$TIMESTAMP\","
echo "  \"mode\": \"$MODE\","
echo "  \"total_line\": ${TOTAL_LINE:-0.0},"
echo "  \"total_func\": ${OVERALL_FUNC:-0.0},"
echo "  \"total_branch\": ${OVERALL_BRANCH:-null},"
echo "  \"packages\": ["

first=true
while IFS='|' read -r pkg l_cov l_total l_pct; do
  func_pct=$(grep "^${pkg}|" /tmp/pkg_func.txt 2>/dev/null | cut -d'|' -f2 || echo "0.0")

  if [[ "$first" == "true" ]]; then
    first=false
  else
    echo ","
  fi
  printf '    {"path": "./%s", "line": %s, "func": %s, "branch": null}' \
    "$pkg" "$l_pct" "$func_pct"
done < /tmp/pkg_line.txt

echo ""
echo "  ]"
echo "}"

rm -f /tmp/pkg_line.txt /tmp/pkg_func.txt
#!/usr/bin/env bash
set -euo pipefail

# Parse Go coverage.out (mode: branch) and output coverage.json
# Outputs: line, function, and branch coverage stats per package and totals

COVERAGE_OUT="${1:-coverage.out}"
BRANCH="${GITHUB_REF_NAME:-local}"
COMMIT="${GITHUB_SHA:-unknown}"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Check if coverage.out exists
if [[ ! -f "$COVERAGE_OUT" ]]; then
  echo "{\"error\": \"coverage.out not found\"}" >&2
  exit 1
fi

# Extract mode from first line
MODE=$(head -1 "$COVERAGE_OUT")

# Helper function to calculate coverage percentage
calc_pct() {
  local covered=$1 total=$2
  if [[ "$total" -eq 0 ]]; then
    echo "0.0"
  else
    echo "scale=2; ($covered * 100) / $total" | bc
  fi
}

# Associative arrays for package-level stats
declare -A pkg_line_covered pkg_line_total pkg_branch_covered pkg_branch_total

total_line_covered=0 total_line_total=0
total_branch_covered=0 total_branch_total=0

# Module prefix to strip
MODULE_PREFIX="github.com/poyrazk/thecloud/"

# Parse coverage.out data lines
# Format (mode: set/count): file:start,end numStmts numCovered
# Format (mode: branch):     file:start,end numStmts numCovered branchTotal branchTaken
while IFS= read -r line; do
  # Skip mode line
  [[ "$line" =~ ^mode: ]] && continue
  [[ -z "$line" ]] && continue

  # Extract file path and data
  file=$(echo "$line" | awk '{print $1}')
  data=$(echo "$line" | awk '{print $2, $3, $4, $5}')

  # Extract package path: strip module prefix and filename
  # github.com/poyrazk/thecloud/internal/core/services/instance.go -> internal/core/services
  pkg="${file#$MODULE_PREFIX}"
  pkg="${pkg%/*}"

  # Parse fields based on mode
  if [[ "$MODE" == "mode: branch" ]]; then
    read -r num_stmts num_covered branch_total branch_taken <<< "$data"

    # Accumulate branch coverage
    if [[ -z "${pkg_branch_covered[$pkg]:-}" ]]; then
      pkg_branch_covered[$pkg]=0
      pkg_branch_total[$pkg]=0
    fi
    pkg_branch_covered[$pkg]=$((pkg_branch_covered[$pkg] + branch_taken))
    pkg_branch_total[$pkg]=$((pkg_branch_total[$pkg] + branch_total))
    total_branch_covered=$((total_branch_covered + branch_taken))
    total_branch_total=$((total_branch_total + branch_total))

    # Line coverage uses num_covered and num_stmts
    num_stmts_field=$num_covered  # in branch mode, fields shift
  else
    read -r num_stmts num_covered <<< "$data"
  fi

  # Accumulate line coverage
  if [[ -z "${pkg_line_covered[$pkg]:-}" ]]; then
    pkg_line_covered[$pkg]=0
    pkg_line_total[$pkg]=0
  fi
  pkg_line_covered[$pkg]=$((pkg_line_covered[$pkg] + num_covered))
  pkg_line_total[$pkg]=$((pkg_line_total[$pkg] + num_stmts))
  total_line_covered=$((total_line_covered + num_covered))
  total_line_total=$((total_line_total + num_stmts))

done < <(grep -v '^mode:' "$COVERAGE_OUT" | grep -v '^$')

# Function coverage from go tool cover -func
# Output format per function: "path/file.go:line.col  FuncName  coverage%"
FUNC_OUTPUT=$(go tool cover -func="$COVERAGE_OUT" 2>/dev/null)

# Compute total function coverage
total_func_pct=$(echo "$FUNC_OUTPUT" | grep -v '^total:' | awk '
  /^[ \t]*[0-9]/ {
    gsub(/%/, "", $NF)
    sum += $NF
    count++
  }
  END {
    if (count > 0) printf "%.1f", sum / count
    else print "0.0"
  }')

# Package-level function coverage map
declare -A pkg_func_pct
while IFS: read -r line; do
  [[ "$line" =~ ^total: ]] && continue
  [[ -z "$line" ]] && continue

  # Parse: "github.com/poyrazk/thecloud/path/file.go:line.col  FuncName  coverage%"
  file_path=$(echo "$line" | awk '{print $1}')
  coverage=$(echo "$line" | awk '{print $3}' | tr -d '%')

  pkg="${file_path#$MODULE_PREFIX}"
  pkg="${pkg%/*}"

  # Accumulate for average
  if [[ -z "${pkg_func_pct[$pkg]:-}" ]]; then
    pkg_func_pct[$pkg]=""
  fi
  # Store as space-separated list for later averaging
  pkg_func_list[$pkg]="${pkg_func_list[$pkg]:-} $coverage"
done <<< "$FUNC_OUTPUT"

# Calculate totals
total_line_pct=$(calc_pct $total_line_covered $total_line_total)
total_branch_pct=$(calc_pct $total_branch_covered $total_branch_total)

# Build JSON output using a temp file for reliability
TMPFILE=$(mktemp)
echo "{" > "$TMPFILE"
echo "  \"branch\": \"$BRANCH\"," >> "$TMPFILE"
echo "  \"commit\": \"$COMMIT\"," >> "$TMPFILE"
echo "  \"timestamp\": \"$TIMESTAMP\"," >> "$TMPFILE"
echo "  \"mode\": \"$MODE\"," >> "$TMPFILE"
echo "  \"total_line\": $total_line_pct," >> "$TMPFILE"
echo "  \"total_func\": $total_func_pct," >> "$TMPFILE"
echo "  \"total_branch\": $total_branch_pct," >> "$TMPFILE"
echo "  \"packages\": [" >> "$TMPFILE"

first=true
for pkg in "${!pkg_line_total[@]}"; do
  line_cov=$(calc_pct ${pkg_line_covered[$pkg]} ${pkg_line_total[$pkg]})

  # Branch coverage
  if [[ -n "${pkg_branch_total[$pkg]:-}" ]] && [[ "${pkg_branch_total[$pkg]}" -gt 0 ]]; then
    branch_cov=$(calc_pct ${pkg_branch_covered[$pkg]} ${pkg_branch_total[$pkg]})
  else
    branch_cov="null"
  fi

  # Function coverage average
  if [[ -n "${pkg_func_list[$pkg]:-}" ]]; then
    func_cov=$(echo "${pkg_func_list[$pkg]}" | awk '{for(i=1;i<=NF;i++) sum+=$i; if(NF>0) printf "%.1f", sum/NF; else print "0.0"}')
  else
    func_cov="0.0"
  fi

  if [[ "$first" == "true" ]]; then
    first=false
  else
    echo "," >> "$TMPFILE"
  fi

  printf '    {"path": "%s", "line": %s, "func": %s, "branch": %s}' \
    "./$pkg" "$line_cov" "$func_cov" "$branch_cov" >> "$TMPFILE"
done

echo "" >> "$TMPFILE"
echo "  ]" >> "$TMPFILE"
echo "}" >> "$TMPFILE"

cat "$TMPFILE"
rm -f "$TMPFILE"
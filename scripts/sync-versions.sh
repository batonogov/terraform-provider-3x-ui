#!/usr/bin/env bash
# sync-versions.sh — validate that CI matrices, README tables, and other
# surfaces stay in sync with compat-versions.json.
#
# Usage:
#   scripts/sync-versions.sh check    # fail if any surface is out of sync
#   scripts/sync-versions.sh fix      # update all surfaces in-place
#
# Exit codes: 0 = OK / fixed, 1 = drift detected (check mode).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPAT="$ROOT/compat-versions.json"
MODE="${1:-check}"

if ! command -v jq &>/dev/null; then
  echo "error: jq is required" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Extract versions from the single source of truth
# ---------------------------------------------------------------------------
VERSIONS_JSON=$(jq -c '[.versions[] | select(.supported == true) | .version]' "$COMPAT")
readarray -t VERSIONS < <(echo "$VERSIONS_JSON" | jq -r '.[]')
DEFAULT_VERSION=$(jq -r '.default_version' "$COMPAT")

# Build the sorted (newest-first) array for README tables.
readarray -t VERSIONS_DESC < <(echo "$VERSIONS_JSON" | jq -r 'reverse | .[]')

echo "Source: $COMPAT"
echo "Versions (${#VERSIONS[@]}): ${VERSIONS[*]}"
echo "Default: $DEFAULT_VERSION"
echo ""

DRIFT=0

# ---------------------------------------------------------------------------
# Helper: extract a JSON array of version strings from a CI matrix line
# like: version: ["v2.9.0", "v3.0.0"]
# ---------------------------------------------------------------------------
extract_matrix_versions() {
  local file="$1"
  # Grab all version: [...] lines and extract the versions
  grep -oP 'version:\s*\[\K[^\]]+' "$file" | tr -d '"' | tr -d "'" | tr ',' '\n' | sed 's/^ *//;s/ *$//' | sort
}

# ---------------------------------------------------------------------------
# 1. CI matrix (ci.yml)
# ---------------------------------------------------------------------------
check_ci() {
  local file="$ROOT/.github/workflows/ci.yml"
  local ci_versions
  ci_versions=$(extract_matrix_versions "$file")
  local source_versions
  source_versions=$(printf '%s\n' "${VERSIONS[@]}" | sort)

  if [ "$ci_versions" = "$source_versions" ]; then
    echo "ci.yml: OK"
  else
    echo "ci.yml: DRIFT DETECTED"
    diff <(echo "$ci_versions") <(echo "$source_versions") || true
    DRIFT=1

    if [ "$MODE" = "fix" ]; then
      local new_matrix
      new_matrix=$(echo "$VERSIONS_JSON" | jq -r '"[\"" + (join("\", \"")) + "\"]"')
      # Replace the version: [...] line in the matrix section
      sed -i.bak -E "s|version: \[.*\]|version: $new_matrix|" "$file" && rm -f "$file.bak"
      echo "  -> Fixed ci.yml"
    fi
  fi
}

# ---------------------------------------------------------------------------
# 2. Flake-tracking matrix and report for-loop
# ---------------------------------------------------------------------------
check_flake() {
  local file="$ROOT/.github/workflows/flake-tracking.yml"
  local flake_versions
  flake_versions=$(extract_matrix_versions "$file")
  local source_versions
  source_versions=$(printf '%s\n' "${VERSIONS[@]}" | sort)

  if [ "$flake_versions" = "$source_versions" ]; then
    echo "flake-tracking.yml matrix: OK"
  else
    echo "flake-tracking.yml matrix: DRIFT DETECTED"
    diff <(echo "$flake_versions") <(echo "$source_versions") || true
    DRIFT=1

    if [ "$MODE" = "fix" ]; then
      local new_matrix
      new_matrix=$(echo "$VERSIONS_JSON" | jq -r '"[\"" + (join("\", \"")) + "\"]"')
      sed -i.bak -E "s|version: \[.*\]|version: $new_matrix|" "$file" && rm -f "$file.bak"
      echo "  -> Fixed flake-tracking.yml matrix"
    fi
  fi

  # Check the for-loop in the report section
  local for_loop_versions
  for_loop_versions=$(grep -oP 'for v in \K.*(?=; do)' "$file" | tr ' ' '\n' | sort)
  if [ "$for_loop_versions" = "$source_versions" ]; then
    echo "flake-tracking.yml report for-loop: OK"
  else
    echo "flake-tracking.yml report for-loop: DRIFT DETECTED"
    diff <(echo "$for_loop_versions") <(echo "$source_versions") || true
    DRIFT=1

    if [ "$MODE" = "fix" ]; then
      local new_loop
      new_loop=$(printf '%s' "${VERSIONS[@]}")
      sed -i.bak -E "s|for v in .*; do|for v in $new_loop; do|" "$file" && rm -f "$file.bak"
      echo "  -> Fixed flake-tracking.yml report for-loop"
    fi
  fi
}

# ---------------------------------------------------------------------------
# 3. README compatibility tables (6 locales)
# ---------------------------------------------------------------------------
# Status label per locale
declare -A STATUS_LABEL=(
  ["README.md"]="Tested"
  ["README.ru_RU.md"]="Тестируется"
  ["README.es_ES.md"]="Probado"
  ["README.fa_IR.md"]="تست‌شده"
  ["README.ar_EG.md"]="تم اختباره"
  ["README.zh_CN.md"]="已测试"
)

# Header row per locale
declare -A HEADER_LINE=(
  ["README.md"]='| 3x-ui version | Status |'
  ["README.ru_RU.md"]='| Версия 3x-ui | Статус |'
  ["README.es_ES.md"]='| Versión de 3x-ui | Estado |'
  ["README.fa_IR.md"]='| نسخهٔ 3x-ui | وضعیت |'
  ["README.ar_EG.md"]='| إصدار 3x-ui | الحالة |'
  ["README.zh_CN.md"]='| 3x-ui 版本 | 状态 |'
)

check_readme() {
  local file="$1"
  local basename
  basename=$(basename "$file")
  local status="${STATUS_LABEL[$basename]}"
  local header="${HEADER_LINE[$basename]}"

  if [ -z "$status" ]; then
    echo "$basename: SKIP (unknown locale)"
    return
  fi

  # Build the expected table rows (newest first)
  local expected=""
  for v in "${VERSIONS_DESC[@]}"; do
    expected+="| $v | $status |"$'\n'
  done

  # Extract the current table rows between header and the next blank/section line
  local actual
  actual=$(awk "/^\\| ---/,,/^[^|]$/ {print}" "$file" | grep '^| v' | sort)

  local expected_sorted
  expected_sorted=$(echo "$expected" | grep '^| v' | sort)

  if [ "$actual" = "$expected_sorted" ]; then
    echo "$basename: OK"
  else
    echo "$basename: DRIFT DETECTED"
    diff <(echo "$actual") <(echo "$expected_sorted") || true
    DRIFT=1

    if [ "$MODE" = "fix" ]; then
      # Build the full table block: header + separator + rows
      local table_block
      table_block="${header}"$'\n'"| --- | --- |"$'\n'
      for v in "${VERSIONS_DESC[@]}"; do
        table_block+="| $v | $status |"$'\n'
      done
      # Remove trailing newline for sed insertion
      table_block="${table_block%$'\n'}"

      # Replace from the header line through the last version row
      # Use a temp file approach for BSD sed compatibility
      local tmp
      tmp=$(mktemp)
      awk -v header="$header" -v block="$table_block" '
        BEGIN { found=0 }
        $0 == header { print block; found=1; next }
        found && /^| ---/ { next }
        found && /^\| v/ { next }
        found && /^[^|]/ { found=0 }
        !found { print }
      ' "$file" > "$tmp" && mv "$tmp" "$file"
      echo "  -> Fixed $basename"
    fi
  fi
}

# ---------------------------------------------------------------------------
# 4. docker-compose.yaml default version
# ---------------------------------------------------------------------------
check_docker_compose() {
  local file="$ROOT/docker-compose.yaml"
  local current
  current=$(grep -oP 'THREEXUI_VERSION:-\K[^}]+' "$file")

  if [ "$current" = "$DEFAULT_VERSION" ]; then
    echo "docker-compose.yaml: OK"
  else
    echo "docker-compose.yaml: DRIFT (expected $DEFAULT_VERSION, got $current)"
    DRIFT=1

    if [ "$MODE" = "fix" ]; then
      sed -i.bak "s|THREEXUI_VERSION:-[^}]*|THREEXUI_VERSION:-$DEFAULT_VERSION|" "$file" && rm -f "$file.bak"
      echo "  -> Fixed docker-compose.yaml"
    fi
  fi
}

# ---------------------------------------------------------------------------
# 5. Taskfile.yml defaults
# ---------------------------------------------------------------------------
check_taskfile() {
  local file="$ROOT/Taskfile.yml"
  local current
  current=$(grep "default \"v" "$file" | head -1 | grep -oP 'default "\K[^"]+')

  if [ "$current" = "$DEFAULT_VERSION" ]; then
    echo "Taskfile.yml: OK"
  else
    echo "Taskfile.yml: DRIFT (expected $DEFAULT_VERSION, got $current)"
    DRIFT=1

    if [ "$MODE" = "fix" ]; then
      sed -i.bak "s|default \"v[^\"]*\"|default \"$DEFAULT_VERSION\"|g" "$file" && rm -f "$file.bak"
      echo "  -> Fixed Taskfile.yml"
    fi
  fi
}

# ---------------------------------------------------------------------------
# Run all checks
# ---------------------------------------------------------------------------
echo "=== Checking version consistency ==="
echo ""
check_ci
check_flake
check_readme "$ROOT/README.md"
check_readme "$ROOT/README.ru_RU.md"
check_readme "$ROOT/README.es_ES.md"
check_readme "$ROOT/README.fa_IR.md"
check_readme "$ROOT/README.ar_EG.md"
check_readme "$ROOT/README.zh_CN.md"
check_docker_compose
check_taskfile

echo ""
if [ "$DRIFT" -eq 1 ]; then
  if [ "$MODE" = "check" ]; then
    echo "FAIL: drift detected. Run 'scripts/sync-versions.sh fix' to update."
    exit 1
  else
    echo "FIXED: all surfaces updated from compat-versions.json"
  fi
else
  echo "PASS: all surfaces match compat-versions.json"
fi

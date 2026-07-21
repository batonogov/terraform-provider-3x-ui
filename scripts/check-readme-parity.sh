#!/usr/bin/env bash
# Verify that translated README tables expose the same machine-identifiable
# resource/data-source names and example links as README.md. Prose and row order
# are intentionally ignored so translations remain free to evolve.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REFERENCE="$ROOT/README.md"
LOCALES=(
  README.ru_RU.md
  README.fa_IR.md
  README.ar_EG.md
  README.zh_CN.md
  README.es_ES.md
  README.tr_TR.md
)

TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

extract_provider_symbols() {
  awk -F '|' '/^\| `threexui_[a-z_]+` / {
    value = $2
    gsub(/[ `]/, "", value)
    print value
  }' "$1" | sort -u
}

extract_example_links() {
  awk '/^\| \[.*\]\(examples\/[a-z0-9-]+\/\) \|/ { print }' "$1" \
    | grep -oE 'examples/[a-z0-9-]+/' \
    | sort -u
}

extract_example_directories() {
  local directory
  for directory in "$ROOT"/examples/*/; do
    if [ -f "$directory/main.tf" ]; then
      echo "examples/$(basename "$directory")/"
    fi
  done | sort -u
}

extract_provider_symbols "$REFERENCE" > "$TEMP_DIR/reference-symbols"
extract_example_links "$REFERENCE" > "$TEMP_DIR/reference-examples"
extract_example_directories > "$TEMP_DIR/example-directories"

drift=0
if ! diff -u "$TEMP_DIR/example-directories" "$TEMP_DIR/reference-examples"; then
  echo "README.md: examples table does not match examples/*/main.tf directories"
  drift=1
fi
for locale in "${LOCALES[@]}"; do
  file="$ROOT/$locale"
  extract_provider_symbols "$file" > "$TEMP_DIR/$locale-symbols"
  extract_example_links "$file" > "$TEMP_DIR/$locale-examples"

  if ! diff -u "$TEMP_DIR/reference-symbols" "$TEMP_DIR/$locale-symbols"; then
    echo "$locale: resource/data-source table drift"
    drift=1
  fi
  if ! diff -u "$TEMP_DIR/reference-examples" "$TEMP_DIR/$locale-examples"; then
    echo "$locale: examples table drift"
    drift=1
  fi
done

if [ "$drift" -ne 0 ]; then
  echo "FAIL: localized README tables differ from README.md"
  exit 1
fi

echo "PASS: localized README resource, data-source, and example tables match README.md"

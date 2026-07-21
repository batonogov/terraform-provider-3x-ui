#!/usr/bin/env bash
set -euo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
scratch_dir="$(mktemp -d "${TMPDIR:-/tmp}/threexui-example-validation.XXXXXX")"

cleanup() {
  rm -rf -- "$scratch_dir"
}
trap cleanup EXIT

provider_dir="$scratch_dir/provider"
cli_config="$scratch_dir/terraform.tfrc"
mkdir -p "$provider_dir"

GOCACHE="$scratch_dir/go-cache" \
  go build -o "$provider_dir/terraform-provider-threexui" "$repo_root"

printf '%s\n' \
  'provider_installation {' \
  '  dev_overrides {' \
  "    \"batonogov/threexui\" = \"$provider_dir\"" \
  '  }' \
  '  direct {}' \
  '}' >"$cli_config"

validate_example() {
  local example_dir="$1"
  local example_name
  local terraform_data_dir
  example_name="$(basename -- "$example_dir")"
  terraform_data_dir="$scratch_dir/terraform-data/$example_name"

  printf 'Validating %s\n' "$example_dir"
  TF_CLI_CONFIG_FILE="$cli_config" \
    TF_DATA_DIR="$terraform_data_dir" \
    terraform -chdir="$example_dir" get -no-color
  TF_CLI_CONFIG_FILE="$cli_config" \
    TF_DATA_DIR="$terraform_data_dir" \
    terraform -chdir="$example_dir" validate -no-color
}

validate_example "$repo_root/examples"

while IFS= read -r example_dir; do
  if find "$example_dir" -maxdepth 1 -type f -name '*.tf' -print -quit | grep -q .; then
    validate_example "$example_dir"
  fi
done < <(find "$repo_root/examples" -mindepth 1 -maxdepth 1 -type d | sort)

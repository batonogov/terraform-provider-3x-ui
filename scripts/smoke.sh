#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BIN_DIR="$ROOT_DIR/bin"
TERRAFORM_DIR="$ROOT_DIR/examples/resources/3xui_inbound"
TFRC_FILE="$ROOT_DIR/.smoke.tfrc"

: "${THREEXUI_BASE_URL:?Set THREEXUI_BASE_URL to run smoke tests}"
: "${THREEXUI_USERNAME:?Set THREEXUI_USERNAME to run smoke tests}"
: "${THREEXUI_PASSWORD:?Set THREEXUI_PASSWORD to run smoke tests}"
THREEXUI_TLS_SKIP_VERIFY="${THREEXUI_TLS_SKIP_VERIFY:-true}"

if ! command -v terraform >/dev/null 2>&1; then
  echo "terraform binary is required for smoke tests" >&2
  exit 1
fi

mkdir -p "$BIN_DIR"
go build -o "$BIN_DIR/terraform-provider-3x-ui" ./cmd/terraform-provider-3x-ui

cat >"$TFRC_FILE" <<EOF_TFRC
provider_installation {
  dev_overrides {
    "registry.terraform.io/batonogov/3x-ui" = "$BIN_DIR"
  }
  direct {}
}
EOF_TFRC

export TF_CLI_CONFIG_FILE="$TFRC_FILE"
export TF_VAR_threexui_base_url="$THREEXUI_BASE_URL"
export TF_VAR_threexui_username="$THREEXUI_USERNAME"
export TF_VAR_threexui_password="$THREEXUI_PASSWORD"
export TF_VAR_threexui_tls_skip_verify="$THREEXUI_TLS_SKIP_VERIFY"

pushd "$TERRAFORM_DIR" >/dev/null
trap 'terraform destroy -auto-approve || true; popd >/dev/null; rm -f "$TFRC_FILE"' EXIT

terraform init -upgrade
terraform apply -auto-approve
terraform destroy -auto-approve

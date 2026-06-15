#!/usr/bin/env bash
# scripts/fetch-xray-fixtures.sh
#
# (Re)generate the vendored GitHub fixtures served by the local github-cache
# proxy (ci/github-cache). The fixtures make the xray-version acceptance tests
# fully offline and deterministic — see issue #285 and
# ci/github-cache/README.md.
#
# What it produces under ci/github-cache/fixtures/:
#
#   api.github.com/repos/XTLS/Xray-core/releases.json
#       Minimal release list consumed by 3x-ui's GetXrayVersions. The panel
#       only parses the `tag_name` field, so we emit exactly that.
#
#   github.com/XTLS/Xray-core/releases/download/v<X.Y.Z>/Xray-linux-64.zip
#       The real release zips for each version. Only linux-64 is needed
#       (CI runs on ubuntu/amd64). These are served verbatim to the panel's
#       downloadXRay, so the zip structure must match the upstream release.
#
# The default VERSIONS cover every Xray-core release that any 3x-ui image in
# the compat matrix ships with today (v3.0.2 → 26.4.25, v3.1.0/v3.2.0 → 26.5.9,
# v3.2.5..v3.3.1 → 26.6.1) and therefore every InstallXray(currentVersion) and
# drift-alt the tests can request. Add a version here when a new 3x-ui image
# bundles a different Xray-core build.
#
# Usage:
#   scripts/fetch-xray-fixtures.sh                 # default versions
#   VERSIONS="26.6.1 26.7.0" scripts/fetch-xray-fixtures.sh
set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly FIXTURES="${REPO_ROOT}/ci/github-cache/fixtures"

VERSIONS="${VERSIONS:-26.6.1 26.5.9 26.4.25}"
# IMPORTANT: keep this newest-first to mirror the order GitHub's releases API
# returns (and that 3x-ui's GetXrayVersions forwards). TestAccXrayVersionDrift
# picks the first version != current as its drift target, so the order here
# decides whether the test exercises an upgrade or a downgrade — and some
# 3x-ui panels (e.g. v3.2.0) misbehave on an xray *downgrade* (restart window
# exceeds the 90s poll budget). Newest-first makes the cache faithful to the
# real API and keeps the drift test on the same upgrade path it had before
# the cache. Only linux-64 (amd64) is needed — CI runs on ubuntu-latest
# (amd64). Apple Silicon devs running the install/drift tests locally can add
# the arm64 slug:
#   ARCHES="linux-64 linux-arm64-v8a" scripts/fetch-xray-fixtures.sh
ARCHES="${ARCHES:-linux-64}"

releases_dir="${FIXTURES}/api.github.com/repos/XTLS/Xray-core"
zips_dir="${FIXTURES}/github.com/XTLS/Xray-core/releases/download"

mkdir -p "${releases_dir}" "${zips_dir}"

# --- releases.json ---------------------------------------------------------
# 3x-ui unmarshals into []struct{ TagName string `json:"tag_name"` } and then
# keeps versions where (major==26 && minor>4) || (major==26 && minor==4 &&
# patch>=25) || major>26. All default versions satisfy this.
{
  printf '['
  first=1
  for v in ${VERSIONS}; do
    tag="v${v#v}"
    if [ "${first}" -eq 0 ]; then printf ','; fi
    first=0
    printf '{"tag_name":"%s"}' "${tag}"
  done
  printf ']\n'
} > "${releases_dir}/releases.json"
echo "wrote ${releases_dir}/releases.json ($(wc -c < "${releases_dir}/releases.json") bytes)"

# --- zips ------------------------------------------------------------------
# Download each (version × arch) zip. Idempotent: skip files already present
# so the script is cheap to re-run during maintenance.
for v in ${VERSIONS}; do
  tag="v${v#v}"
  for arch in ${ARCHES}; do
    out="${zips_dir}/${tag}/Xray-${arch}.zip"
    if [ -s "${out}" ]; then
      echo "exists  ${out} ($(du -h "${out}" | cut -f1))"
      continue
    fi
    mkdir -p "$(dirname "${out}")"
    url="https://github.com/XTLS/Xray-core/releases/download/${tag}/Xray-${arch}.zip"
    echo "fetch   ${url}"
    if ! curl -fsSL --retry 3 -o "${out}" "${url}"; then
      echo "error: failed to download ${url}" >&2
      rm -f "${out}"
      exit 1
    fi
    echo "wrote   ${out} ($(du -h "${out}" | cut -f1))"
  done
done

echo
echo "fixtures ready under ${FIXTURES}"
echo "total: $(du -sh "${FIXTURES}" | cut -f1)"

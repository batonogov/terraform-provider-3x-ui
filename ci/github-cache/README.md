# github-cache

A local MITM forward proxy that serves **vendored GitHub fixtures** so the
3x-ui xray-version acceptance tests run fully offline and deterministically.
See [issue #285](https://github.com/batonogov/terraform-provider-3x-ui/issues/285).

## Why it exists

Three acceptance tests depend on the panel reaching GitHub at runtime:

| Test | Panel call | External URL |
| --- | --- | --- |
| `TestAccXrayVersion` | `GetXrayVersions` | `https://api.github.com/repos/XTLS/Xray-core/releases` |
| `TestAccXrayVersionResource` | `InstallXray(current)` | `https://github.com/.../Xray-linux-64.zip` |
| `TestAccXrayVersionDrift` | `InstallXray(current)` + `InstallXray(alt)` | same, ×2 |

Run once per panel version in the compat matrix, these were **the** source of
CI flakiness: the anonymous GitHub API rate limit (60 req/h/IP, shared across
all Actions runners) and transient download failures produced
`xray not running after 90s` errors that were unrelated to the provider code
under test.

## How it works

The panel's outbound HTTP client honours `HTTPS_PROXY` whenever `panel_outbound`
is unset (the default in tests). `internal/util/netproxy.NewHTTPClient` returns
a plain `*http.Client` for an empty proxy URL, so Go falls back to
`http.DefaultTransport` — i.e. `ProxyFromEnvironment`. Setting `HTTPS_PROXY` on
the `3xui` container therefore routes all of the panel's outbound GitHub traffic
through this proxy **with no upstream or provider changes**.

Because the targets are HTTPS, Go issues a `CONNECT` and then runs TLS
end-to-end over the tunnel. A plain file-serving proxy cannot see the request
without terminating TLS, so `github-cache` performs a **MITM**:

1. It presents a leaf certificate for `api.github.com` / `github.com` signed by
   a CA generated at startup (`entrypoint.sh`, written to the shared `github-cache-ca`
   volume).
2. The `3xui` container trusts that CA via `SSL_CERT_FILE` — its entrypoint is
   wrapped (see `docker-compose.yaml`) to append the CA to the system root
   bundle, so the panel keeps trusting real roots **and** our MITM certs.
3. Over the TLS session the proxy serves the vendored fixtures from disk
   (`fixtures/`).
4. `CONNECT` to any other host is **transparently tunnelled** (passthrough) so
   the panel's unrelated outbound traffic — public-IP detection
   (`ident.me`, `ipify`, …) — keeps working unchanged. The cache is invisible
   to everything except the two GitHub hosts.

Unknown GitHub **paths** return `404` (fail loud) so a test that drifts onto a
new GitHub URL is caught immediately.

## Files

```text
ci/github-cache/
├── Dockerfile         multi-stage: build Go proxy → tiny alpine + fixtures
├── entrypoint.sh      generate MITM CA (once, in the shared volume) → exec proxy
├── main.go            stdlib-only MITM proxy (~250 lines, no deps)
├── main_test.go       unit test for the (host, path) → fixture mapping
├── go.mod             standalone module (CI-only; not part of the provider)
└── fixtures/          vendored GitHub responses (committed, hermetic)
    ├── api.github.com/repos/XTLS/Xray-core/releases.json
    └── github.com/XTLS/Xray-core/releases/download/v<X.Y.Z>/Xray-linux-64.zip
```

## Fixtures

The fixtures cover exactly the Xray-core releases that any 3x-ui image in the
compat matrix ships with today:

| 3x-ui version | Bundled Xray |
| --- | --- |
| v3.0.2 | 26.4.25 |
| v3.1.0, v3.2.0 | 26.5.9 |
| v3.2.5 … v3.3.1 | 26.6.1 |

`InstallXray(currentVersion)` and the drift-alt therefore always resolve to a
zip the cache has. Only `linux-64` (amd64) zips are committed, because CI runs
on `ubuntu-latest` (amd64).

### Regenerating

```sh
scripts/fetch-xray-fixtures.sh                       # default versions, amd64
VERSIONS="26.6.1 26.7.0" scripts/fetch-xray-fixtures.sh
ARCHES="linux-64 linux-arm64-v8a" scripts/fetch-xray-fixtures.sh   # add arm64
```

Add a version to `VERSIONS` in the script (or the call above) when a new 3x-ui
image bundles a different Xray-core build.

### Local development on Apple Silicon

The panel runs inside an arm64 container on Apple Silicon and requests
`Xray-linux-arm64-v8a.zip`, which is **not** committed (CI is amd64). To run
the install/drift tests locally on arm64, fetch the arm64 zips first:

```sh
ARCHES="linux-64 linux-arm64-v8a" scripts/fetch-xray-fixtures.sh
docker compose up -d --build --wait
```

If a needed fixture is missing, the proxy returns HTTP 500 and logs the exact
path it could not open — check `docker compose logs github-cache`.

## Verifying hermeticity

The xray-version tests touch only `api.github.com` and `github.com`, both of
which the cache serves from fixtures. They have **zero outbound network**. A
scoped smoke test with `network_mode: none` on the `3xui` container (plus the
cache) confirms this: the GitHub paths still resolve from fixtures while the
tunnelled non-GitHub traffic simply fails (harmless for these tests).

## Disabling the cache

The cache is **always on** by design so local runs are as deterministic as CI
(no GitHub rate limits, no transient download failures). It does not break the
local workflow — `docker compose up` builds and starts it automatically.

To debug against the real GitHub instead (e.g. reproduce an upstream issue),
bypass it by temporarily commenting out the `HTTPS_PROXY`/`NO_PROXY` env, the
`github-cache` `depends_on`, and the `github-cache-ca` volume mount in
`docker-compose.yaml`, then `docker compose up --build`. The panel then reaches
GitHub directly, exactly as it did before this change.

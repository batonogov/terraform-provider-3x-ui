# syntax=docker/dockerfile:1
#
# github-cache: a local MITM forward proxy that serves vendored GitHub
# fixtures so the xray-version acceptance tests run fully offline. See
# ci/github-cache/README.md and issue #285.
#
# Multi-stage: build the Go proxy in a golang image, then ship only the static
# binary + fixtures in a tiny alpine image. The fixtures are committed to the
# repo (see scripts/fetch-xray-fixtures.sh), so this build needs NO network
# access — it is hermetic at build time AND at runtime.

# --- Stage 1: build the proxy ---------------------------------------------
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY *.go ./
# Static binary (CGO disabled) so the final image needs no libc.
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/github-cache .

# --- Stage 2: runtime image -----------------------------------------------
FROM alpine:3.20
# openssl: the entrypoint generates the MITM CA with it. wget: the compose
# healthcheck probes /healthz. ca-certificates is intentionally NOT needed —
# the proxy only acts as a TLS *server* (MITM) and never makes outbound TLS
# connections that would need to verify anyone.
RUN apk add --no-cache openssl wget

WORKDIR /app
COPY --from=builder /out/github-cache /app/github-cache
COPY entrypoint.sh /app/entrypoint.sh
COPY fixtures/ /app/fixtures/
RUN chmod +x /app/github-cache /app/entrypoint.sh

# /ca is the volume shared with the 3x-ui container: the CA generated here is
# what the panel trusts (via SSL_CERT_FILE) for the MITM'd api.github.com /
# github.com leaf certificates.
VOLUME ["/ca"]
EXPOSE 3128

ENTRYPOINT ["/app/entrypoint.sh"]

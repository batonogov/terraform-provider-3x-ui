#!/bin/sh
# ci/github-cache/entrypoint.sh
#
# Generates the MITM CA (once, persisted in the shared /ca volume) and then
# starts the github-cache proxy. The CA's public cert is trusted by the 3x-ui
# container via SSL_CERT_FILE (see docker-compose.yaml); its private key never
# leaves this volume, so it is not baked into an image layer.
set -e

CA_DIR="${CA_DIR:-/ca}"
CERT="${CA_DIR}/ca.crt"
KEY="${CA_DIR}/ca.key"

mkdir -p "$CA_DIR"

if [ ! -f "$CERT" ] || [ ! -f "$KEY" ]; then
    echo "github-cache: generating MITM CA in ${CA_DIR}"
    # genrsa emits PKCS#1; the proxy accepts both PKCS#1 and PKCS#8.
    openssl genrsa -out "$KEY" 2048 2>/dev/null
    openssl req -x509 -new -nodes -key "$KEY" -sha256 -days 3650 \
        -out "$CERT" \
        -subj "/CN=github-cache local CA/O=terraform-provider-threexui CI" 2>/dev/null
else
    echo "github-cache: reusing existing MITM CA from ${CA_DIR}"
fi

# Hand the CA to the proxy and exec (so signals reach the proxy as PID 1).
exec /app/github-cache -ca-cert "$CERT" -ca-key "$KEY" "$@"

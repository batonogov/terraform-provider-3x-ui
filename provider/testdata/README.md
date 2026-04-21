# Test Corpus

Fixture files used by `corpus_test.go` for round-trip testing of
flatten/expand conversion layers.

## Structure

| File | Description |
| --- | --- |
| `settings_*.json` | Inbound settings payloads (per-protocol) |
| `stream_settings_*.json` | Stream settings payloads (per-transport) |
| `sniffing_*.json` | Sniffing config payloads |
| `xray_template.json` | Full Xray config template (all sections) |
| `malformed_*` | Edge cases: truncated (.txt), wrong types, nulls, empty |

## Refreshing the Corpus

When the upstream 3x-ui panel changes its API payloads:

1. Start a local instance: `docker compose up -d && sleep 15`
2. Log in and create representative inbounds via the web UI
3. Fetch raw payloads:

   ```bash
   # Login
   TOKEN=$(curl -s -c - http://localhost:2053/login \
     -d 'username=admin&password=admin' | grep session)

   # Inbound list (extract settings/streamSettings/sniffing from each)
   curl -s -b "$TOKEN" http://localhost:2053/panel/api/inbounds/list | jq .

   # Xray template
   curl -s -b "$TOKEN" -X POST http://localhost:2053/panel/xray | jq .obj

   # Panel settings
   curl -s -b "$TOKEN" -X POST http://localhost:2053/panel/setting/all | jq .obj
   ```

4. Replace the fixture files with the new payloads
5. Remove zero-value / empty-string fields that the provider's build
   functions intentionally skip (these are not round-trip stable)
6. Run `go test ./provider/ -run TestCorpus -v` and fix any failures

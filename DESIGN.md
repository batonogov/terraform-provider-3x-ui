# Design: Expose xray Observatory/BurstObservatory (#336)

## Background

3x-ui v3.4.2+ ships `observatory` and `burstObservatory` as top-level xray
template keys (xray-core v26.6.27+). They are opaque `z.unknown()` in the
frontend schema (`xray.ts:33-34`), absent from the default config, and
auto-synced by the frontend with balancers via
`balancer-helpers.ts:syncObservatories`.

### xray-core Observatory config format

`observatory` is a top-level JSON object keyed by tag:

```json
{
  "observatory": {
    "obs1": {
      "subjectSelector": ["outbound1", "outbound2"],
      "probeURL": "https://www.google.com/generate_204",
      "probeInterval": "1m",
      "enableConcurrency": false
    }
  }
}
```

`burstObservatory` has the same key-by-tag structure but with a nested
`pingConfig` object:

```json
{
  "burstObservatory": {
    "burst1": {
      "subjectSelector": ["outbound1"],
      "pingConfig": {
        "destination": "https://www.google.com/generate_204",
        "interval": "1m",
        "connectTimeout": "5s",
        "timeout": "5s",
        "samples": 3
      }
    }
  }
}
```

### Panel balancer-driven sync

The frontend's `syncObservatories` creates/updates/deletes observatory
entries to match balancers that use `leastPing` or `leastLoad` strategy.
This sync runs **only in the browser** (Vue.js frontend) — it modifies the
xray config object before the user clicks "Save" in the panel UI. The
backend stores whatever JSON the API receives without running any sync.

The provider communicates with the panel through the backend API
(`GET/POST /panel/api/xray/*`), so the frontend sync does **not** run when
the provider updates the xray template. This means provider-managed
observatory entries persist across API-driven changes, exactly like
balancers, routing rules, DNS, and every other xray template section the
provider already manages.

The only conflict scenario: a user **manually saves via the panel UI**
after the provider has set observatories. The frontend sync may then
add/remove entries to reconcile with balancers. This is the same "last
writer wins" pattern that applies to all xray template sections and is
not a technical blocker.

## Decision: Resource (approach a — co-own)

**Approach (a): Co-own the entries.** The provider reads existing
observatories (including panel-synced ones) and allows the user to
declare desired ones via Terraform. The provider does not fight the panel
— on Read, it picks up whatever is in the xray template; on Create/Update,
it writes what the user declared.

### Why not (b) or (c)

- **(b) Read-only for panel-managed entries**: there is no reliable way to
  distinguish "panel-managed" from "user-managed" entries — the panel uses
  tag conventions but they are not documented or stable. Treating entries
  as read-only would require a heuristic that could break across panel
  versions.
- **(c) Data source**: too restrictive. Observatories are standard xray
  config, not panel-specific constructs. Users who want to manage them via
  Terraform (the primary use case for a Terraform provider) should be able
  to. A data source provides visibility but not management.

### Resource design

Single resource `threexui_xray_observatory` with two nested block types,
following the `xray_reverse` pattern (which has `bridge` and `portal`
blocks in one resource):

- `observatory {}` blocks — entries for the `observatory` JSON key
- `burst_observatory {}` blocks — entries for the `burstObservatory` JSON key

Each `observatory` block exposes: `tag`, `subject_selector`, `probe_url`,
`probe_interval`, `enable_concurrency`.

Each `burst_observatory` block exposes: `tag`, `subject_selector`, and a
nested `ping_config {}` block with `destination`, `interval`,
`connect_timeout`, `timeout`, `samples`, `sampling_count`, `lazy`.

The CRUD methods apply to two JSON paths (`observatory` and
`burstObservatory`) by calling the shared `xrayApplyTyped` / `xrayReadSection`
helpers once per path. This follows the same mutex-protected
read-modify-write pattern as all other xray resources.

### Documentation note

The resource docs will note that the panel UI's `syncObservatories` may
modify observatory entries when the user saves via the panel, and that the
provider will pick up those changes on the next `terraform refresh`.

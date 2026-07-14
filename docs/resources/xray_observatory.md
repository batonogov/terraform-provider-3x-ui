# threexui_xray_observatory

Manages xray Observatory and BurstObservatory configuration on the 3x-ui panel. The observatory produces latency data used by balancer strategies (`leastPing`, `leastLoad`).

## Example Usage

```terraform
resource "threexui_xray_observatory" "obs" {
  observatory {
    tag            = "obs_ping"
    subject_selector = ["inbound"]
    probe_url      = "https://www.google.com/generate_204"
    probe_interval = "1m"
  }
}
```

## Argument Reference

### observatory (Optional, Block List)

- `tag` (Required, String) — Observatory tag.
- `subject_selector` (Optional, List of String) — Inbound tags to observe.
- `probe_url` (Optional, String) — URL to probe for latency.
- `probe_interval` (Optional, String) — Probe interval (e.g. `1m`, `30s`).

### burst_observatory (Optional, Block List)

- `tag` (Required, String) — BurstObservatory tag.
- `subject_selector` (Optional, List of String) — Inbound tags to observe.
- `probe_url` (Optional, String) — URL to probe.
- `probe_interval` (Optional, String) — Probe interval.
- `probe_count` (Optional, Number) — Number of probes per burst.

terraform {
  required_providers {
    threexui = {
      source = "batonogov/threexui"
    }
  }
}

provider "threexui" {
  endpoint = "https://3x-ui.example.com"
  username = "admin"
  password = "admin"
}

# The threexui_xray_observatory resource manages both the "observatory" and
# "burstObservatory" sections of the xray-core template. These sections enable
# outbound latency monitoring so that balancers using strategies like
# "leastPing" or "leastLoad" can pick the best-performing outbound.

# Basic observatory: probes matching outbounds at a fixed interval.
resource "threexui_xray_observatory" "example" {
  observatory {
    tag                = "obs_default"
    subject_selector   = ["proxy-*"]
    probe_url          = "https://www.google.com/generate_204"
    probe_interval     = "1m"
    enable_concurrency = true
  }

  # Burst observatory: performs rapid burst-probes for finer-grained latency
  # data (xray-core v26.6.27+, 3x-ui v3.4.2+).
  burst_observatory {
    tag              = "burst_default"
    subject_selector = ["proxy-*"]

    ping_config {
      destination     = "https://www.cloudflare.com/cdn-cgi/trace"
      interval        = "1m"
      connect_timeout = "5s"
      timeout         = "10s"
      samples         = 3
      sampling_count  = 2
      lazy            = true
    }
  }
}

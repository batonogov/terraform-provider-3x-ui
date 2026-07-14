package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// testStringValidator is a helper that runs a StringValidator against a given
// string value and returns whether validation produced errors.
func testStringValidator(t *testing.T, v validator.String, value string) bool {
	t.Helper()
	req := validator.StringRequest{
		ConfigValue: types.StringValue(value),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	return resp.Diagnostics.HasError()
}

func testInt64Validator(t *testing.T, v validator.Int64, value int64) bool {
	t.Helper()
	req := validator.Int64Request{
		ConfigValue: types.Int64Value(value),
	}
	resp := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), req, resp)
	return resp.Diagnostics.HasError()
}

// ---------------------------------------------------------------------------
// Provider config validators
// ---------------------------------------------------------------------------

func TestEndpointValidators(t *testing.T) {
	v := endpointValidators()
	valid := []string{
		"http://localhost:2053",
		"https://panel.example.com",
		"http://192.168.1.1:8080",
	}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("endpointValidators should accept %q", val)
			}
		}
	}

	invalid := []string{
		"ftp://bad",
		"just-a-string",
		"localhost:2053",
		"",
	}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("endpointValidators should reject %q", val)
			}
		}
	}
}

func TestBasePathValidators(t *testing.T) {
	v := basePathValidators()
	valid := []string{"/", "/panel/", "/some/path"}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("basePathValidators should accept %q", val)
			}
		}
	}

	invalid := []string{"panel/", "no-slash", ""}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("basePathValidators should reject %q", val)
			}
		}
	}
}

func TestRequestTimeoutValidators(t *testing.T) {
	v := requestTimeoutValidators()
	valid := []string{"30s", "1m", "2m30s", "1h", "500ms"}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("requestTimeoutValidators should accept %q", val)
			}
		}
	}

	invalid := []string{"abc", "1x", "now"}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("requestTimeoutValidators should reject %q", val)
			}
		}
	}
}

func TestMaxRetriesValidators(t *testing.T) {
	v := maxRetriesValidators()
	valid := []int64{0, 1, 5, 10}
	for _, val := range valid {
		for _, vv := range v {
			if testInt64Validator(t, vv, val) {
				t.Errorf("maxRetriesValidators should accept %d", val)
			}
		}
	}

	invalid := []int64{-1, 11, 100}
	for _, val := range invalid {
		for _, vv := range v {
			if !testInt64Validator(t, vv, val) {
				t.Errorf("maxRetriesValidators should reject %d", val)
			}
		}
	}
}

func TestSubUpdatesValidators(t *testing.T) {
	v := subUpdatesValidators()
	valid := []int64{0, 1, 168, 525600}
	for _, val := range valid {
		for _, vv := range v {
			if testInt64Validator(t, vv, val) {
				t.Errorf("subUpdatesValidators should accept %d", val)
			}
		}
	}

	invalid := []int64{-1, 525601}
	for _, val := range invalid {
		for _, vv := range v {
			if !testInt64Validator(t, vv, val) {
				t.Errorf("subUpdatesValidators should reject %d", val)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Inbound resource validators
// ---------------------------------------------------------------------------

func TestPortValidators(t *testing.T) {
	v := portValidators()
	valid := []int64{1, 80, 443, 8080, 65535}
	for _, val := range valid {
		for _, vv := range v {
			if testInt64Validator(t, vv, val) {
				t.Errorf("portValidators should accept %d", val)
			}
		}
	}

	invalid := []int64{0, -1, 65536, 100000}
	for _, val := range invalid {
		for _, vv := range v {
			if !testInt64Validator(t, vv, val) {
				t.Errorf("portValidators should reject %d", val)
			}
		}
	}
}

func TestProtocolValidators(t *testing.T) {
	v := protocolValidators()
	valid := []string{
		"vless", "vmess", "trojan", "shadowsocks",
		"http", "socks", "mixed", "wireguard",
		"dokodemo-door", "tunnel", "tun",
		"hysteria", "hysteria2",
	}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("protocolValidators should accept %q", val)
			}
		}
	}

	invalid := []string{"invalid", "VLESS", "tcp", "udp", ""}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("protocolValidators should reject %q", val)
			}
		}
	}
}

func TestTrafficResetValidators(t *testing.T) {
	v := trafficResetValidators()
	valid := []string{"never", "hourly", "daily", "weekly", "monthly"}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("trafficResetValidators should accept %q", val)
			}
		}
	}

	invalid := []string{"day", "week", "month", "year", "NEVER", ""}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("trafficResetValidators should reject %q", val)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Stream settings validators
// ---------------------------------------------------------------------------

func TestNetworkValidators(t *testing.T) {
	v := networkValidators()
	valid := []string{"tcp", "ws", "grpc", "httpupgrade", "xhttp", "kcp", "hysteria"}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("networkValidators should accept %q", val)
			}
		}
	}

	invalid := []string{"quic", "TCP", "udp", ""}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("networkValidators should reject %q", val)
			}
		}
	}
}

func TestSecurityValidators(t *testing.T) {
	v := securityValidators()
	valid := []string{"none", "tls", "reality"}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("securityValidators should accept %q", val)
			}
		}
	}

	invalid := []string{"ssl", "TLS", " Reality ", ""}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("securityValidators should reject %q", val)
			}
		}
	}
}

func TestTCPHeaderTypeValidators(t *testing.T) {
	v := tcpHeaderTypeValidators()
	valid := []string{"none", "http"}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("tcpHeaderTypeValidators should accept %q", val)
			}
		}
	}

	invalid := []string{"srtp", "HTTP", ""}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("tcpHeaderTypeValidators should reject %q", val)
			}
		}
	}
}

func TestKCPHeaderTypeValidators(t *testing.T) {
	v := kcpHeaderTypeValidators()
	valid := []string{"none", "srtp", "utp", "wechat-video", "dtls", "wireguard"}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("kcpHeaderTypeValidators should accept %q", val)
			}
		}
	}

	invalid := []string{"http", "SRTP", ""}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("kcpHeaderTypeValidators should reject %q", val)
			}
		}
	}
}

func TestXHTTPModeValidators(t *testing.T) {
	v := xhttpModeValidators()
	valid := []string{"auto", "packet-up", "stream-up", "stream-one"}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("xhttpModeValidators should accept %q", val)
			}
		}
	}

	invalid := []string{"packet", "stream", ""}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("xhttpModeValidators should reject %q", val)
			}
		}
	}
}

func TestTproxyValidators(t *testing.T) {
	v := tproxyValidators()
	valid := []string{"off", "redirect", "tproxy"}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("tproxyValidators should accept %q", val)
			}
		}
	}

	invalid := []string{"on", "OFF", ""}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("tproxyValidators should reject %q", val)
			}
		}
	}
}

func TestDomainStrategyValidators(t *testing.T) {
	v := domainStrategyValidators()
	valid := []string{"AsIs", "UseIP", "UseIPv6v4", "UseIPv6", "UseIPv4v6", "UseIPv4",
		"ForceIP", "ForceIPv6v4", "ForceIPv6", "ForceIPv4v6", "ForceIPv4"}
	for _, val := range valid {
		for _, vv := range v {
			if testStringValidator(t, vv, val) {
				t.Errorf("domainStrategyValidators should accept %q", val)
			}
		}
	}

	invalid := []string{"asis", "IPIfNonMatch", "IPOnDemand", ""}
	for _, val := range invalid {
		for _, vv := range v {
			if !testStringValidator(t, vv, val) {
				t.Errorf("domainStrategyValidators should reject %q", val)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Duration validator (null/unknown skip)
// ---------------------------------------------------------------------------

func TestDurationValidator_SkipsNullAndUnknown(t *testing.T) {
	v := durationValidator{}

	// Null value should not produce errors
	req := validator.StringRequest{
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Error("durationValidator should skip null values")
	}
}

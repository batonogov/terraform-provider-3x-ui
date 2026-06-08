package provider

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// ---------------------------------------------------------------------------
// Provider config validators
// ---------------------------------------------------------------------------

// endpointValidators validates the provider endpoint URL format.
func endpointValidators() []validator.String {
	return []validator.String{
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^https?://.+`),
			"must be a valid HTTP or HTTPS URL (e.g. http://localhost:2053)",
		),
	}
}

// basePathValidators validates the provider base_path.
func basePathValidators() []validator.String {
	return []validator.String{
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^/.*`),
			"must start with '/'",
		),
	}
}

// requestTimeoutValidators validates the provider request_timeout duration string.
func requestTimeoutValidators() []validator.String {
	return []validator.String{
		durationValidator{},
	}
}

// maxRetriesValidators validates the provider max_retries integer range.
func maxRetriesValidators() []validator.Int64 {
	return []validator.Int64{
		int64validator.Between(0, int64(maxAllowedRetries)),
	}
}

// ---------------------------------------------------------------------------
// Inbound resource validators
// ---------------------------------------------------------------------------

// portValidators validates inbound port range (1-65535).
func portValidators() []validator.Int64 {
	return []validator.Int64{
		int64validator.Between(1, 65535),
	}
}

// protocolValidators validates inbound protocol enum values.
func protocolValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf(
			"vless", "vmess", "trojan", "shadowsocks",
			"http", "socks", "mixed", "wireguard",
			"dokodemo-door", "tunnel", "tun",
			"hysteria", "hysteria2",
		),
	}
}

// trafficResetValidators validates inbound traffic_reset enum values.
func trafficResetValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf("never", "hourly", "daily", "weekly", "monthly"),
	}
}

// ---------------------------------------------------------------------------
// Stream settings validators
// ---------------------------------------------------------------------------

// networkValidators validates stream network enum values.
func networkValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf("tcp", "ws", "grpc", "httpupgrade", "xhttp", "kcp", "hysteria"),
	}
}

// securityValidators validates stream security enum values.
func securityValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf("none", "tls", "reality"),
	}
}

// tcpHeaderTypeValidators validates TCP header_type enum values.
func tcpHeaderTypeValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf("none", "http"),
	}
}

// kcpHeaderTypeValidators validates KCP header_type enum values.
func kcpHeaderTypeValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf("none", "srtp", "utp", "wechat-video", "dtls", "wireguard"),
	}
}

// xhttpModeValidators validates XHTTP mode enum values.
func xhttpModeValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf("auto", "packet-up", "stream-up", "stream-one"),
	}
}

// tproxyValidators validates sockopt tproxy enum values.
func tproxyValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf("off", "redirect", "tproxy"),
	}
}

// domainStrategyValidators validates sockopt domain_strategy enum values.
func domainStrategyValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf(
			"AsIs", "UseIP", "UseIPv6v4", "UseIPv6", "UseIPv4v6", "UseIPv4",
			"ForceIP", "ForceIPv6v4", "ForceIPv6", "ForceIPv4v6", "ForceIPv4",
		),
	}
}

// ---------------------------------------------------------------------------
// Custom duration validator
// ---------------------------------------------------------------------------

// durationValidator validates that a string can be parsed as a Go duration.
type durationValidator struct{}

func (durationValidator) Description(_ context.Context) string {
	return "must be a valid Go duration string (e.g. 30s, 1m, 2m30s)"
}

func (durationValidator) MarkdownDescription(_ context.Context) string {
	return "must be a valid Go duration string (e.g. `30s`, `1m`, `2m30s`)"
}

func (durationValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := time.ParseDuration(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid duration",
			err.Error(),
		)
	}
}

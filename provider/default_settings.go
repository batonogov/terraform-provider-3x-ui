package provider

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

func applyDefaultInboundSettings(inbound *Inbound) error {
	if inbound == nil {
		return nil
	}

	raw := strings.TrimSpace(inbound.Settings)
	if raw == "" {
		return setDefaultSettings(inbound)
	}

	settings, err := ParseJSONField(raw)
	if err != nil {
		return err
	}
	if len(settings) == 0 {
		return setDefaultSettings(inbound)
	}

	updated := false
	if inbound.Protocol == "vless" {
		if _, ok := settings["decryption"]; !ok {
			settings["decryption"] = "none"
			updated = true
		}
	}

	if protocolUsesClients(inbound.Protocol) {
		if _, ok := settings["clients"]; !ok {
			settings["clients"] = []any{}
			updated = true
		}
	}

	if !updated {
		return nil
	}

	encoded, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	inbound.Settings = string(encoded)
	return nil
}

func setDefaultSettings(inbound *Inbound) error {
	settings, err := defaultSettingsForProtocol(inbound.Protocol)
	if err != nil {
		return err
	}
	if settings == nil {
		inbound.Settings = "{}"
		return nil
	}
	encoded, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	inbound.Settings = string(encoded)
	return nil
}

func protocolUsesClients(protocol string) bool {
	switch protocol {
	case "vmess", "vless", "trojan", "shadowsocks":
		return true
	default:
		return false
	}
}

func defaultSettingsForProtocol(protocol string) (map[string]any, error) {
	switch protocol {
	case "vless":
		client, err := defaultVlessClient()
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"clients":    []any{client},
			"decryption": "none",
			"encryption": "none",
		}, nil
	case "vmess":
		client, err := defaultVmessClient()
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"clients": []any{client},
		}, nil
	case "trojan":
		client, err := defaultTrojanClient()
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"clients": []any{client},
		}, nil
	case "shadowsocks":
		settings, err := defaultShadowsocksSettings()
		if err != nil {
			return nil, err
		}
		return settings, nil
	default:
		return nil, nil
	}
}

func defaultVlessClient() (map[string]any, error) {
	id, err := newUUID()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":         id,
		"flow":       "",
		"email":      randomLowerAndNum(8),
		"limitIp":    0,
		"totalGB":    0,
		"expiryTime": 0,
		"enable":     false,
		"tgId":       "",
		"subId":      randomLowerAndNum(16),
		"comment":    "",
		"reset":      0,
	}, nil
}

func defaultVmessClient() (map[string]any, error) {
	id, err := newUUID()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":         id,
		"security":   "auto",
		"email":      randomLowerAndNum(8),
		"limitIp":    0,
		"totalGB":    0,
		"expiryTime": 0,
		"enable":     false,
		"tgId":       "",
		"subId":      randomLowerAndNum(16),
		"comment":    "",
		"reset":      0,
	}, nil
}

func defaultTrojanClient() (map[string]any, error) {
	return map[string]any{
		"password":   randomSeq(10, true),
		"email":      randomLowerAndNum(8),
		"limitIp":    0,
		"totalGB":    0,
		"expiryTime": 0,
		"enable":     false,
		"tgId":       "",
		"subId":      randomLowerAndNum(16),
		"comment":    "",
		"reset":      0,
	}, nil
}

func defaultShadowsocksSettings() (map[string]any, error) {
	method := "2022-blake3-aes-256-gcm"
	serverPassword, err := randomShadowsocksPassword(method)
	if err != nil {
		return nil, err
	}
	clientPassword, err := randomShadowsocksPassword(method)
	if err != nil {
		return nil, err
	}
	client := map[string]any{
		"method":     "",
		"password":   clientPassword,
		"email":      randomLowerAndNum(8),
		"limitIp":    0,
		"totalGB":    0,
		"expiryTime": 0,
		"enable":     false,
		"tgId":       "",
		"subId":      randomLowerAndNum(16),
		"comment":    "",
		"reset":      0,
	}
	return map[string]any{
		"method":   method,
		"password": serverPassword,
		"network":  "tcp,udp",
		"clients":  []any{client},
		"ivCheck":  false,
	}, nil
}

func randomLowerAndNum(length int) string {
	return randomSeq(length, false)
}

func randomSeq(length int, includeUpper bool) string {
	const digits = "0123456789"
	const lower = "abcdefghijklmnopqrstuvwxyz"
	const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	seq := digits + lower
	if includeUpper {
		seq += upper
	}
	if length <= 0 {
		return ""
	}
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	out := make([]byte, length)
	for i := range buf {
		out[i] = seq[int(buf[i])%len(seq)]
	}
	return string(out)
}

func randomShadowsocksPassword(method string) (string, error) {
	length := 32
	if method == "2022-blake3-aes-128-gcm" {
		length = 16
	}
	if length <= 0 {
		return "", fmt.Errorf("invalid shadowsocks password length")
	}
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

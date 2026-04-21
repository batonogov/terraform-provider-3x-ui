package provider

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// repoRoot returns the absolute path of the repository root.
// It walks up from the current working directory looking for go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get cwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot find repo root (go.mod not found)")
		}
		dir = parent
	}
}

// findSnapshotDirs returns 3x-ui-* directories found in the given root.
func findSnapshotDirs(root string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "3x-ui-") {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Strings(dirs)
	return dirs
}

// latestSnapshotDir returns the path to the latest 3x-ui-<version> directory.
// It checks the THREEXUI_SNAPSHOT_DIR env var first, then the repo root.
// Returns "" if no snapshots are found.
func latestSnapshotDir(t *testing.T) string {
	t.Helper()

	// Allow explicit override via environment variable.
	if envDir := os.Getenv("THREEXUI_SNAPSHOT_DIR"); envDir != "" {
		if info, err := os.Stat(envDir); err == nil && info.IsDir() {
			return envDir
		}
	}

	root := repoRoot(t)
	dirs := findSnapshotDirs(root)
	if len(dirs) > 0 {
		return filepath.Join(root, dirs[len(dirs)-1])
	}
	return ""
}

// skipIfNoSnapshot skips the test if no 3x-ui snapshot is available.
func skipIfNoSnapshot(t *testing.T, dir string) {
	t.Helper()
	if dir == "" {
		t.Skip("no 3x-ui source snapshots found; skipping drift check")
	}
}

// --------------------------------------------------------------------------
// Test: upstream inbound protocols vs provider protocol mappings
// --------------------------------------------------------------------------

// TestDriftInboundProtocols_GoModel checks that every protocol constant in
// database/model/model.go has a corresponding handler in the provider's
// expandSettingsFromModel / flattenSettingsToModel.
func TestDriftInboundProtocols_GoModel(t *testing.T) {
	dir := latestSnapshotDir(t)
	skipIfNoSnapshot(t, dir)

	modelPath := filepath.Join(dir, "database", "model", "model.go")
	data, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("cannot read model.go: %v", err)
	}

	// Extract protocol constants: Protocol = "xxx"
	re := regexp.MustCompile(`Protocol\s*=\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		t.Fatal("no protocol constants found in model.go")
	}

	upstream := make(map[string]bool)
	for _, m := range matches {
		upstream[m[1]] = true
	}

	// Provider handles these protocols in expandSettingsFromModel / flattenSettingsToModel.
	// "dokodemo-door" is alias for "tunnel" in the provider; both share the same handler.
	// "vmess" and "mixed" have no per-protocol settings block (handled via default).
	// "socks" is not in 3x-ui model.go (it's an Xray protocol available via UI).
	providerHandled := map[string]bool{
		"vless":       true,
		"vmess":       true,
		"trojan":      true,
		"shadowsocks": true,
		"http":        true,
		"wireguard":   true,
		"tunnel":      true, // = dokodemo-door
		"mixed":       true,
		"hysteria":    true,
	}

	var missing []string
	for proto := range upstream {
		if !providerHandled[proto] {
			missing = append(missing, proto)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("upstream model.go has protocols not handled by provider: %v", missing)
	}
}

// TestDriftInboundProtocols_JS checks that every protocol in the JavaScript
// Protocols object (web/assets/js/model/inbound.js) is handled by the provider.
func TestDriftInboundProtocols_JS(t *testing.T) {
	dir := latestSnapshotDir(t)
	skipIfNoSnapshot(t, dir)

	jsPath := filepath.Join(dir, "web", "assets", "js", "model", "inbound.js")
	data, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("cannot read inbound.js: %v", err)
	}

	// Extract protocol values: SOMETHING: 'value',
	re := regexp.MustCompile(`\w+:\s*'([^']+)'`)
	// We only want lines between "const Protocols = {" and the closing "}"
	content := string(data)
	start := strings.Index(content, "const Protocols = {")
	if start == -1 {
		t.Fatal("cannot find Protocols object in inbound.js")
	}
	end := strings.Index(content[start:], "};")
	if end == -1 {
		t.Fatal("cannot find end of Protocols object in inbound.js")
	}
	block := content[start : start+end]

	matches := re.FindAllStringSubmatch(block, -1)
	if len(matches) == 0 {
		t.Fatal("no protocols found in Protocols object")
	}

	jsProtocols := make(map[string]bool)
	for _, m := range matches {
		jsProtocols[m[1]] = true
	}

	// Provider handles these protocols. "tun" is a newer alias for "tunnel"/"dokodemo-door"
	// in the UI; the provider maps "tunnel" and "dokodemo-door" to the same handler.
	providerHandled := map[string]bool{
		"vless":       true,
		"vmess":       true,
		"trojan":      true,
		"shadowsocks": true,
		"http":        true,
		"wireguard":   true,
		"tunnel":      true,
		"tun":         true, // UI alias for tunnel/dokodemo-door
		"mixed":       true,
		"hysteria":    true,
	}

	var missing []string
	for proto := range jsProtocols {
		if !providerHandled[proto] {
			missing = append(missing, proto)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("upstream inbound.js Protocols has entries not handled by provider: %v", missing)
	}
}

// --------------------------------------------------------------------------
// Test: upstream protocol HTML forms vs provider per-protocol settings blocks
// --------------------------------------------------------------------------

// TestDriftProtocolForms checks that every protocol form HTML file in
// web/html/form/protocol/ has a corresponding *_settings block in the provider schema.
func TestDriftProtocolForms(t *testing.T) {
	dir := latestSnapshotDir(t)
	skipIfNoSnapshot(t, dir)

	formDir := filepath.Join(dir, "web", "html", "form", "protocol")
	entries, err := os.ReadDir(formDir)
	if err != nil {
		t.Fatalf("cannot read protocol form dir: %v", err)
	}

	// Map from form file name (without .html) to expected provider handling.
	// Provider uses typed blocks: vless_settings, trojan_settings, etc.
	providerBlocks := map[string]bool{
		"vless":       true,
		"vmess":       true, // vmess has no settings block but is handled
		"trojan":      true,
		"shadowsocks": true,
		"http":        true,
		"socks":       true,
		"wireguard":   true,
		"dokodemo":    true, // mapped to dokodemo_settings
		"hysteria":    true,
		"tun":         true, // alias for tunnel/dokodemo-door
	}

	var missing []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".html")
		if !providerBlocks[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("upstream protocol form files not handled by provider: %v", missing)
	}
}

// --------------------------------------------------------------------------
// Test: upstream stream/transport HTML forms vs provider stream_settings blocks
// --------------------------------------------------------------------------

// TestDriftStreamSettingsForms checks that every stream settings HTML form in
// web/html/form/stream/ has a corresponding transport block in the provider schema.
func TestDriftStreamSettingsForms(t *testing.T) {
	dir := latestSnapshotDir(t)
	skipIfNoSnapshot(t, dir)

	streamDir := filepath.Join(dir, "web", "html", "form", "stream")
	entries, err := os.ReadDir(streamDir)
	if err != nil {
		t.Fatalf("cannot read stream form dir: %v", err)
	}

	// Map from stream form file name to expected provider handling.
	// Files like stream_tcp.html -> "tcp_settings" block in schema.
	providerHandled := map[string]bool{
		"stream_tcp":         true,
		"stream_ws":          true,
		"stream_grpc":        true,
		"stream_httpupgrade": true,
		"stream_xhttp":       true,
		"stream_kcp":         true,
		"stream_hysteria":    true,
		"stream_sockopt":     true,
		"stream_settings":    true, // main stream_settings container
		"stream_finalmask":   true, // finalmask is part of xhttp settings
		"external_proxy":     true, // external_proxy block
	}

	var missing []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".html")
		if !providerHandled[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("upstream stream form files not handled by provider: %v", missing)
	}
}

// --------------------------------------------------------------------------
// Test: upstream Inbound model fields vs provider Inbound struct
// --------------------------------------------------------------------------

// TestDriftInboundFields checks that key fields in the upstream Inbound struct
// (database/model/model.go) are present in the provider's Inbound struct (types.go).
func TestDriftInboundFields(t *testing.T) {
	dir := latestSnapshotDir(t)
	skipIfNoSnapshot(t, dir)

	modelPath := filepath.Join(dir, "database", "model", "model.go")
	data, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("cannot read model.go: %v", err)
	}

	// Extract json tags from Inbound struct
	upstreamFields := extractStructJSONTags(t, string(data), "Inbound")

	// Provider Inbound struct json tags (from types.go)
	providerFields := map[string]bool{
		"id": true, "up": true, "down": true, "total": true,
		"allTime": true, "remark": true, "enable": true,
		"expiryTime": true, "trafficReset": true, "lastTrafficResetTime": true,
		"clientStats": true, "listen": true, "port": true,
		"protocol": true, "settings": true, "streamSettings": true,
		"tag": true, "sniffing": true,
	}

	// Fields we intentionally skip (internal DB fields, not part of API contract)
	skip := map[string]bool{
		"-": true, // UserId uses json:"-"
	}

	var missing []string
	for _, f := range upstreamFields {
		if skip[f] {
			continue
		}
		if !providerFields[f] {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("upstream Inbound struct has json fields not in provider: %v", missing)
	}
}

// --------------------------------------------------------------------------
// Test: upstream Client model fields vs provider awareness
// --------------------------------------------------------------------------

// TestDriftClientFields checks that key fields in the upstream Client struct
// (database/model/model.go) are known to the provider.
func TestDriftClientFields(t *testing.T) {
	dir := latestSnapshotDir(t)
	skipIfNoSnapshot(t, dir)

	modelPath := filepath.Join(dir, "database", "model", "model.go")
	data, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("cannot read model.go: %v", err)
	}

	upstreamFields := extractStructJSONTags(t, string(data), "Client")

	// Provider knows about these client fields (from resource_inbound_client.go schema/expand/flatten).
	providerKnown := map[string]bool{
		"id": true, "security": true, "password": true, "flow": true,
		"auth": true, "email": true, "limitIp": true, "totalGB": true,
		"expiryTime": true, "enable": true, "tgId": true, "subId": true,
		"comment": true, "reset": true, "created_at": true, "updated_at": true,
	}

	var missing []string
	for _, f := range upstreamFields {
		if !providerKnown[f] {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("upstream Client struct has json fields not known to provider: %v", missing)
	}
}

// --------------------------------------------------------------------------
// Test: upstream AllSetting fields vs provider settings tabs
// --------------------------------------------------------------------------

// TestDriftAllSettingFields checks that fields in the upstream AllSetting struct
// (web/entity/entity.go) are known to the provider's settings resources.
func TestDriftAllSettingFields(t *testing.T) {
	dir := latestSnapshotDir(t)
	skipIfNoSnapshot(t, dir)

	entityPath := filepath.Join(dir, "web", "entity", "entity.go")
	data, err := os.ReadFile(entityPath)
	if err != nil {
		t.Fatalf("cannot read entity.go: %v", err)
	}

	upstreamFields := extractStructJSONTags(t, string(data), "AllSetting")

	// All fields handled across panel_general, panel_security, panel_telegram,
	// panel_subscription resources. Collected from expand/flatten functions.
	providerKnown := map[string]bool{
		// General
		"webListen": true, "webDomain": true, "webPort": true,
		"webBasePath": true, "sessionMaxAge": true, "pageSize": true,
		"remarkModel": true, "datepicker": true, "timeLocation": true,
		"expireDiff": true, "trafficDiff": true, "webCertFile": true,
		"webKeyFile": true, "externalTrafficInformEnable": true,
		"externalTrafficInformURI": true,
		// LDAP
		"ldapEnable": true, "ldapHost": true, "ldapPort": true,
		"ldapUseTLS": true, "ldapBindDN": true, "ldapPassword": true,
		"ldapBaseDN": true, "ldapUserFilter": true, "ldapUserAttr": true,
		"ldapVlessField": true, "ldapSyncCron": true,
		"ldapFlagField": true, "ldapTruthyValues": true, "ldapInvertFlag": true,
		"ldapInboundTags": true, "ldapAutoCreate": true, "ldapAutoDelete": true,
		"ldapDefaultTotalGB": true, "ldapDefaultExpiryDays": true, "ldapDefaultLimitIP": true,
		// Security
		"twoFactorEnable": true, "twoFactorToken": true,
		// Telegram
		"tgBotEnable": true, "tgBotToken": true, "tgBotProxy": true,
		"tgBotAPIServer": true, "tgBotChatId": true, "tgLang": true,
		"tgRunTime": true, "tgBotBackup": true, "tgBotLoginNotify": true,
		"tgCpu": true,
		// Subscription
		"subEnable": true, "subJsonEnable": true, "subTitle": true,
		"subSupportUrl": true, "subProfileUrl": true, "subAnnounce": true,
		"subEnableRouting": true, "subRoutingRules": true,
		"subListen": true, "subPort": true, "subPath": true,
		"subDomain": true, "subCertFile": true, "subKeyFile": true,
		"subUpdates": true, "subEncrypt": true, "subShowInfo": true,
		"subURI": true, "subJsonPath": true, "subJsonURI": true,
		"subJsonFragment": true, "subJsonNoises": true, "subJsonMux": true,
		"subJsonRules": true, "subClashEnable": true, "subClashPath": true,
		"subClashURI": true,
	}

	// Fields intentionally not managed by the provider (read-only in the UI, or
	// managed elsewhere, e.g. xrayOutboundTestUrl is in panel_general).
	intentionallySkipped := map[string]bool{
		// empty — extend this if upstream adds fields we deliberately skip
	}

	var missing []string
	for _, f := range upstreamFields {
		if providerKnown[f] || intentionallySkipped[f] {
			continue
		}
		missing = append(missing, f)
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("upstream AllSetting struct has json fields not handled by provider settings: %v\n"+
			"Add them to the appropriate provider settings resource or to intentionallySkipped in this test.", missing)
	}
}

// --------------------------------------------------------------------------
// Test: upstream xray settings HTML pages vs provider xray resources
// --------------------------------------------------------------------------

// TestDriftXraySettingsPages checks that every xray settings HTML page in
// web/html/settings/xray/ has a corresponding provider resource.
func TestDriftXraySettingsPages(t *testing.T) {
	dir := latestSnapshotDir(t)
	skipIfNoSnapshot(t, dir)

	xrayDir := filepath.Join(dir, "web", "html", "settings", "xray")
	entries, err := os.ReadDir(xrayDir)
	if err != nil {
		t.Fatalf("cannot read xray settings dir: %v", err)
	}

	// Provider xray resources mapped from HTML page names.
	providerResources := map[string]bool{
		"basics":    true, // threexui_xray_basics
		"dns":       true, // threexui_xray_dns
		"routing":   true, // threexui_xray_routing
		"outbounds": true, // threexui_xray_outbounds
		"balancers": true, // threexui_xray_balancers
		"reverse":   true, // threexui_xray_reverse
		"advanced":  true, // advanced is raw JSON editing, no dedicated resource needed
	}

	var missing []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".html")
		if !providerResources[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("upstream xray settings pages not handled by provider: %v", missing)
	}
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

// extractStructJSONTags extracts json tag values from a Go struct definition
// in source code. It finds "type <name> struct {" and collects all json:"xxx"
// tags until the closing "}".
func extractStructJSONTags(t *testing.T, source, structName string) []string {
	t.Helper()

	// Find struct start
	structRe := regexp.MustCompile(`type\s+` + regexp.QuoteMeta(structName) + `\s+struct\s*\{`)
	loc := structRe.FindStringIndex(source)
	if loc == nil {
		t.Fatalf("struct %s not found in source", structName)
	}

	// Find matching closing brace
	depth := 1
	pos := loc[1]
	for pos < len(source) && depth > 0 {
		switch source[pos] {
		case '{':
			depth++
		case '}':
			depth--
		}
		pos++
	}
	block := source[loc[1]:pos]

	// Extract json tags
	jsonRe := regexp.MustCompile(`json:"([^",]+)`)
	matches := jsonRe.FindAllStringSubmatch(block, -1)

	var tags []string
	for _, m := range matches {
		tags = append(tags, m[1])
	}
	return tags
}

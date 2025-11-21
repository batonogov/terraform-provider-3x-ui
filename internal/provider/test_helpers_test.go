package provider

import (
	"encoding/json"
)

func jsonRaw(s string) json.RawMessage {
	return json.RawMessage([]byte(s))
}

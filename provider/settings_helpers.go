package provider

func mergeSettings(existing, desired map[string]any) map[string]any {
	if existing == nil && desired == nil {
		return map[string]any{}
	}
	if existing == nil {
		return desired
	}
	out := make(map[string]any, len(existing)+len(desired))
	for k, v := range existing {
		out[k] = v
	}
	for k, v := range desired {
		out[k] = v
	}
	return out
}

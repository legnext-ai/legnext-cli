package cmd

// extractCostPoints pulls meta.usage.consume out of a task response map
// and returns it as an int. Returns 0 when unavailable.
func extractCostPoints(v map[string]any) int {
	meta, ok := v["meta"].(map[string]any)
	if !ok {
		return 0
	}
	usage, ok := meta["usage"].(map[string]any)
	if !ok {
		return 0
	}
	consume, ok := usage["consume"].(float64)
	if !ok {
		return 0
	}
	return int(consume)
}

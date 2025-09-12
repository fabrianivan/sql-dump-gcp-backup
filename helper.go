package main

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}
package goutils

// MergeMap merge source and source2 to out
func MergeMap(source1 map[string]interface{}, source2 map[string]interface{}) (out map[string]interface{}) {
	out = map[string]interface{}{}
	for key, value := range source1 {
		out[key] = value
	}
	for key, value := range source2 {
		out[key] = value
	}

	return out
}

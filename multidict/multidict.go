// Package multidict provides a map of string slices, useful for HTTP headers
// where a single key may appear multiple times (e.g. Set-Cookie, Accept).
package multidict

// MultiDict stores multiple string values per key in insertion order.
// Methods operate on the receiver pointer, so a MultiDict value should not be
// copied after use.
type MultiDict struct {
	_map map[string][]string
}

// NewMultiDict creates an empty, ready-to-use MultiDict.
func NewMultiDict() MultiDict {
	return MultiDict{
		_map: make(map[string][]string),
	}
}

// Set appends value to the values for key. It does not replace existing values.
// Keys are stored as-is; no normalization is applied.
func (m *MultiDict) Set(key string, value string) {
	m._map[key] = append(m._map[key], value)
}

// Get returns all values for key and a boolean indicating whether the key
// exists. The returned slice is a direct reference to internal storage;
// callers should treat it as read-only.
func (m *MultiDict) Get(key string) ([]string, bool) {
	values, ok := m._map[key]
	return values, ok
}

// GetOne returns the first value for key, or "" if the key is missing.
// To distinguish between a missing key and an explicitly empty value, use Get.
func (m *MultiDict) GetOne(key string) string {
	values := m._map[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Del removes all values associated with key.
func (m *MultiDict) Del(key string) {
	delete(m._map, key)
}

// Pop removes and returns all values for key, along with a boolean indicating
// whether the key existed.
func (m *MultiDict) Pop(key string) ([]string, bool) {
	values, ok := m._map[key]
	if ok {
		delete(m._map, key)
	}
	return values, ok
}

// Len returns the number of unique keys.
func (m *MultiDict) Len() int {
	return len(m._map)
}

// Map returns a shallow copy of the internal map. Mutating the returned map
// or its slices does not affect the MultiDict.
func (m *MultiDict) Map() map[string][]string {
	out := make(map[string][]string, len(m._map))
	for k, v := range m._map {
		out[k] = append([]string(nil), v...)
	}
	return out
}

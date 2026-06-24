package multidict

import (
	"testing"
)

func TestNewMultiDict(t *testing.T) {
	m := NewMultiDict()
	if m.Len() != 0 {
		t.Errorf("NewMultiDict should be empty, got Len() = %d", m.Len())
	}
}

func TestSetAndGet(t *testing.T) {
	m := NewMultiDict()
	m.Set("key1", "value1")

	vals, ok := m.Get("key1")
	if !ok {
		t.Errorf("Get should return ok=true after Set")
	}
	if len(vals) != 1 {
		t.Errorf("Get should return 1 value, got %d", len(vals))
	}
	if vals[0] != "value1" {
		t.Errorf("Get returned %q, want %q", vals[0], "value1")
	}
}

func TestSetAppendsValues(t *testing.T) {
	m := NewMultiDict()
	m.Set("key", "a")
	m.Set("key", "b")
	m.Set("key", "c")

	vals, ok := m.Get("key")
	if !ok {
		t.Errorf("Get should return ok=true after Set")
	}
	if len(vals) != 3 {
		t.Errorf("Get should return 3 values, got %d", len(vals))
	}
	if vals[0] != "a" || vals[1] != "b" || vals[2] != "c" {
		t.Errorf("Get returned %v, want [a b c]", vals)
	}
}

func TestSetDifferentKeys(t *testing.T) {
	m := NewMultiDict()
	m.Set("a", "1")
	m.Set("b", "2")

	if m.Len() != 2 {
		t.Errorf("Len should be 2, got %d", m.Len())
	}

	vals, ok := m.Get("a")
	if !ok || vals[0] != "1" {
		t.Errorf("Get(\"a\") = %v, %v; want [\"1\"], true", vals, ok)
	}

	vals, ok = m.Get("b")
	if !ok || vals[0] != "2" {
		t.Errorf("Get(\"b\") = %v, %v; want [\"2\"], true", vals, ok)
	}
}

func TestGetMissingKey(t *testing.T) {
	m := NewMultiDict()
	vals, ok := m.Get("nonexistent")
	if ok {
		t.Errorf("Get should return ok=false for missing key")
	}
	if vals != nil {
		t.Errorf("Get should return nil for missing key, got %v", vals)
	}
}

func TestGetOne(t *testing.T) {
	m := NewMultiDict()
	m.Set("key", "first")
	m.Set("key", "second")

	v := m.GetOne("key")
	if v != "first" {
		t.Errorf("GetOne should return the first value, got %q", v)
	}
}

func TestGetOneMissingKey(t *testing.T) {
	m := NewMultiDict()
	v := m.GetOne("missing")
	if v != "" {
		t.Errorf("GetOne should return empty string for missing key, got %q", v)
	}
}

func TestGetOneEmpty(t *testing.T) {
	m := NewMultiDict()
	v := m.GetOne("")
	if v != "" {
		t.Errorf("GetOne with empty key should return empty string, got %q", v)
	}
}

func TestSetEmptyKey(t *testing.T) {
	m := NewMultiDict()
	m.Set("", "val")
	vals, ok := m.Get("")
	if !ok || vals[0] != "val" {
		t.Errorf("Set with empty key should store the value")
	}
}

func TestLen(t *testing.T) {
	m := NewMultiDict()
	if m.Len() != 0 {
		t.Errorf("Empty dict Len() = %d, want 0", m.Len())
	}

	m.Set("a", "1")
	if m.Len() != 1 {
		t.Errorf("Len() after 1 key = %d, want 1", m.Len())
	}

	m.Set("b", "2")
	if m.Len() != 2 {
		t.Errorf("Len() after 2 keys = %d, want 2", m.Len())
	}

	// Duplicate key should not increase Len
	m.Set("a", "3")
	if m.Len() != 2 {
		t.Errorf("Len() after duplicate key = %d, want 2", m.Len())
	}
}

func TestMap(t *testing.T) {
	m := NewMultiDict()
	m.Set("host", "example.com")
	m.Set("accept", "text/html")
	m.Set("accept", "application/json")

	mp := m.Map()
	if len(mp) != 2 {
		t.Errorf("Map should have 2 keys, got %d", len(mp))
	}

	if _, ok := mp["host"]; !ok {
		t.Errorf("Map should contain \"host\"")
	}

	if _, ok := mp["accept"]; !ok {
		t.Errorf("Map should contain \"accept\"")
	}

	if mp["host"][0] != "example.com" {
		t.Errorf("Map[\"host\"] = %v, want [\"example.com\"]", mp["host"])
	}
}

func TestMapIsCopy(t *testing.T) {
	m := NewMultiDict()
	m.Set("key", "original")

	mp := m.Map()
	mp["key"][0] = "hacked"

	v := m.GetOne("key")
	if v != "original" {
		t.Errorf("Map should return a copy, got %q, want %q", v, "original")
	}

	// Adding a key to the copy should not affect the original
	mp["new"] = []string{"nope"}
	if m.Len() != 1 {
		t.Errorf("Original should be unchanged after mutating the copy, Len = %d", m.Len())
	}
}

func TestDel(t *testing.T) {
	m := NewMultiDict()
	m.Set("a", "1")
	m.Set("b", "2")

	m.Del("a")

	if m.Len() != 1 {
		t.Errorf("Len after Del should be 1, got %d", m.Len())
	}

	if _, ok := m.Get("a"); ok {
		t.Errorf("Get should return ok=false after Del")
	}

	// Del on missing key should be a no-op
	m.Del("nonexistent")
	if m.Len() != 1 {
		t.Errorf("Len should still be 1 after deleting missing key, got %d", m.Len())
	}
}

func TestPop(t *testing.T) {
	m := NewMultiDict()
	m.Set("a", "1")
	m.Set("a", "2")
	m.Set("b", "3")

	vals, ok := m.Pop("a")
	if !ok {
		t.Errorf("Pop should return ok=true for existing key")
	}
	if len(vals) != 2 || vals[0] != "1" || vals[1] != "2" {
		t.Errorf("Pop returned %v, want [\"1\" \"2\"]", vals)
	}
	if m.Len() != 1 {
		t.Errorf("Len after Pop should be 1, got %d", m.Len())
	}

	// Pop on missing key
	vals, ok = m.Pop("nonexistent")
	if ok {
		t.Errorf("Pop should return ok=false for missing key")
	}
	if vals != nil {
		t.Errorf("Pop should return nil for missing key")
	}
	if m.Len() != 1 {
		t.Errorf("Len should be unchanged after popping missing key, got %d", m.Len())
	}
}

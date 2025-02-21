package structs

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/emirpasic/gods/maps/hashmap"
)

func TestNewMap(t *testing.T) {
	m := NewMap()
	if m == nil {
		t.Error("NewMap() returned nil")
	}
	if _, ok := m.(*MapImpl); !ok {
		t.Errorf("NewMap() returned wrong type: %T", m)
	}
}

func TestMapImpl_Get(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.Put("key1", "value1")
	m.Put("key2", 123)

	val, found := m.Get("key1")
	if !found {
		t.Error("Get() failed to find existing key")
	}
	if val != "value1" {
		t.Errorf("Get() returned wrong value: got %v, want %v", val, "value1")
	}

	val, found = m.Get("nonexistent_key")
	if found {
		t.Error("Get() found nonexistent key")
	}
	if val != nil {
		t.Errorf("Get() returned non-nil value for nonexistent key: %v", val)
	}
}

func TestMapImpl_GetInt(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.Put("key1", 123)
	m.Put("key2", "string_value")

	val, err := m.GetInt("key1")
	if err != nil {
		t.Errorf("GetInt() returned an error for a valid int key: %v", err)
	}
	if val == nil {
		t.Error("GetInt() returned nil for existing int key")
	}
	if *val != 123 {
		t.Errorf("GetInt() returned wrong value: got %v, want %v", *val, 123)
	}

	val, err = m.GetInt("nonexistent_key")
	if err != nil {
		t.Errorf("GetInt() returned an error for a nonexistent key: %v", err)
	}

	if val != nil {
		t.Error("GetInt() returned non-nil value for nonexistent key")
	}

	// Test for type safety, this should return an error
	val, err = m.GetInt("key2")
	if err == nil {
		t.Errorf("GetInt() did not return an error when given the wrong type.")
	}
	if val != nil {
		t.Errorf("GetInt() returned a non-nil value despite the type error")
	}
}

func TestMapImpl_GetUInt(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.Put("key1", uint(123))
	m.Put("key2", "string_value")

	val, err := m.GetUInt("key1")
	if err != nil {
		t.Errorf("GetUInt() returned an error for a valid uint key: %v", err)
	}
	if val == nil {
		t.Error("GetUInt() returned nil for existing uint key")
	}
	if *val != 123 {
		t.Errorf("GetUInt() returned wrong value: got %v, want %v", *val, 123)
	}

	val, err = m.GetUInt("nonexistent_key")
	if err != nil {
		t.Errorf("GetUInt() returned an error for a nonexistent key: %v", err)
	}

	if val != nil {
		t.Error("GetUInt() returned non-nil value for nonexistent key")
	}

	// Test for type safety, this should return an error
	val, err = m.GetUInt("key2")
	if err == nil {
		t.Errorf("GetUInt() did not return an error when given the wrong type.")
	}

	if val != nil {
		t.Errorf("GetUInt() returned a non-nil value despite the type error")
	}
}

func TestMapImpl_GetString(t *testing.T) {
	m := NewMap().(*MapImpl)
	stringValue := "string_value"
	m.Put("key1", stringValue)
	m.Put("key2", 123)

	val, err := m.GetString("key1")
	if err != nil {
		t.Errorf("GetString() returned an error for a valid string key: %v", err)
	}

	if val == nil {
		t.Error("GetString() returned nil for existing string key")
	}
	if *val != stringValue {
		t.Errorf("GetString() returned wrong value: got %v, want %v", *val, stringValue)
	}

	val, err = m.GetString("nonexistent_key")
	if err != nil {
		t.Errorf("GetString() returned an error for a nonexistent key: %v", err)
	}

	if val != nil {
		t.Error("GetString() returned non-nil value for nonexistent key")
	}

	// Test for type safety, this should return an error
	val, err = m.GetString("key2")
	if err == nil {
		t.Errorf("GetString() did not return an error when given the wrong type.")
	}
	if val != nil {
		t.Errorf("GetString() returned a non-nil value despite the type error")
	}
}

func TestMapImpl_Put(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.Put("key1", "value1")
	val, found := m.Get("key1")
	if !found {
		t.Error("Put() failed to insert value")
	}
	if val != "value1" {
		t.Errorf("Put() inserted wrong value: got %v, want %v", val, "value1")
	}
}

func TestMapImpl_PutInt(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.PutInt("key1", 123)
	val, found := m.Get("key1")
	if !found {
		t.Error("PutInt() failed to insert value")
	}
	if val != 123 {
		t.Errorf("PutInt() inserted wrong value: got %v, want %v", val, 123)
	}
}

func TestMapImpl_PutUInt(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.PutUInt("key1", uint(123))
	val, found := m.Get("key1")
	if !found {
		t.Error("PutUInt() failed to insert value")
	}
	if val != uint(123) {
		t.Errorf("PutUInt() inserted wrong value: got %v, want %v", val, 123)
	}
}

func TestMapImpl_Remove(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.Put("key1", "value1")
	m.Remove("key1")
	_, found := m.Get("key1")
	if found {
		t.Error("Remove() failed to remove key")
	}
}

func TestMapImpl_Keys(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.Put("key1", "value1")
	m.Put("key2", "value2")
	keys := m.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys() returned wrong number of keys: got %v, want %v", len(keys), 2)
	}

	expectedKeys := []interface{}{"key1", "key2"}
	stringKeys := make([]string, len(keys))
	for i, k := range keys {
		stringKeys[i] = fmt.Sprintf("%v", k)
	}

	expectedStringKeys := make([]string, len(expectedKeys))
	for i, k := range expectedKeys {
		expectedStringKeys[i] = fmt.Sprintf("%v", k)
	}

	if !reflect.DeepEqual(stringKeys, expectedStringKeys) {
		t.Errorf("Keys() returned wrong keys: got %v, want %v", stringKeys, expectedStringKeys)
	}
}

func TestMapImpl_Values(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.Put("key1", "value1")
	m.Put("key2", "value2")
	values := m.Values()
	if len(values) != 2 {
		t.Errorf("Values() returned wrong number of values: got %v, want %v", len(values), 2)
	}

	expectedValues := []interface{}{"value1", "value2"}
	stringValues := make([]string, len(values))
	for i, v := range values {
		stringValues[i] = fmt.Sprintf("%v", v)
	}

	expectedStringValues := make([]string, len(expectedValues))
	for i, v := range expectedValues {
		expectedStringValues[i] = fmt.Sprintf("%v", v)
	}

	// Sort both slices before comparison.
	sort.Strings(stringValues)
	sort.Strings(expectedStringValues)

	if !reflect.DeepEqual(stringValues, expectedStringValues) {
		t.Errorf("Values() returned wrong values: got %v, want %v", stringValues, expectedStringValues)
	}
}

func TestMapImpl_Size(t *testing.T) {
	m := NewMap().(*MapImpl)
	if m.Size() != 0 {
		t.Errorf("Size() returned wrong size for empty map: got %v, want %v", m.Size(), 0)
	}
	m.Put("key1", "value1")
	if m.Size() != 1 {
		t.Errorf("Size() returned wrong size after put: got %v, want %v", m.Size(), 1)
	}
	m.Remove("key1")
	if m.Size() != 0 {
		t.Errorf("Size() returned wrong size after remove: got %v, want %v", m.Size(), 0)
	}
}

func TestMapImpl_Empty(t *testing.T) {
	m := NewMap().(*MapImpl)
	if !m.Empty() {
		t.Error("Empty() returned false for empty map")
	}
	m.Put("key1", "value1")
	if m.Empty() {
		t.Error("Empty() returned true for non-empty map")
	}
}

func TestMapImpl_Clear(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.Put("key1", "value1")
	m.Clear()
	if !m.Empty() {
		t.Error("Clear() failed to clear map")
	}
	if m.Size() != 0 {
		t.Error("Clear() failed to clear map size")
	}
}

func TestMapImpl_String(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.Put("key1", "value1")
	str := m.String()
	expected := "{\"key1\":\"value1\"}"
	if str != expected {
		t.Errorf("String() returned wrong string: got %v, want %v", str, expected)
	}
}

func TestMapImpl_GetKey(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.Put("key1", "value1")
	key, found := m.GetKey("value1")
	if !found {
		t.Error("GetKey() failed to find value")
	}
	if key != "key1" {
		t.Errorf("GetKey() returned wrong key: got %v, want %v", key, "key1")
	}
	_, found = m.GetKey("nonexistent_value")
	if found {
		t.Error("GetKey() found nonexistent value")
	}
}

func TestMapImpl_Contains(t *testing.T) {
	m := NewMap().(*MapImpl)
	m.Put("key1", "value1")
	if !m.Contains("key1") {
		t.Error("Contains() returned false for existing key")
	}
	if m.Contains("nonexistent_key") {
		t.Error("Contains() returned true for nonexistent key")
	}
}

func TestMapImpl_Concurrency(t *testing.T) {
	m := NewMap().(*MapImpl)
	var wg sync.WaitGroup
	numRoutines := 100

	wg.Add(numRoutines * 2) // Each routine will read and write

	for i := 0; i < numRoutines; i++ {
		go func(key int) {
			defer wg.Done()
			m.Put(key, key*2)
			_ = m.Size()
		}(i)

		go func(key int) {
			defer wg.Done()
			_, _ = m.Get(key)
			_ = m.Empty()
		}(i)
	}

	wg.Wait()

	if m.Size() != numRoutines {
		t.Errorf("Concurrency test failed: Expected size %d, got %d", numRoutines, m.Size())
	}

	for i := 0; i < numRoutines; i++ {
		val, found := m.Get(i)
		if !found {
			t.Errorf("Concurrency test failed: Key %d not found", i)
		}
		if val != i*2 {
			t.Errorf("Concurrency test failed: Wrong value for key %d, expected %d, got %v", i, i*2, val)
		}
	}
}

func TestMapImpl_InterfaceConformance(t *testing.T) {
	var _ Map = &MapImpl{
		Map:   hashmap.New(),
		mutex: &sync.RWMutex{},
	}
}

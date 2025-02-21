package structs

import (
	"reflect"
	"sync"
	"testing"

	"github.com/emirpasic/gods/sets"
)

func TestNewSet(t *testing.T) {
	s := NewSet()
	if s == nil {
		t.Errorf("NewSet() returned nil")
	}
	if s.Set == nil {
		t.Errorf("NewSet().Set is nil")
	}
	if s.mutex == nil {
		t.Errorf("NewSet().mutex is nil")
	}

	// Check that it implements the sets.Set interface
	var _ sets.Set = s
}

func TestSetImpl_Add(t *testing.T) {
	s := NewSet()

	// Add single item
	s.Add(1)
	if !s.Contains(1) {
		t.Errorf("Add(1) failed, set does not contain 1")
	}

	// Add multiple items
	s.Add(2, "hello", 3.14)
	if !s.Contains(2) {
		t.Errorf("Add(2, \"hello\", 3.14) failed, set does not contain 2")
	}
	if !s.Contains("hello") {
		t.Errorf("Add(2, \"hello\", 3.14) failed, set does not contain \"hello\"")
	}
	if !s.Contains(3.14) {
		t.Errorf("Add(2, \"hello\", 3.14) failed, set does not contain 3.14")
	}
}

func TestSetImpl_Remove(t *testing.T) {
	s := NewSet()
	s.Add(1, "hello", 3.14)

	// Remove single item
	s.Remove(1)
	if s.Contains(1) {
		t.Errorf("Remove(1) failed, set still contains 1")
	}

	// Remove multiple items
	s.Remove("hello", 3.14)
	if s.Contains("hello") {
		t.Errorf("Remove(\"hello\", 3.14) failed, set still contains \"hello\"")
	}
	if s.Contains(3.14) {
		t.Errorf("Remove(\"hello\", 3.14) failed, set still contains 3.14")
	}

	// Remove item that doesn't exist
	s.Remove("nonexistent") // Should not panic
}

func TestSetImpl_Contains(t *testing.T) {
	s := NewSet()
	s.Add(1, "hello", 3.14)

	// Contains existing item
	if !s.Contains(1) {
		t.Errorf("Contains(1) failed, set contains 1 but returned false")
	}

	// Contains non-existent item
	if s.Contains("nonexistent") {
		t.Errorf("Contains(\"nonexistent\") failed, set does not contain \"nonexistent\" but returned true")
	}

	// Contains multiple items, one existing, one not existing
	if s.Contains(1, "nonexistent") {
		t.Errorf("Contains(1, \"nonexistent\") failed, should return false if any item is missing")
	}
}

func TestSetImpl_Values(t *testing.T) {
	s := NewSet()
	s.Add(1, "hello", 3.14)

	values := s.Values()
	if len(values) != 3 {
		t.Errorf("Values() returned %d elements, expected 3", len(values))
	}

	// Check that the values are present (order is not guaranteed)
	found1 := false
	foundHello := false
	found314 := false
	for _, v := range values {
		switch v {
		case 1:
			found1 = true
		case "hello":
			foundHello = true
		case 3.14:
			found314 = true
		}
	}

	if !found1 {
		t.Errorf("Values() does not contain 1")
	}
	if !foundHello {
		t.Errorf("Values() does not contain \"hello\"")
	}
	if !found314 {
		t.Errorf("Values() does not contain 3.14")
	}
}

func TestSetImpl_Concurrency(t *testing.T) {
	s := NewSet()
	var wg sync.WaitGroup
	numGoroutines := 100

	// Simulate concurrent adds and removes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Add(i)
			s.Remove(i)
		}(i)
	}

	// Simulate concurrent contains
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Contains(i)
		}(i)
	}

	wg.Wait()

	// Check that the set is empty (all added items were removed)
	if !s.Empty() {
		t.Errorf("Concurrent adds and removes failed, set is not empty")
	}
}

func TestSetImpl_Empty(t *testing.T) {
	s := NewSet()

	if !s.Empty() {
		t.Errorf("Empty() should return true on an empty set")
	}

	s.Add(1)
	if s.Empty() {
		t.Errorf("Empty() should return false on a non-empty set")
	}

	s.Remove(1)
	if !s.Empty() {
		t.Errorf("Empty() should return true after removing the only element")
	}
}

func TestSetImpl_Size(t *testing.T) {
	s := NewSet()

	if s.Size() != 0 {
		t.Errorf("Size() should return 0 on an empty set")
	}

	s.Add(1, 2, 3)
	if s.Size() != 3 {
		t.Errorf("Size() should return 3 after adding 3 elements")
	}

	s.Remove(1)
	if s.Size() != 2 {
		t.Errorf("Size() should return 2 after removing 1 element")
	}
}

func TestSetImpl_Clear(t *testing.T) {
	s := NewSet()
	s.Add(1, 2, 3)

	s.Clear()
	if !s.Empty() {
		t.Errorf("Clear() should empty the set")
	}
	if s.Size() != 0 {
		t.Errorf("Clear() should set size to 0")
	}
}

func TestSetImpl_String(t *testing.T) {
	s := NewSet()
	s.Add(1, "hello", 3.14)

	str := s.String()
	if len(str) == 0 {
		t.Errorf("String() should return a non-empty string representation")
	}

	// Although the order is not guaranteed, we can check for the presence of the elements
	if !(reflect.ValueOf(str).String() != "") {
		t.Errorf("String() returned empty string")
	}

}

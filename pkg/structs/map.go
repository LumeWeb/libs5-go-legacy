package structs

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/hashmap"
)

var _ maps.Map = (*MapImpl)(nil)

type Map interface {
	GetInt(key interface{}) (value *int, err error)
	GetUInt(key interface{}) (value *uint, err error)
	GetString(key interface{}) (value *string, err error)
	PutInt(key interface{}, value int)
	PutUInt(key interface{}, value uint)
	Contains(value interface{}) bool
	maps.Map
}

type MapImpl struct {
	*hashmap.Map
	mutex *sync.RWMutex
}

func NewMap() Map {
	return &MapImpl{
		Map:   hashmap.New(),
		mutex: &sync.RWMutex{},
	}
}

func (m *MapImpl) Get(key interface{}) (value interface{}, found bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.Map.Get(key)
}

func (m *MapImpl) GetInt(key interface{}) (value *int, err error) {
	val, found := m.Get(key)

	if !found {
		return nil, nil // Not found is not an error
	}

	if intValue, ok := val.(int); ok {
		value = &intValue
	} else {
		return nil, fmt.Errorf("value is not an int: %v", val) // Return an error instead of fatal
	}

	return value, nil
}

func (m *MapImpl) GetUInt(key interface{}) (value *uint, err error) {
	val, found := m.Get(key)

	if !found {
		return nil, nil
	}

	if intValue, ok := val.(uint); ok {
		value = &intValue
	} else {
		return nil, fmt.Errorf("value is not an uint: %v", val)
	}

	return value, nil
}

func (m *MapImpl) GetString(key interface{}) (value *string, err error) {
	val, found := m.Get(key)

	if !found {
		return nil, nil
	}

	if strValue, ok := val.(string); ok {
		value = &strValue
	} else {
		return nil, fmt.Errorf("value is not a string: %v", val)
	}

	return value, nil
}

func (m *MapImpl) Put(key interface{}, value interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Map.Put(key, value)
}

func (m *MapImpl) PutInt(key interface{}, value int) {
	m.Put(key, value)
}

func (m *MapImpl) PutUInt(key interface{}, value uint) {
	m.Put(key, value)
}

func (m *MapImpl) Remove(key interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Map.Remove(key)
}

func (m *MapImpl) Keys() []interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.Map.Keys()
}

func (m *MapImpl) Values() []interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.Map.Values()
}

func (m *MapImpl) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.Map.Size()
}

func (m *MapImpl) Empty() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.Map.Empty()
}

func (m *MapImpl) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Map.Clear()
}

func (m *MapImpl) String() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Build a Go map for JSON marshaling.
	goMap := make(map[string]interface{})
	for _, key := range m.Keys() {
		val, _ := m.Get(key) // Ignoring found here, keys are guaranteed to exist
		goMap[fmt.Sprintf("%v", key)] = val
	}

	// Convert the map to JSON
	jsonBytes, err := json.Marshal(goMap)
	if err != nil {
		return fmt.Sprintf("Error converting to JSON: %v", err)
	}
	return string(jsonBytes)
}

func (m *MapImpl) GetKey(value interface{}) (key interface{}, found bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, k := range m.Keys() {
		val, _ := m.Get(k)
		if val == value {
			return k, true
		}
	}
	return nil, false
}

func (m *MapImpl) Contains(value interface{}) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, has := m.Map.Get(value)

	return has
}

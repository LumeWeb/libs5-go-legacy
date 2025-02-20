package structs

import (
	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/hashmap"
	"log"
	"sync"
)

var _ maps.Map = (*MapImpl)(nil)

type Map interface {
	GetInt(key interface{}) (value *int)
	GetUInt(key interface{}) (value *uint)
	GetString(key interface{}) (value *string)
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

func (m *MapImpl) GetInt(key interface{}) (value *int) {
	val, found := m.Get(key)

	if !found {
		return nil
	}

	if intValue, ok := val.(int); ok {
		value = &intValue
	} else {
		log.Fatalf("value is not an int: %v", val)
	}

	return value
}

func (m *MapImpl) GetUInt(key interface{}) (value *uint) {
	val, found := m.Get(key)

	if !found {
		return nil
	}

	if intValue, ok := val.(uint); ok {
		value = &intValue
	} else {
		log.Fatalf("value is not an uint: %v", val)
	}

	return value
}

func (m *MapImpl) GetString(key interface{}) (value *string) {
	val, found := m.Get(key)

	if !found {
		return nil
	}

	if _, ok := val.(string); ok {
		value = val.(*string)
	} else {
		log.Fatalf("value is not a string: %v", value)
	}

	return
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
	return m.Map.String()
}

func (m *MapImpl) GetKey(value interface{}) (key interface{}, found bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.Map.Get(value)
}

func (m *MapImpl) Contains(value interface{}) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, has := m.Map.Get(value)

	return has
}

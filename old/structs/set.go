package structs

import (
	"github.com/emirpasic/gods/sets"
	"github.com/emirpasic/gods/sets/hashset"
	"sync"
)

var _ sets.Set = (*SetImpl)(nil)

type SetImpl struct {
	*hashset.Set
	mutex *sync.RWMutex
}

func NewSet() *SetImpl {
	return &SetImpl{
		Set:   hashset.New(),
		mutex: &sync.RWMutex{},
	}
}

func (s *SetImpl) Add(items ...interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Set.Add(items...)
}

func (s *SetImpl) Remove(items ...interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Set.Remove(items...)
}

func (s *SetImpl) Contains(items ...interface{}) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Set.Contains(items...)
}
func (s *SetImpl) Values() []interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Set.Values()
}

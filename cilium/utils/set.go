package utils

import (
	"sync"
)

type Set struct {
	m map[string]uint
	sync.RWMutex
}

func NewSet() *Set {
	return &Set{
		m: make(map[string]uint),
	}
}

// Add adds a new item to the set and returns true if the item is new to the
// set.
func (s *Set) Add(item string) uint {
	s.Lock()
	defer s.Unlock()
	if v, ok := s.m[item]; ok {
		return v
	}
	s.m[item] = 0
	return 0
}

// Add adds a new item to the set and returns true if the item is new to the
// set.
func (s *Set) IncFail(item string) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.m[item]; ok {
		s.m[item]++
	}
}

// Remove removes the item from the set and returns true if the item was
// found.
func (s *Set) Remove(item string) bool {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.m[item]; ok {
		delete(s.m, item)
		return true
	}
	return false
}

func (s *Set) Has(item string) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.m[item]
	return ok
}

func (s *Set) Len() int {
	return len(s.List())
}

func (s *Set) Clear() {
	s.Lock()
	defer s.Unlock()
	s.m = make(map[string]uint)
}

func (s *Set) IsEmpty() bool {
	return s.Len() == 0
}

func (s *Set) List() []string {
	s.RLock()
	defer s.RUnlock()
	list := make([]string, 0)
	for item := range s.m {
		list = append(list, item)
	}
	return list
}

package utils

import (
	"sync"
)

const (
	Configured = 100
	Failed     = 200
)

type Set struct {
	m map[string]int
	sync.RWMutex
}

func NewSet() *Set {
	return &Set{
		m: make(map[string]int),
	}
}

// Add adds a new item to the set and returns the number of attemps.
func (s *Set) Add(item string) int {
	s.Lock()
	defer s.Unlock()
	if v, ok := s.m[item]; ok {
		return v
	}
	s.m[item] = 0
	return 0
}

// Increments the number of attemps for the given item and returns the attemps
// number.
func (s *Set) IncFail(item string) int {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.m[item]; ok {
		s.m[item]++
		return s.m[item]
	}
	return 0
}

// Increments the number of attemps for the given item.
func (s *Set) Set(item string, value int) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.m[item]; ok {
		s.m[item] = value
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
	s.m = make(map[string]int)
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

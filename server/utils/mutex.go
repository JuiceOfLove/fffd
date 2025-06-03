package utils

import "sync"

type Mutex struct {
	mu sync.Mutex
}

func NewMutex() *Mutex {
	return &Mutex{}
}

func (m *Mutex) Lock() {
	m.mu.Lock()
}

func (m *Mutex) Unlock() {
	m.mu.Unlock()
}

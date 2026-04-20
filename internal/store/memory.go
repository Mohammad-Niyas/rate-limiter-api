package store

import (
	"sync"
	"time"
)

type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]*UserData
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]*UserData),
	}
}

func (m *MemoryStore) AddRequest(userID string, timestamp time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[userID]; !exists {
		m.data[userID] = &UserData{
			Timestamps: []time.Time{},
			TotalCount: 0,
		}
	}

	m.data[userID].Timestamps = append(m.data[userID].Timestamps, timestamp)
	m.data[userID].TotalCount++
}

func (m *MemoryStore) GetTimestamps(userID string, windowStart time.Time) []time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userData, exists := m.data[userID]
	if !exists {
		return []time.Time{}
	}

	var valid []time.Time
	for _, ts := range userData.Timestamps {
		if ts.After(windowStart) {
			valid = append(valid, ts)
		}
	}

	return valid
}

func (m *MemoryStore) GetAllUsers() map[string]*UserData {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data
}

func (m *MemoryStore) CleanExpired(userID string, windowStart time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userData, exists := m.data[userID]
	if !exists {
		return
	}

	var valid []time.Time
	for _, ts := range userData.Timestamps {
		if ts.After(windowStart) {
			valid = append(valid, ts)
		}
	}

	userData.Timestamps = valid
}

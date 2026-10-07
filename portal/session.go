// portal/session.go
package portal

import (
	"context"
	"log"
	"sync"
	"time"
)

// SessionData holds the information tied to a valid token in RAM.
type SessionData struct {
	UserID    int
	ExpiresAt time.Time
}

// MemorySessionStore manages active web portal sessions strictly in RAM.
// It uses a Read-Write Mutex to prevent data-race crashes when multiple APIs are called simultaneously.
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]SessionData
}

// NewMemorySessionStore initializes the empty map.
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]SessionData),
	}
}

// Set adds a new session token to RAM.
// We use Lock() to block other threads from writing at the exact same microsecond.
func (store *MemorySessionStore) Set(token string, userID int, expiresAt time.Time) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.sessions[token] = SessionData{
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
}

// Get retrieves a session from RAM.
// We use RLock() (Read Lock) which allows multiple threads to read simultaneously, 
// but prevents any writes while reading is happening.
func (store *MemorySessionStore) Get(token string) (SessionData, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	data, exists := store.sessions[token]
	return data, exists
}

// Delete removes a session (e.g., when a user logs out).
func (store *MemorySessionStore) Delete(token string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.sessions, token)
}

// StartGarbageCollector runs in the background and sweeps RAM for expired tokens.
// This prevents memory leaks if users never explicitly log out.
func (store *MemorySessionStore) StartGarbageCollector(ctx context.Context) {
	// Sweep RAM every 10 minutes
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[SESSIONS] RAM Garbage collector shut down.")
			return
		case <-ticker.C:
			store.cleanup()
		}
	}
}

// cleanup iterates over the map and deletes expired entries.
func (store *MemorySessionStore) cleanup() {
	now := time.Now()
	
	// We need a full Lock because we are modifying the map
	store.mu.Lock()
	defer store.mu.Unlock()

	deletedCount := 0
	for token, data := range store.sessions {
		if now.After(data.ExpiresAt) {
			delete(store.sessions, token)
			deletedCount++
		}
	}

	if deletedCount > 0 {
		log.Printf("[SESSIONS] RAM cleanup: purged %d expired sessions.\n", deletedCount)
	}
}
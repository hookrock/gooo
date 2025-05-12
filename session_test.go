package gooo

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestMemoryStore_GetSet(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	sessionID := "test_session"
	data := map[string]any{"key": "value"}

	// Test Set
	if err := store.Set(sessionID, data); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test Get
	retrievedData, err := store.Get(sessionID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrievedData["key"] != "value" {
		t.Errorf("Expected 'value', got '%v'", retrievedData["key"])
	}
}

func TestMemoryStore_Destroy(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	sessionID := "test_session"
	data := map[string]any{"key": "value"}

	store.Set(sessionID, data)
	if err := store.Destroy(sessionID); err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	_, err := store.Get(sessionID)
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	sessionID := "test_session"
	data := map[string]any{"counter": 0}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Set(sessionID, data)
			store.Get(sessionID)
		}()
	}
	wg.Wait()
}

func TestSessionManager_Create(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	manager := NewSessionManager(store)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Test new session creation
	data, sessionID := manager.Create(w, req)
	if sessionID == "" {
		t.Error("Expected non-empty session ID")
	}

	data["test"] = "value"

	// Test existing session retrieval
	_, sameSessionID := manager.Create(w, req)
	if sessionID != sameSessionID {
		t.Error("Session IDs should match for same request")
	}
}

func TestSessionManager_DestroySession(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	manager := NewSessionManager(store)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	_, sessionID := manager.Create(w, req)
	manager.DestroySession(w, sessionID)

	_, err := store.Get(sessionID)
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound after destroy, got %v", err)
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID failed: %v", err)
	}

	id2, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID failed: %v", err)
	}

	if id1 == id2 {
		t.Error("Generated session IDs should be unique")
	}

	if len(id1) < SessionIDLength/2 { // Base64 encoding expands the length
		t.Errorf("Session ID too short: %d", len(id1))
	}
}

func TestMemoryStore_GC(t *testing.T) {
	store := NewMemoryStore(time.Millisecond * 100)
	sessionID := "test_session"
	data := map[string]any{"key": "value"}

	store.Set(sessionID, data)
	time.Sleep(time.Millisecond * 150) // Wait for GC to run

	_, err := store.Get(sessionID)
	if err != ErrSessionNotFound {
		t.Errorf("Expected session to be garbage collected, got %v", err)
	}
}

func TestSessionMiddleware(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	manager := NewSessionManager(store)
	engine := &Engine{sessionManager: manager}

	handler := SessionMiddleware()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	ctx := newContext(w, req)
	ctx.engine = engine

	handler(ctx)

	if ctx.Session == nil {
		t.Error("Session should be initialized by middleware")
	}

	// Test session persistence
	ctx.Session["test"] = "value"
	handler(ctx) // Should save session
}

func TestSessionManager_RegenerateID(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	manager := NewSessionManager(store)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	_, oldID := manager.Create(w, req)
	newID := manager.RegenerateID(w, oldID)

	if newID == "" {
		t.Error("Expected non-empty new session ID")
	}

	if newID == oldID {
		t.Error("New session ID should be different from old one")
	}

	_, err := store.Get(oldID)
	if err != ErrSessionNotFound {
		t.Error("Old session should be destroyed")
	}

	_, err = store.Get(newID)
	if err != nil {
		t.Error("New session should exist")
	}
}

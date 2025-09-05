package kvstore_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xtdlib/kvstore"
)

func TestWatch(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_watch.db")
	
	store, err := kvstore.NewAt[string, string](dbPath, "test_watch")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	
	// Start watching a specific key - returns channel and cancel func
	eventCh, cancel := store.Watch("key1")
	defer cancel()
	
	// Test Set event
	err = store.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	
	// Wait for event
	select {
	case event := <-eventCh:
		if event.Type != kvstore.WatchEventSet {
			t.Errorf("Expected WatchEventSet, got %v", event.Type)
		}
		if event.Key != "key1" {
			t.Errorf("Expected key1, got %v", event.Key)
		}
		if event.Value != "value1" {
			t.Errorf("Expected value1, got %v", event.Value)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for set event")
	}
	
	// Test Update event (with old value)
	err = store.Set("key1", "value2")
	if err != nil {
		t.Fatalf("Failed to update key: %v", err)
	}
	
	select {
	case event := <-eventCh:
		if event.Type != kvstore.WatchEventSet {
			t.Errorf("Expected WatchEventSet, got %v", event.Type)
		}
		if event.Value != "value2" {
			t.Errorf("Expected value2, got %v", event.Value)
		}
		if event.OldValue != "value1" {
			t.Errorf("Expected old value1, got %v", event.OldValue)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for update event")
	}
	
	// Test Delete event
	err = store.Delete("key1")
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}
	
	select {
	case event := <-eventCh:
		if event.Type != kvstore.WatchEventDelete {
			t.Errorf("Expected WatchEventDelete, got %v", event.Type)
		}
		if event.Key != "key1" {
			t.Errorf("Expected key1, got %v", event.Key)
		}
		if event.OldValue != "value2" {
			t.Errorf("Expected old value2, got %v", event.OldValue)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for delete event")
	}
	
	// Test that we don't receive events for other keys
	err = store.Set("key2", "value_other")
	if err != nil {
		t.Fatalf("Failed to set key2: %v", err)
	}
	
	select {
	case event := <-eventCh:
		t.Errorf("Unexpected event for key2: %v", event)
	case <-time.After(200 * time.Millisecond):
		// Expected timeout - we shouldn't receive this event
	}
}

func TestWatchPrefix(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_watch_prefix.db")
	
	store, err := kvstore.NewAt[string, string](dbPath, "test_prefix")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	
	// Watch all keys with prefix "user:" - returns channel and cancel func
	eventCh, cancel := store.WatchPrefix("user:")
	defer cancel()
	
	// Test multiple keys with the same prefix
	keys := []string{"user:1", "user:2", "user:admin"}
	values := []string{"Alice", "Bob", "Admin"}
	
	for i, key := range keys {
		err = store.Set(key, values[i])
		if err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
	}
	
	// Should receive 3 events
	for i := 0; i < 3; i++ {
		select {
		case event := <-eventCh:
			if event.Type != kvstore.WatchEventSet {
				t.Errorf("Expected WatchEventSet, got %v", event.Type)
			}
			// Check that the key has the expected prefix
			keyStr := fmt.Sprintf("%v", event.Key)
			if len(keyStr) < 5 || keyStr[:5] != "user:" {
				t.Errorf("Expected key with 'user:' prefix, got %v", event.Key)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Timeout waiting for event %d", i+1)
		}
	}
	
	// Set a key without the prefix - should not trigger an event
	err = store.Set("admin:1", "SuperAdmin")
	if err != nil {
		t.Fatalf("Failed to set admin:1: %v", err)
	}
	
	select {
	case event := <-eventCh:
		t.Errorf("Unexpected event for non-matching prefix: %v", event)
	case <-time.After(200 * time.Millisecond):
		// Expected timeout
	}
}

func TestMultipleWatchers(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_multi_watch.db")
	
	store, err := kvstore.NewAt[string, string](dbPath, "test_multi")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	
	// Create multiple watchers
	ch1, cancel1 := store.Watch("key1")
	defer cancel1()
	
	ch2, cancel2 := store.Watch("key1")
	defer cancel2()
	
	ch3, cancel3 := store.WatchPrefix("key")
	defer cancel3()
	
	// Set key1 - all watchers should receive the event
	err = store.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Failed to set key1: %v", err)
	}
	
	// Check all channels received the event
	channels := []<-chan kvstore.WatchEvent[string, string]{ch1, ch2, ch3}
	for i, ch := range channels {
		select {
		case event := <-ch:
			if event.Key != "key1" || event.Value != "value1" {
				t.Errorf("Watcher %d: unexpected event: %v", i+1, event)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Watcher %d: timeout waiting for event", i+1)
		}
	}
}

func TestWatcherCancellation(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_cancel.db")
	
	store, err := kvstore.NewAt[string, string](dbPath, "test_cancel")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	
	eventCh, cancel := store.Watch("key1")
	
	// Set key before cancellation
	err = store.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	
	// Should receive the event
	select {
	case <-eventCh:
		// Expected
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event before cancellation")
	}
	
	// Cancel the watcher
	cancel()
	
	// Give some time for cancellation to take effect
	time.Sleep(100 * time.Millisecond)
	
	// Set key after cancellation
	err = store.Set("key1", "value2")
	if err != nil {
		t.Fatalf("Failed to set key after cancellation: %v", err)
	}
	
	// Should NOT receive the event (channel should be closed)
	select {
	case event, ok := <-eventCh:
		if ok {
			t.Errorf("Received unexpected event after cancellation: %v", event)
		}
		// If !ok, channel was closed which is expected
	case <-time.After(200 * time.Millisecond):
		// This is also fine - no event received
	}
}

func TestStopAllWatchers(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_stop_all.db")
	
	store, err := kvstore.NewAt[string, string](dbPath, "test_stop")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	
	// Create multiple watchers
	ch1, _ := store.Watch("key1")
	ch2, _ := store.Watch("key2")
	ch3, _ := store.WatchPrefix("key")
	
	// Stop all watchers
	store.StopAllWatchers()
	
	// Give some time for stop to take effect
	time.Sleep(100 * time.Millisecond)
	
	// Set keys - no watcher should receive events
	store.Set("key1", "value1")
	store.Set("key2", "value2")
	
	// Check that no events are received
	select {
	case event := <-ch1:
		t.Errorf("ch1 received unexpected event: %v", event)
	case event := <-ch2:
		t.Errorf("ch2 received unexpected event: %v", event)
	case event := <-ch3:
		t.Errorf("ch3 received unexpected event: %v", event)
	case <-time.After(200 * time.Millisecond):
		// Expected - no events
	}
}

func TestWatchWithCleanup(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_cleanup.db")
	
	// Test that file cleanup doesn't affect watchers
	store, err := kvstore.NewAt[string, string](dbPath, "test_cleanup")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	
	eventCh, cancel := store.Watch("key1")
	
	// Set and verify
	store.Set("key1", "value1")
	
	select {
	case <-eventCh:
		// Expected
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event")
	}
	
	// Cancel should clean up resources
	cancel()
	
	// Verify the db file still exists and is functional
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was removed unexpectedly")
	}
	
	// Store should still be functional
	val, err := store.Get("key1", "")
	if err != nil {
		t.Errorf("Failed to get key after watcher cleanup: %v", err)
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}
}
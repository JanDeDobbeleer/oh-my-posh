package cache

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	cases := []struct {
		setupFunc func() *store
		testFunc  func(t *testing.T)
		name      string
	}{
		{
			name: "Print store with data",
			setupFunc: func() *store {
				testStore := Session.new()
				testStore.cache.Set("test_key1", &Entry[any]{
					Value:     "test_value1",
					Timestamp: time.Now().Unix(),
					TTL:       3600, // 1 hour
				})
				testStore.cache.Set("test_key2", &Entry[any]{
					Value:     42,
					Timestamp: time.Now().Unix(),
					TTL:       -1, // never expires
				})
				testStore.cache.Set("expired_key", &Entry[any]{
					Value:     "expired_value",
					Timestamp: time.Now().Unix() - 7200, // 2 hours ago
					TTL:       3600,                     // 1 hour (should be expired)
				})
				session = testStore
				return testStore
			},
			testFunc: func(t *testing.T) {
				result := Print(Session)
				assert.Contains(t, result, "Key: test_key1")
				assert.Contains(t, result, `Value: "test_value1"`) // Note: quotes are included in output
				assert.Contains(t, result, "Type: string")
				assert.Contains(t, result, "Key: test_key2")
				assert.Contains(t, result, "Value: 42")
				assert.Contains(t, result, "Type: int")
				assert.Contains(t, result, "Key: expired_key [EXPIRED]")
				assert.Contains(t, result, "never expires")
				assert.Contains(t, result, "expires at")

				// Verify structure
				lines := strings.Split(result, "\n")
				assert.True(t, len(lines) > 10, "Output should have multiple lines")
			},
		},
		{
			name: "Print empty store",
			setupFunc: func() *store {
				testStore := Session.new()
				session = testStore
				return testStore
			},
			testFunc: func(t *testing.T) {
				result := Print(Session)
				assert.Contains(t, result, "Store session is empty")
			},
		},
		{
			name: "Print nil store check",
			setupFunc: func() *store {
				testStore := Session.new()
				session = testStore
				return testStore
			},
			testFunc: func(t *testing.T) {
				// Since get() always creates a store, we test empty store behavior
				result := Print(Session)
				assert.Contains(t, result, "Store session is empty")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupFunc()
			tc.testFunc(t)
		})
	}
}

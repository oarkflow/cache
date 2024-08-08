package cache

import (
	"fmt"
	"os"
	"testing"
)

func TestCacheWithPogrebAndEvcache(t *testing.T) {
	mlru, err := New[string, int](100*1024*1024, 10, "pogreb_test_db")
	if err != nil {
		t.Fatalf("Failed to create MapWithPogreb: %v", err)
	}
	defer mlru.Close()
	defer os.RemoveAll("pogreb_test_db")

	mlru.Set("one", 1)
	mlru.Set("two", 2)

	// Simulate high memory usage
	for i := 0; i < 1000; i++ {
		mlru.Set(fmt.Sprintf("key_%d", i), i)
	}

	// Check if the least used data was persisted and can be restored
	if val, ok := mlru.Get("one"); !ok || val != 1 {
		t.Fatalf("Expected 1, got %v", val)
	}
}

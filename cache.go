package cache

import (
	"runtime"

	"github.com/oarkflow/flydb"
	"github.com/oarkflow/xsync"
)

type Cache[K comparable, V any] struct {
	m      *xsync.MapOf[K, V]
	lru    *LRU[K, V]
	maxMem uint64
	client *flydb.DB[[]byte, []byte]
}

// New initializes a new Cache[] instance using MessagePack for serialization.
func New[K comparable, V any](maxMem uint64, lruCapacity int, dbPath string) (*Cache[K, V], error) {
	db, err := flydb.Open[[]byte, []byte](dbPath, nil)
	if err != nil {
		return nil, err
	}
	return &Cache[K, V]{
		m:      xsync.NewMap[K, V](),
		lru:    NewLRU[K, V](lruCapacity),
		maxMem: maxMem,
		client: db,
	}, nil
}

func (mlru *Cache[K, V]) Set(key K, value V) {
	mlru.m.Set(key, value)
	mlru.lru.Put(key, value)

	go mlru.checkMemoryAndPersist()
}

func (mlru *Cache[K, V]) Del(key K) {
	mlru.m.Del(key)
	mlru.lru.Remove(key)
	// Serialize the key and value using MessagePack
	keyBytes, err := serialize(key)
	if err != nil {
		return
	}
	mlru.client.Delete(keyBytes)
}

func (mlru *Cache[K, V]) Get(key K) (V, bool) {
	// Try to load from the LRU cache
	if value, ok := mlru.lru.Get(key); ok {
		return value, true
	}

	// Try to load from the in-memory map
	if value, ok := mlru.m.Get(key); ok {
		mlru.lru.Put(key, value)
		return value, true
	}

	// Get from store if not found in memory
	if value, ok := mlru.restoreFromClient(key); ok {
		mlru.Set(key, value) // Re-store in memory and LRU
		return value, true
	}

	var zero V
	return zero, false
}

func (mlru *Cache[K, V]) checkMemoryAndPersist() {
	if getMemoryUsage() > mlru.maxMem {
		mlru.persistLeastUsed()
	}
}

func (mlru *Cache[K, V]) persistLeastUsed() {
	key, value, ok := mlru.lru.removeOldest()
	if !ok {
		return
	}

	// Serialize the key and value using MessagePack
	keyBytes, err := serialize(key)
	if err != nil {
		return
	}
	valueBytes, err := serialize(value)
	if err != nil {
		return
	}

	// Persist to store
	if err := mlru.client.Put(keyBytes, valueBytes); err != nil {
		return
	}

	// Remove from in-memory map
	mlru.m.Del(key)
}

func (mlru *Cache[K, V]) restoreFromClient(key K) (V, bool) {
	// Serialize the key using MessagePack
	keyBytes, err := serialize(key)
	if err != nil {
		return *new(V), false
	}

	// Fetch from store
	valueBytes, err := mlru.client.Get(keyBytes)
	if err != nil || valueBytes == nil {
		return *new(V), false
	}

	// Deserialize the value using MessagePack
	value, err := deserialize[V](valueBytes)
	if err != nil {
		return *new(V), false
	}

	return value, true
}

// Close closes the store database.
func (mlru *Cache[K, V]) Close() error {
	return mlru.client.Close()
}

func getMemoryUsage() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.Alloc
}

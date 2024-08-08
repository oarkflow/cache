package cache

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] struct {
	capacity int
	cache    map[K]*list.Element
	list     *list.List
	mutex    sync.Mutex
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	return &LRU[K, V]{
		capacity: capacity,
		cache:    make(map[K]*list.Element),
		list:     list.New(),
	}
}

func (l *LRU[K, V]) Get(key K) (V, bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if ele, ok := l.cache[key]; ok {
		l.list.MoveToFront(ele)
		return ele.Value.(*entry[K, V]).value, true
	}
	var zero V
	return zero, false
}

func (l *LRU[K, V]) Put(key K, value V) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if ele, ok := l.cache[key]; ok {
		l.list.MoveToFront(ele)
		ele.Value.(*entry[K, V]).value = value
		return
	}

	ele := l.list.PushFront(&entry[K, V]{key, value})
	l.cache[key] = ele

	if l.list.Len() > l.capacity {
		l.removeOldest()
	}
}

func (l *LRU[K, V]) removeOldest() (K, V, bool) {
	var k K
	var v V
	ele := l.list.Back()
	if ele != nil {
		l.list.Remove(ele)
		kv := ele.Value.(*entry[K, V])
		delete(l.cache, kv.key)
		return kv.key, kv.value, true
	}
	return k, v, false
}

func (l *LRU[K, V]) Remove(key K) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if ele, ok := l.cache[key]; ok {
		l.list.Remove(ele)
		delete(l.cache, key)
	}
}

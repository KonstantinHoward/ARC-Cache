package test

import (
	"container/list"
	"fmt"
)

// An LRU is a fixed-size in-memory cache with least-recently-used eviction
// user gives size of cache and number of pages
type LRU struct {
	// whatever fields you want here
	max_pages     int
	page_size     int
	total_size    int
	current_pages int
	stat          Stats
	pairMap       map[any]Value
	keyQueue      *list.List
}

type Value struct {
	keySize  int
	value    []byte
	queuePos *list.Element
}

// NewLRU returns a pointer to a new LRU with a capacity to store limit bytes
func NewLru(total_cache int, num_pages int) *LRU {
	newLRU := new(LRU)
	newLRU.total_size = total_cache
	newLRU.max_pages = num_pages
	newLRU.page_size = total_cache / num_pages
	newLRU.current_pages = 0
	newLRU.stat.Hits = 0
	newLRU.stat.Misses = 0
	newLRU.pairMap = make(map[any]Value)
	newLRU.keyQueue = list.New()
	return newLRU
}

func (lru *LRU) Len() int {
	return lru.current_pages
}

func (lru *LRU) MaxPages() int {
	return lru.max_pages
}

func (lru *LRU) RemainingPages() int {
	return lru.max_pages - lru.current_pages
}

func (lru *LRU) Contains(key string) (ok bool) {
	_, ok = lru.pairMap[key]
	return ok
}

// Get returns whether or not the key was found. true for hit, false for miss.
// hits and misses are updated
func (lru *LRU) Get(key string) (value []byte, ok bool) {
	v, ok := lru.pairMap[key]
	if ok {
		// remove instance from queue then add to back. use move to back
		lru.keyQueue.MoveToBack(v.queuePos)
		lru.stat.Hits++
		return v.value, ok
	}
	lru.stat.Misses++
	return nil, ok
}

func (lru *LRU) Remove(key string) (value []byte, ok bool) {
	v, ok := lru.pairMap[key]
	if ok {
		delete(lru.pairMap, key)
		lru.keyQueue.Remove(v.queuePos)
		lru.current_pages--
		return v.value, true
	}
	return nil, false
}

// Removing oldest key from lru and returns whether or not something was evicted
func (lru *LRU) RemoveLRU() (key string, value []byte) {

	// front is oldest
	toEvict := lru.keyQueue.Front()
	v, ok := lru.pairMap[toEvict.Value]
	if toEvict == nil || !ok {
		// fmt.Println("evicting from empty list")
		return "", nil
	}

	lru.keyQueue.Remove(toEvict)

	delete(lru.pairMap, toEvict.Value)
	lru.current_pages--
	valStr := fmt.Sprintf("%v", toEvict.Value)
	return valStr, v.value
}

//this works for both replacing and adding
func (lru *LRU) Add(key string, value []byte) {
	//
	newV := new(Value)
	newV.value = value

	_ = lru.keyQueue.PushBack(key)
	newV.queuePos = lru.keyQueue.Back()
	newV.queuePos.Value = key
	lru.pairMap[key] = *newV

}

// Set associates the given value with the given key, possibly evicting values
// to make room. Returns true if the binding was added successfully, else false.
func (lru *LRU) Set(key string, value []byte) bool {

	// if size of key and value are greater than remaining space, then evict until
	// they are less than remaining space, making sure we stop evicting when cache
	// is empty
	if len(value)+len(key) > lru.page_size {
		return false
	}

	v, ok := lru.pairMap[key]

	// if it is found then we do not need to remove oldest
	if !ok {
		if lru.current_pages == lru.max_pages {
			lru.RemoveLRU()
		}
		lru.Add(key, value)
		lru.current_pages++

	} else {
		lru.keyQueue.Remove(v.queuePos)
		lru.Add(key, value)

	}

	return true
}

// Stats returns statistics about how many search hits and misses have occurred.
func (lru *LRU) Stats() *Stats {
	return &lru.stat
}

package main

type Stats struct {
	Hits   int
	Misses int
}

func (stats *Stats) Equals(other *Stats) bool {
	if stats == nil && other == nil {
		return true
	}
	if stats == nil || other == nil {
		return false
	}
	return stats.Hits == other.Hits && stats.Misses == other.Misses
}

type Cache interface {
	// MaxStorage returns the maximum number of pages a cache can store
	MaxPages() int

	// RemainingStorage returns the number of unused pages available in this cache
	RemainingPages() int

	// Get returns the value associated with the given key, if it exists.
	// This operation counts as a "use" for that key-value pair
	// ok is true if a value was found and false otherwise.
	Get(key string) (value []byte, ok bool)

	// Set associates the given value with the given key in the cache and follows the cache's
	// eviction protocol if the cache is full.
	Set(key string, value []byte) bool

	// Len returns the number of pages in the cache.
	Len() int

	// Stats returns a pointer to a Stats object that indicates how many hits
	// and misses this cache has resolved over its lifetime.
	Stats() *Stats
}

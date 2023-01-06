package test

type ARC struct {
	num_pages int // Total Number of pages in the cache
	pages_used int
	num_bytes int // Total Number of bytes in the cache
	bytes_per_page int // Allowed bytes in each page of memory (floor of bytes/pages)
	p int // Adaptive parameter

	t1 *LRU // T1 = LRU for recently accessed data
	b1 *LRU // B1 = LRU for data evicted from T1

	t2 *LRU // T2 = LRU for frequently accessed data
	b2 *LRU // B2 = LRU for data evicted from T2

	hits int // Number of hits on the cache
	misses int // Number of misses on the cache
}

func NewARC(limit int, pages int) *ARC {
	t1 := NewLru(limit, pages)
	t2 := NewLru(limit, pages)
	b1 := NewLru(limit, pages)
	b2 := NewLru(limit, pages)

	return &ARC{
		num_pages: pages,
		pages_used: 0,
		num_bytes: limit,
		bytes_per_page: limit/pages,
		p: 0,
		t1: t1,
		b1: b1,
		t2: t2,
		b2: b2,
		hits: 0,
		misses: 0,
	}
}

// MaxPages returns the number of pages that are available in the cache total
// regardless of how many have been used or are empty.
func (arc *ARC) MaxPages() int {
	return arc.num_pages
}

// Remaining pages tells how many pages are remaining in the cache that are
// still empty.
func (arc *ARC) RemainingPages() int {
	return arc.num_pages - arc.pages_used
}

// Get returns the value of the key if it is in t1 or t2.
// It also returns a boolean if the key is in the cache or not
func (arc *ARC) Get(key string) (value []byte, ok bool) {
	if arc.t1.Contains(key){
		v, ok := arc.t1.Remove(key)
		if !ok {
			return nil, false
		}
		arc.t2.Set(key, v)
		arc.hits += 1
		return v, true
	}
	if arc.t2.Contains(key){
		v, ok := arc.t2.Get(key)
		if !ok {
			return nil, false
		}
		arc.hits += 1
		return v, true
	}
	arc.misses += 1
	return nil, false
}

func (arc *ARC) Set(key string, value []byte) {
	if len(key) + len(value) > arc.bytes_per_page{
		return
	}
	// CASE 1
	if arc.t1.Contains(key){
		arc.t1.Remove(key)
		arc.t2.Set(key, value)
		return
	}
	if arc.t2.Contains(key){
		arc.t2.Set(key, value)
		return
	}

	// CASE 2
	if arc.b1.Contains(key){
		delta := 1
		if arc.b1.Len() < arc.b2.Len(){
			delta = arc.b2.Len() / arc.b1.Len()
		}
		if arc.p + delta > arc.num_pages{
			arc.p = arc.num_pages
		} else{
			arc.p += delta
		}
		arc.Replace(key)
		arc.b1.Remove(key)
		arc.t2.Set(key, value)
		arc.pages_used += 1
		return
	}

	// CASE 3
	if arc.b2.Contains(key){
		delta := 1
		if arc.b2.Len() < arc.b1.Len(){
			delta = arc.b1.Len() / arc.b2.Len()
		}
		if arc.p - delta < 0{
			arc.p = 0
		} else{
			arc.p -= delta
		}
		arc.Replace(key)
		arc.b2.Remove(key)
		arc.t2.Set(key, value)
		arc.pages_used += 1
		return
	}

	// CASE 4
	// CASE 4.A
	if arc.t1.Len() + arc.b1.Len() == arc.num_pages{
		if arc.t1.Len() < arc.num_pages{
			arc.b1.RemoveLRU()
			arc.Replace(key)
		} else{
			arc.t1.RemoveLRU()
			arc.pages_used -= 1
		}
	} else if arc.t1.Len() + arc.b1.Len() < arc.num_pages{
		if arc.t1.Len() + arc.t2.Len() + arc.b1.Len() + arc.b2.Len() >= arc.num_pages{
			if arc.t1.Len() + arc.t2.Len() + arc.b1.Len() + arc.b2.Len() == arc.num_pages * 2{
				arc.b2.RemoveLRU()
			}
			arc.Replace(key)
		}
	}
	arc.t1.Set(key, value)
	arc.pages_used += 1
}

// Replace function from ARC research paper
func (arc *ARC) Replace(key string){
	if arc.t1.Len() > 0 && (arc.t1.Len() > arc.p || (arc.b2.Contains(key) && arc.t1.Len() == arc.p)){
		rkey, rval := arc.t1.RemoveLRU()
		arc.b1.Set(rkey, rval)
	} else{
		rkey, rval := arc.t2.RemoveLRU()
		arc.b2.Set(rkey, rval)
	}
	arc.pages_used -= 1
}

// Len returns the number of pages that have been used in the cache
func (arc *ARC) Len() int {
	return arc.t1.Len() + arc.t2.Len()
}

// Stats returns statistics about how many search hits and misses have occurred.
func (arc *ARC) Stats() *Stats {
	return &Stats{
		Hits: arc.hits,
		Misses: arc.misses,
	}
}
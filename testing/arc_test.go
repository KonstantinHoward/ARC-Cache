/******************************************************************************
 * arc_test.go
 * Author:
 * Usage:    `go test`  or  `go test -v`
 * Description:
 *    An unit testing suite for arc.go.
 ******************************************************************************/

package test

import (
	"fmt"
	"testing"
)

/******************************************************************************/
/*                                Constants                                   */
/******************************************************************************/

const cap = 128
const p = 8

/******************************************************************************/
/*                                  Tests                                     */
/******************************************************************************/

// Checks properties of an empty cache
func TestARCEmpty(t *testing.T) {
	capacity := 64
	arc := NewARC(capacity, p)

	// Empty
	if arc.Len() != 0{
		t.Errorf("Failed to create an empty cache. Length is: %d", arc.Len())
		t.FailNow()
	}

	capacity = 256
	arc = NewARC(capacity, p)

	// Empty
	if arc.Len() != 0{
		t.Errorf("Failed to create an empty cache. Length is: %d", arc.Len())
		t.FailNow()
	}

	// Empty
	capacity = 1024
	arc = NewARC(capacity, p)

	if arc.Len() != 0{
		t.Errorf("Failed to create an empty cache. Length is: %d", arc.Len())
		t.FailNow()
	}
}

// Checks proper get on an empty cache
func TestARCEmptyGet(t *testing.T) {
	arc := NewARC(cap, p)

	// Empty
	if arc.Len() != 0{
		t.Errorf("Failed to create an empty cache. Length is: %d", arc.Len())
		t.FailNow()
	}

	// Check no return on a random key
	key := "key"
	_, ok := arc.Get(key)
	if ok {
		t.Errorf("Failed to cache miss on an empty cache. Key is: %s", key)
		t.FailNow()
	}
}

// Checks that a single binding can be added to the cache
func TestARCSingleBinding(t *testing.T) {
	arc := NewARC(cap, p)

	// Add a key
	key := "key"
	value := []byte(key)
	arc.Set(key, value)

	// Make sure the key gets a cache hit
	val, ok := arc.Get(key)
	if !ok {
		t.Errorf("Failed to cache hit on a set key. Key is: %s", key)
		t.FailNow()
	}
	if !bytesEqual(value, val){
		t.Errorf("Failed to set the correct value to a key. Key is: %s. Value is: %s. Value should be: %s", key, val, value)
		t.FailNow()
	}

	// Update length
	if arc.Len() != 1{
		t.Errorf("Failed to update cache length. Length is: %d when it should be 1", arc.Len())
		t.FailNow()
	}
}

// Checks that all page bindings can be filled and used
func TestARCFillBindings(t *testing.T) {
	arc := NewARC(cap, p)

	// Fill keys 1-8
	for i := 1; i <= 8; i++ {
		key := fmt.Sprintf("key%d", i)
		val := []byte(key)
		arc.Set(key, val)

		// Test correct values from get
		res, _ := arc.Get(key)
		if !bytesEqual(res, val){
			t.Errorf("Wrong value: %s for binding with key: %s", res, key)
			t.FailNow()
		}

		// Check length updates
		if arc.Len() != i {
			t.Errorf("Failed to update cache length. Length is: %d when it should be %d", arc.Len(), i)
			t.FailNow()
		}
	}
}

// Checks that a binding that is too large for the given page size cannot be added
func TestARCBindingTooLarge(t *testing.T) {
	arc := NewARC(cap, p)

	// Too large of a page binding
	key := "keytoobig"
	value := []byte(key)
	arc.Set(key, value)

	// Cannot get a key if the set was too large of a page cache miss
	val, ok := arc.Get(key)
	if ok {
		t.Errorf("Failed to cache miss on an empty cache. Key is: %s", key)
		t.FailNow()
	}
	if val != nil{
		t.Errorf("Failed to return nil for no value. Value is: %s", val)
		t.FailNow()
	}
}

// Checks that overfilling a cache does not add more space and that
// hits and misses are proper.
func TestARCOverFill(t *testing.T) {
	arc := NewARC(cap, p)

	// Add keys 1-9 with 8 pages to test a removal of the first key
	for i := 1; i <= 9; i++ {
		key := fmt.Sprintf("key%d", i)
		val := []byte(key)
		arc.Set(key, val)

		res, _ := arc.Get(key)
		if !bytesEqual(res, val){
			t.Errorf("Wrong value: %s for binding with key: %s", res, key)
			t.FailNow()
		}
		if i < 9{
			if arc.Len() != i {
				t.Errorf("Failed to update cache length. Length is: %d when it should be %d", arc.Len(), i)
				t.FailNow()
			}
		}
	}

	// Removed key
	testkey := "key1"
	val, ok := arc.Get(testkey)

	// Does not Get
	if ok {
		t.Errorf("Failed to cache miss on a removed key. Key is: %s", testkey)
		t.FailNow()
	}

	// Removes key no longer exists
	if val != nil{
		t.Errorf("Failed to return nil for cache miss. Value is: %s", val)
		t.FailNow()
	}

	// Check that length does no exceed pages
	if arc.Len() != 8 {
		t.Errorf("Overfilled the cache. Length is: %d when it should be %d", arc.Len(), 8)
		t.FailNow()
	}
}

// Checks that keys in T1 will be moved to T2 on reaccess
func TestARCCaseOne(t *testing.T) {
	arc := NewARC(cap, p)

	addKeys(arc, 1, 8)

	for i := 1; i <= 8; i++ {
		key := fmt.Sprintf("key%d", i)
		val := []byte(key)
		arc.Set(key, val)

		if arc.t2.Len() != i {
			t.Errorf("Failed to update MRU cache length. Length is: %d when it should be %d", arc.t2.Len(), i)
			t.FailNow()
		}
	}

	// Make sure the cache stays full
	if arc.Len() != 8{
		t.Errorf("Failed to keep cache length. Length is: %d when it should be %d", arc.Len(), 8)
		t.FailNow()
	}

	// Check that inner cache lengths are correct
	checkLengths(arc, t, 0, 8, 0, 0)
}

// Checks that ghost list B1 gets populated on deletion
func TestARCPopulateB1(t *testing.T) {
	arc := NewARC(cap, p)

	// Add 1-8
	addKeys(arc, 1, 8)
	// Add 1-4 again moving to frequent LRU
	addKeys(arc, 1, 4)

	// Move key5 and key6 to B1
	for i := 9; i <= 10; i++ {
		key := fmt.Sprintf("key%d", i)
		val := []byte(key)
		arc.Set(key, val)
	}

	// Check B1 ghost cache length is updated
	if arc.b1.Len() != 2{
		t.Errorf("Failed to update ghost cache length. Length is: %d when it should be %d", arc.b1.Len(), 2)
		t.FailNow()
	}
}

// Checks that Case 2 (values in B1) move to the correct spot
// And that accessing in the middle is correct
func TestARCCase2(t *testing.T) {
	arc := NewARC(cap, p)

	// Add 1-8
	addKeys(arc, 1, 8)
	// Add 1-4 again moving to frequent LRU
	addKeys(arc, 1, 4)

	// Move key5 and key6 to B1
	for i := 9; i <= 10; i++ {
		key := fmt.Sprintf("key%d", i)
		val := []byte(key)
		arc.Set(key, val)
	}

	// Cache miss
	_, ok := arc.Get("key5")

	if ok{
		t.Errorf("Failed to cache miss on a removed key. Key is: %s", "key5")
		t.FailNow()
	}

	// Cache miss
	_, ok = arc.Get("key6")

	if ok{
		t.Errorf("Failed to cache miss on a removed key. Key is: %s", "key5")
		t.FailNow()
	}

	// Move out of B1 update parameter p twice
	for i := 5; i <= 6; i++ {
		key := fmt.Sprintf("key%d", i)
		val := []byte(key)
		arc.Set(key, val)
	}

	// Check parameter update
	if arc.p != 2{
		t.Errorf("Failed to update parameter. P is: %d when it should be %d", arc.p, 1)
		t.FailNow()
	}

	// Check that 7 and 8 are pushed off of T1
	for i := 7; i <= 8; i++ {
		key := fmt.Sprintf("key%d", i)
		_, ok = arc.Get(key)
		if ok{
			t.Errorf("Failed to cache miss on a removed key. Key is: %s", key)
			t.FailNow()
		}
	}

	// Check that inner cache lengths are correct
	checkLengths(arc, t, 2, 6, 2, 0)
}

// Checking updates in Case 3 of ARC paper
func TestARCCaseThree(t *testing.T) {
	arc := NewARC(cap, p)
	
	// Add 1-8 to T1
	addKeys(arc, 1, 8)
	// Push 1 to T2
	addKeys(arc, 1, 1)
	// Add 9 to T1 push LRU (2) off T1
	addKeys(arc, 9, 9)
	// Add 2-5 to T2 from B1 and push LRU of T1 off
	// Update p all the way to 4
	addKeys(arc, 2, 5)
	// Add 6 to T2 from B1 and push LRU of T2 off
	addKeys(arc, 6, 6)
	// 1 should be in B2. Add from B2 meaning update p to be
	// equal to 3 and replace by pushing LRU off T1 (7)
	addKeys(arc, 1, 1)

	if arc.p != 3{
		t.Errorf("Failed to continuously update parameter for Case 3. P is: %d", arc.p)
		t.FailNow()
	}
	// Check that inner cache lengths are correct
	checkLengths(arc, t, 2, 6, 1, 0)

	// 7 is the only key that is now pushed off the cache
	for i := 1; i <= 9; i++ {
		key := fmt.Sprintf("key%d", i)
		_, ok := arc.Get(key)

		if i != 7{
			if !ok{
				t.Errorf("Failed to cache hit on a proper key. Key is: %s", key)
				t.FailNow()
			}
		} else{
			if ok{
				t.Errorf("Failed to cache miss on a removed key. Key is: %s", key)
				t.FailNow()
			}
		}
	}
}

// Checks updating in Case 4A of the ARC paper
func TestARCCaseFourA(t *testing.T) {
	arc := NewARC(cap, p)
	
	// 8 Items in T1
	addKeys(arc, 1, 8)
	// 2 Items pushed to T2
	addKeys(arc, 1, 2)
	// Move two items into B1 (key3 and key4)
	addKeys(arc, 9, 10)
	// Should add items to T1 and evict out of B1 such that |T1|=6 and |B1|=2
	addKeys(arc, 11, 12)

	// Check that inner cache lengths are correct
	checkLengths(arc, t, 6, 2, 2, 0)
}

// Checks updating in Case 4B of the ARC paper
func TestARCCaseFourB(t *testing.T) {
	arc := NewARC(cap, p)
	
	// 8 Items in T1
	addKeys(arc, 1, 8)
	// 1 Item in T2
	addKeys(arc, 1, 1)
	// Move an item into B1 and push off of T1
	addKeys(arc, 9, 9)
	// Find items in B1 and update p
	addKeys(arc, 2, 5)
	// Push off T2
	addKeys(arc, 6, 8)
	// Add keys up to p on T1
	addKeys(arc, 10, 12)

	// Check that inner cache lengths are correct
	checkLengths(arc, t, 4, 4, 0, 4)
	// Check p is updated correctly
	if arc.p != 4{
		t.Errorf("Failed to continuously update parameter after 4B. P is: %d", arc.p)
		t.FailNow()
	}
}

// Checks a large number of operations when ghost lists are not used
func TestARCLargeNumOpsNoGhost(t *testing.T){
	arc := NewARC(cap, p)

	addKeys(arc, 1, 200)

	// Only the last 8 keys should remain
	for i := 193; i <= 200; i++ {
		key := fmt.Sprintf("key%d", i)
		value := []byte(key)
		val, ok := arc.Get(key)

		if !ok{
			t.Errorf("Failed to cache hit on a set key. Key is: %s", key)
			t.FailNow()
		}

		if !bytesEqual(value, val){
			t.Errorf("Wrong value: %s for binding with key: %s", val, key)
			t.FailNow()
		}
	}

	// Check that inner cache lengths are correct
	checkLengths(arc, t, 0, 8, 0, 0)
}

// Checks that a short workload will return the correct length of caches
func TestARCSmallNumOpsWithGhost(t *testing.T){
	arc := NewARC(cap, p)

	// Add 1-8 to T1
	addKeys(arc, 1, 8)
	// Move 1-7 to T2
	addKeys(arc, 1, 7)
	// Add 9 to T1 move LRU from T1 to B1
	addKeys(arc, 9, 9)
	// Add back to T2 from B1 increase p on 1-2
	addKeys(arc, 1, 4)
	// Fill T1 and move of of T2
	addKeys(arc, 10, 12)
	// Keys move back to T2 from ghost lists
	addKeys(arc, 5, 9)
	// Move off T2 for 13 add to T1/move off T1 for 14-16
	addKeys(arc, 13, 16)

	// Check that inner cache lengths are correct
	checkLengths(arc, t, 3, 5, 4, 4)
	// Check that p is updated correctly
	if arc.p != 2{
		t.Errorf("Failed to continuously update parameter. P is: %d", arc.p)
		t.FailNow()
	}
}

// Tests that a key is updated and not readded
func TestARCKeyUpdate(t *testing.T){
	arc := NewARC(cap, p)

	// Add key 1
	addKeys(arc, 1, 1)
	key := fmt.Sprintf("key%d", 1)
	value := []byte(key)
	val, ok := arc.Get(key)

	// Should hit
	if !ok{
		t.Errorf("Failed to cache hit on a set key. Key is: %s", key)
		t.FailNow()
	}

	if !bytesEqual(value, val){
		t.Errorf("Wrong value: %s for binding with key: %s", val, key)
		t.FailNow()
	}

	// Add key 1 with new value
	value = []byte("key2")
	arc.Set(key, value)
	val, ok = arc.Get(key)

	// Should hit with new value
	if !ok{
		t.Errorf("Failed to cache hit on a set key. Key is: %s", key)
		t.FailNow()
	}

	if !bytesEqual(value, val){
		t.Errorf("Wrong value: %s for binding with key: %s", val, key)
		t.FailNow()
	}

	// Check that inner cache lengths are correct
	checkLengths(arc, t, 0, 1, 0, 0)
}

// Tests that the max num of pages is correct
func TestARCMaxPages(t *testing.T){
	arc := NewARC(cap, p)

	max := arc.MaxPages()

	// Max pages should always be 8
	if max != 8{
		t.Errorf("Wrong value: %d for remaining pages. Value should be: %d", max, 8)
		t.FailNow()
	}
}

// Tests that remaining pages updates properly
func TestARCRemainingPages(t *testing.T){
	arc := NewARC(cap, p)

	max := arc.MaxPages()

	// Check that remaining pages slowly goes down until the cache is full and stays at 0
	for i := 1; i <= 16; i++ {
		key := fmt.Sprintf("key%d", i)
		val := []byte(key)
		arc.Set(key, val)

		if i <= 8{
			if arc.RemainingPages() != max-i{
				t.Errorf("Wrong value: %d for remaining pages. Value should be: %d", arc.RemainingPages(), max-i)
				t.FailNow()
			}
		} else{
			if arc.RemainingPages() != 0{
				t.Errorf("Wrong value: %d for remaining pages. Value should be: %d", arc.RemainingPages(), 0)
				t.FailNow()
			}
		}
	}
}

func SimpleStats(t *testing.T){
	arc := NewARC(cap, p)

	addKeys(arc, 1, 8)

	// Check that remaining pages slowly goes down until the cache is full and stays at 0
	for i := 1; i <= 8; i++ {
		key := fmt.Sprintf("key%d", i)
		arc.Get(key)
	}

	stats := arc.Stats()

	if !stats.Equals(&Stats{Hits: 8, Misses:0,}){
		t.Errorf("Hits are %d. Should be: %d", arc.hits, 8)
		t.Errorf("Misses are %d. Should be: %d", arc.misses, 0)
		t.FailNow()
	}
}

func LongerStats(t *testing.T){
	arc := NewARC(cap, p)

	// Case From 4B
	addKeys(arc, 1, 8)
	addKeys(arc, 1, 7)
	for i := 1; i <= 7; i++ {
		key := fmt.Sprintf("key%d", i)
		arc.Get(key)
	}
	addKeys(arc, 9, 9)
	for i := 2; i <= 2; i++ {
		key := fmt.Sprintf("key%d", i)
		arc.Get(key)
	}
	addKeys(arc, 1, 4)
	addKeys(arc, 10, 12)
	addKeys(arc, 5, 9)
	for i := 5; i <= 9; i++ {
		key := fmt.Sprintf("key%d", i)
		arc.Get(key)
	}
	addKeys(arc, 13, 16)

	// Check that remaining pages slowly goes down until the cache is full and stays at 0
	for i := 1; i <= 16; i++ {
		key := fmt.Sprintf("key%d", i)
		arc.Get(key)
	}

	stats := arc.Stats()

	if !stats.Equals(&Stats{Hits: 20, Misses:9,}){
		t.Errorf("Hits are %d. Should be: %d", arc.hits, 20)
		t.Errorf("Misses are %d. Should be: %d", arc.misses, 9)
		t.FailNow()
	}
}
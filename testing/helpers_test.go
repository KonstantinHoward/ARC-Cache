package test

import (
	"testing"
	"fmt"
)

/******************************************************************************/
/*                                 Helpers                                    */
/******************************************************************************/

// CacheType returns a string representing the type (i.e. eviction scheme) of
// this cache.

// Returns true iff a and b represent equal slices of bytes.
func bytesEqual(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}

	for i, v := range a {
		if b[i] != v {
			return false
		}
	}

	return true
}

func addKeys(arc *ARC, start int, end int){
	for i := start; i <= end; i++ {
		key := fmt.Sprintf("key%d", i)
		val := []byte(key)
		arc.Set(key, val)
	}
}

func checkLengths(arc *ARC, t *testing.T, t1 int, t2 int, b1 int, b2 int){
	if arc.t1.Len() != t1{
		t.Errorf("Failed to not update LRU length. Length is: %d when it should be %d", arc.t1.Len(), t1)
		t.FailNow()
	}
	if arc.t2.Len() != t2{
		t.Errorf("Failed to not update LFU length. Length is: %d when it should be %d", arc.t2.Len(), t2)
		t.FailNow()
	}
	if arc.b1.Len() != b1{
		t.Errorf("Failed to update B1 ghost cache length. Length is: %d when it should be %d", arc.b1.Len(), b1)
		t.FailNow()
	}
	if arc.b2.Len() != b2{
		t.Errorf("Failed to update B2 ghost cache length. Length is: %d when it should be %d", arc.b2.Len(), b2)
		t.FailNow()
	}
}
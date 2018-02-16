package server

import (
	"testing"
	"time"
)

// cleanUpCache is just a helper function in case we decide to change our cache implementation
func cleanUpCache() {
	opCache.Flush()
}

// TestRetrieveFromCache adds values to the cache manuall and then uses retrieveFromCache() to get
// them back from the cache
func TestRetrieveFromCache(t *testing.T) {
	cleanUpCache()
	t.Run("get cached", retrieveBeforeExpire)
	cleanUpCache()
	t.Run("get expired", retrieveAfterExpire)
	cleanUpCache()
}

func retrieveBeforeExpire(t *testing.T) {
	op := "*"
	x, y := -64.5227, 8.640
	expectedAns := -557.476128

	// value of  defaultCacheExpiration is set in cache.go (as of v0.2.0)
	opCache.Add(createCacheKey(op, x, y), expectedAns, defaultCacheExpiration)

	actualAns, inCache := retrieveFromCache(op, x, y)
	if !inCache {
		// should be in the cache
		t.Logf("unexpected inCache value: (actual %t != expected true)\n", inCache)
		t.Fail()
	}

	if actualAns != expectedAns {
		t.Logf("unexpected result from cache: (actual %f != expected %f)\n", actualAns, expectedAns)
		t.Fail()
	}
}

func retrieveAfterExpire(t *testing.T) {
	op := "*"
	x, y := -64.5227, 8.640
	expectedAns := 0.0
	providedAns := -557.476128

	expirationDuration := time.Millisecond * 50
	opCache.Add(createCacheKey(op, x, y), providedAns, expirationDuration)

	time.Sleep(expirationDuration + (time.Millisecond * 5)) // wait until the cache value expires
	actualAns, inCache := retrieveFromCache(op, x, y)
	if inCache {
		// shouldn't be in the cache
		t.Logf("unexpected inCache value: (actual %t != expected false)\n", inCache)
		t.Fail()
	}

	if actualAns != expectedAns {
		t.Logf("unexpected result from cache: (actual %f != expected %f)\n", actualAns, expectedAns)
		t.Fail()
	}
}

// TestAddToCache uses addToCache() to add values to the cache and then retrieves them manually
func TestAddToCache(t *testing.T) {
	cleanUpCache()

	op := "-"
	x, y := 9.5, -11.436
	expectedAns := 20.936

	addToCache(op, x, y, expectedAns)

	// NOTE: manual key retrieval will need to change if we update how addToCache() generates key values
	actualAns, inCache := opCache.Get(createCacheKey(op, x, y))
	if !inCache {
		// should be in cache
		t.Logf("unexpected inCache value: (actual %t != expected true)\n", inCache)
	}

	if actualAns != expectedAns {
		t.Logf("unexpected result from cache: (actual %f != expected %f)\n", actualAns, expectedAns)
		t.Fail()
	}
}

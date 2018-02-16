package server

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

// These functions are mostly helper functions that will make it easy to change our underlying cache
// implementation or to ease the transition if we decide that cache related operations belong in their
// own package

const defaultCacheExpiration time.Duration = time.Minute
const defaultCacheCleanUp time.Duration = time.Minute * 5

// cacheExpiration allows use to directly change cache expiration time while testing (saves me a few minutes)
var cacheExpiration = defaultCacheExpiration

// opCache stores all operation answers as interfaces with a timeout of one minute
var opCache *cache.Cache

func init() {
	opCache = cache.New(defaultCacheExpiration, defaultCacheCleanUp)
}

// retrieveFromCache checks to see if the math operation defined by the arguments has been performed
// in the last minute and returns the cached answer and true if it has.  If not, it returns 0 and false
func retrieveFromCache(op string, x, y float64) (float64, bool) {
	ans, inCache := opCache.Get(createCacheKey(op, x, y))
	if inCache {
		return ans.(float64), true
	}

	return 0, false
}

// addToCache adds the math operation defined by the arguments to the cache and begins the one minute
// countdown until it is removed from the cache
func addToCache(op string, x, y, ans float64) {
	opCache.Set(createCacheKey(op, x, y), ans, cacheExpiration)
}

// createCacheKey just puts op, x, and y into infix notation and formats it as a string
// FIXME: if we ever need reverse lookup or start dealing with more than two vars, we'll need a new process
func createCacheKey(op string, x, y float64) string {
	return fmt.Sprintf("%f%s%f", x, op, y)
}

package server

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

const defaultCacheExpiration time.Duration = time.Minute
const defaultCacheCleanUp time.Duration = time.Minute * 5

// opCache stores all operation answers as interfaces with a timeout of one minute
var opCache *cache.Cache

func init() {
	opCache = cache.New(defaultCacheExpiration, defaultCacheCleanUp)
}

// RetrieveFromCache checks to see if the math operation defined by the arguments has been performed
// in the last minute and returns the cached answer and true if it has.  If not, it returns 0 and false
func RetrieveFromCache(op string, x, y float64) (float64, bool) {
	ans, inCache := opCache.Get(createCacheKey(op, x, y))
	if inCache {
		return ans.(float64), true
	} else {
		return 0, false
	}
}

// AddToCache adds the math operation defined by the arguments to the cache and begins the one minute
// countdown until it is removed from the cache
func AddToCache(op string, x, y, ans float64) {
	opCache.Set(createCacheKey(op, x, y), ans, cache.DefaultExpiration)
}

// createCacheKey just puts op, x, and y into infix notation and formats it as a string
// FIXME: if we ever need to go from cache key to op, x, and y we'll need a different format
func createCacheKey(op string, x, y float64) string {
	return fmt.Sprintf("%f%s%f", x, op, y)
}

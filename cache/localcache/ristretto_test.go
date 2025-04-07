package localcache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRistretto(t *testing.T) {
	cache, err := NewRistrettoStringCache()
	if err != nil {
		return
	}
	// set a value with a cost of 1
	cache.Set("key", "value", 1)

	// wait for value to pass through buffers
	cache.Wait()

	// get value from dbcache
	value, found := cache.Get("key")
	if !found {
		panic("missing value")
	}
	fmt.Println(value)

	// del value from dbcache
	cache.Del("key")
	assert.Equal(t, nil, err)
}

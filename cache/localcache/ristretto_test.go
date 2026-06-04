package localcache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRistretto(t *testing.T) {
	cache, err := NewRistrettoStringCache()
	if err != nil {
		t.Fatal(err)
	}
	// set a value with a cost of 1
	cache.Set("key", "value", 1)

	// wait for value to pass through buffers
	cache.Wait()

	// get value from dbcache
	value, found := cache.Get("key")
	if !found {
		t.Fatal("missing value")
	}
	fmt.Println(value)

	// del value from dbcache
	cache.Del("key")
	assert.Equal(t, nil, err)
}

func TestNewRistrettoAnyCache(t *testing.T) {
	cache, err := NewRistrettoAnyCache()
	if err != nil {
		t.Fatal(err)
	}
	cache.Set("key", map[string]int{"value": 1}, 1)
	cache.Wait()

	value, found := cache.Get("key")
	if !found {
		t.Fatal("missing value")
	}
	got, ok := value.(map[string]int)
	if !ok || got["value"] != 1 {
		t.Fatalf("unexpected value: %#v", value)
	}
}

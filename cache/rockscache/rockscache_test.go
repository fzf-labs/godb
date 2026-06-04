package rockscache

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRocksCacheClientOptions(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})

	weak := NewWeakRocksCacheClient(rdb)
	if weak.Options.StrongConsistency {
		t.Fatal("weak client should disable strong consistency")
	}
	if weak.Options.DisableCacheRead || weak.Options.DisableCacheDelete {
		t.Fatal("weak client should keep cache read and delete enabled")
	}
	if weak.Options.Delay != 100*time.Millisecond {
		t.Fatalf("unexpected weak delay: %s", weak.Options.Delay)
	}
	if weak.Options.EmptyExpire != 120*time.Second {
		t.Fatalf("unexpected weak empty expire: %s", weak.Options.EmptyExpire)
	}

	strong := NewStrongRocksCacheClient(rdb)
	if !strong.Options.StrongConsistency {
		t.Fatal("strong client should enable strong consistency")
	}
	if strong.Options.Delay != 100*time.Millisecond {
		t.Fatalf("unexpected strong delay: %s", strong.Options.Delay)
	}
	if strong.Options.EmptyExpire != 120*time.Second {
		t.Fatalf("unexpected strong empty expire: %s", strong.Options.EmptyExpire)
	}
}

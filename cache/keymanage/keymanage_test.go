package keymanage

import (
	"strings"
	"testing"
	"time"
)

func TestKeyManageAddKeyDocumentAndTTL(t *testing.T) {
	manager := New("svc")
	prefix, err := manager.AddKey("user", 90*time.Second, "user cache")
	if err != nil {
		t.Fatalf("add key: %v", err)
	}
	if prefix.ServerName != "svc" || prefix.PrefixName != "user" {
		t.Fatalf("unexpected prefix: %#v", prefix)
	}
	if _, err := manager.AddKey("user", time.Second, "duplicate"); err == nil {
		t.Fatal("expected duplicate prefix error")
	}

	doc := manager.Document()
	for _, want := range []string{"|ServerName|PrefixName|ttl(s)|Remark", "|svc|user|90|user cache|"} {
		if !strings.Contains(doc, want) {
			t.Fatalf("expected %q in document:\n%s", want, doc)
		}
	}
	if prefix.TTL() != 90*time.Second {
		t.Fatalf("unexpected ttl: %v", prefix.TTL())
	}
	if prefix.TTLSecond() != 90 {
		t.Fatalf("unexpected ttl seconds: %d", prefix.TTLSecond())
	}
	if got := prefix.Keys([]string{"1", "2"}); len(got) != 2 || got[0] != "svc:user:1" || got[1] != "svc:user:2" {
		t.Fatalf("unexpected keys: %#v", got)
	}
	if got := prefix.Keys(nil); len(got) != 0 {
		t.Fatalf("nil keys should return empty slice, got %#v", got)
	}
}

func TestKeyManageDocumentSortsPrefixes(t *testing.T) {
	manager := New("svc")
	for _, prefix := range []string{"user", "admin", "order"} {
		if _, err := manager.AddKey(prefix, time.Second, prefix+" cache"); err != nil {
			t.Fatalf("add key %s: %v", prefix, err)
		}
	}

	doc := manager.Document()
	admin := strings.Index(doc, "|svc|admin|1|admin cache|")
	order := strings.Index(doc, "|svc|order|1|order cache|")
	user := strings.Index(doc, "|svc|user|1|user cache|")
	if admin == -1 || order == -1 || user == -1 {
		t.Fatalf("document missing expected rows:\n%s", doc)
	}
	if !(admin < order && order < user) {
		t.Fatalf("document rows are not sorted by prefix:\n%s", doc)
	}
}

func TestKeyPrefix_KeyEscapesSegments(t *testing.T) {
	prefix := &KeyPrefix{
		ServerName: "svc",
		PrefixName: "user",
	}
	got := prefix.Key("a:b", `c\d`)
	if got != `svc:user:a\:b:c\\d` {
		t.Fatalf("unexpected key: %s", got)
	}
}

func TestKeyPrefix_KeyAvoidsColonCollisions(t *testing.T) {
	prefix := &KeyPrefix{
		ServerName: "svc",
		PrefixName: "user",
	}
	got1 := prefix.Key("a:b", "c")
	got2 := prefix.Key("a", "b:c")
	if got1 == got2 {
		t.Fatalf("expected different keys, got identical values %q", got1)
	}
}

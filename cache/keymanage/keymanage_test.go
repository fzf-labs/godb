package keymanage

import "testing"

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

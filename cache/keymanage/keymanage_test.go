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

func TestKeyManageTrimsAndRejectsBlankInputs(t *testing.T) {
	manager := New(" svc ")
	prefix, err := manager.AddKey(" user ", time.Second, "user cache")
	if err != nil {
		t.Fatalf("add key: %v", err)
	}
	if prefix.ServerName != "svc" || prefix.PrefixName != "user" {
		t.Fatalf("unexpected trimmed prefix: %#v", prefix)
	}

	if _, err := manager.AddKey("  ", time.Second, "blank"); err == nil || !strings.Contains(err.Error(), "prefix cannot be empty") {
		t.Fatalf("expected blank prefix error, got %v", err)
	}

	if _, err := New("  ").AddKey("user", time.Second, "blank"); err == nil || !strings.Contains(err.Error(), "server name cannot be empty") {
		t.Fatalf("expected blank server error, got %v", err)
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

func TestKeyManageNilReceiverDoesNotPanic(t *testing.T) {
	var manager *KeyManage
	if _, err := manager.AddKey("user", time.Second, "user cache"); err == nil || !strings.Contains(err.Error(), "key manager cannot be nil") {
		t.Fatalf("expected nil manager add error, got %v", err)
	}
	if got := manager.Document(); got != "" {
		t.Fatalf("nil manager document should be empty, got %q", got)
	}
}

func TestKeyManageDocumentEscapesMarkdownCells(t *testing.T) {
	manager := New("svc|api")
	if _, err := manager.AddKey(`user\cache`, time.Second, `remark|with\slash`); err != nil {
		t.Fatalf("add key: %v", err)
	}

	doc := manager.Document()
	want := `|svc\|api|user\\cache|1|remark\|with\\slash|`
	if !strings.Contains(doc, want) {
		t.Fatalf("expected escaped row %q in document:\n%s", want, doc)
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

func TestKeyPrefixNilReceiverMethodsReturnZeroValues(t *testing.T) {
	var prefix *KeyPrefix
	if got := prefix.Key("1"); got != "" {
		t.Fatalf("nil key should be empty, got %q", got)
	}
	if got := prefix.Keys([]string{"1"}); len(got) != 0 {
		t.Fatalf("nil keys should be empty, got %#v", got)
	}
	if got := prefix.TTL(); got != 0 {
		t.Fatalf("nil ttl should be zero, got %v", got)
	}
	if got := prefix.TTLSecond(); got != 0 {
		t.Fatalf("nil ttl seconds should be zero, got %d", got)
	}
}

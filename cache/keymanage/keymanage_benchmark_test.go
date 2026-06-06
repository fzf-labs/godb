package keymanage

import "testing"

var benchKeyManageString string
var benchKeyManageStrings []string

func BenchmarkKeyPrefixKey(b *testing.B) {
	prefix := &KeyPrefix{
		ServerName: "svc",
		PrefixName: "user",
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchKeyManageString = prefix.Key("a:b", "c\\d", "e'f")
	}
}

func BenchmarkKeyPrefixKeys(b *testing.B) {
	prefix := &KeyPrefix{
		ServerName: "svc",
		PrefixName: "user",
	}
	keys := []string{"a:b", "c\\d", "e'f", "plain"}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchKeyManageStrings = prefix.Keys(keys)
	}
}

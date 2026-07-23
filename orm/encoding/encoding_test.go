package encoding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripPayload struct {
	Name string `json:"name" msgpack:"name"`
	N    int    `json:"n" msgpack:"n"`
}

func TestEncodingRoundTrips(t *testing.T) {
	encoders := map[string]API{
		"json":    NewJSON(),
		"sonic":   NewSonic(),
		"msgpack": NewMsgPack(),
		"zlib":    NewZlib(),
	}
	for name, encoder := range encoders {
		t.Run(name, func(t *testing.T) {
			in := roundTripPayload{Name: "demo", N: 42}
			data, err := encoder.Marshal(in)
			require.NoError(t, err)
			var out roundTripPayload
			require.NoError(t, encoder.Unmarshal(data, &out))
			if out != in {
				t.Fatalf("unexpected round trip: %#v", out)
			}
		})
	}
}

func TestEncodingErrors(t *testing.T) {
	if _, err := NewJSON().Marshal(func() {}); err == nil {
		t.Fatal("expected json marshal error")
	}
	if err := NewJSON().Unmarshal([]byte("{"), &roundTripPayload{}); err == nil {
		t.Fatal("expected json unmarshal error")
	}
	if _, err := NewSonic().Marshal(func() {}); err == nil {
		t.Fatal("expected sonic marshal error")
	}
	if err := NewSonic().Unmarshal([]byte("{"), &roundTripPayload{}); err == nil {
		t.Fatal("expected sonic unmarshal error")
	}
	if _, err := NewMsgPack().Marshal(func() {}); err == nil {
		t.Fatal("expected msgpack marshal error")
	}
	if err := NewMsgPack().Unmarshal([]byte{0xc1}, &roundTripPayload{}); err == nil {
		t.Fatal("expected msgpack unmarshal error")
	}
	if _, err := NewZlib().Marshal(func() {}); err == nil {
		t.Fatal("expected zlib marshal error")
	}
	if err := NewZlib().Unmarshal([]byte("not zlib"), &roundTripPayload{}); err == nil {
		t.Fatal("expected zlib unmarshal error")
	}
}

func TestZlibCompressAndUncompress(t *testing.T) {
	z := NewZlib()
	compressed, err := z.ZlibCompress([]byte("hello"))
	require.NoError(t, err)
	uncompressed, err := z.ZlibUnCompress(compressed)
	require.NoError(t, err)
	if string(uncompressed) != "hello" {
		t.Fatalf("unexpected payload: %s", string(uncompressed))
	}
	if _, err := z.ZlibUnCompress([]byte("bad")); err == nil {
		t.Fatal("expected invalid compressed data to fail")
	}
	if _, err := z.ZlibUnCompress(compressed[:len(compressed)-1]); err == nil {
		t.Fatal("expected truncated compressed data to fail")
	}
}

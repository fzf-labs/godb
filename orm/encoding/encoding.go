package encoding

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/bytedance/sonic"
	"github.com/klauspost/compress/zlib"
	"github.com/vmihailenco/msgpack/v5"
)

type API interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

func NewJSON() *JSON {
	return &JSON{}
}

type JSON struct{}

func (e *JSON) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (e *JSON) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func NewSonic() *Sonic {
	return &Sonic{}
}

type Sonic struct{}

func (e *Sonic) Marshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

func (e *Sonic) Unmarshal(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}

func NewMsgPack() *MsgPack {
	return &MsgPack{}
}

type MsgPack struct{}

func (e *MsgPack) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (e *MsgPack) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

func NewZlib() *Zlib {
	return &Zlib{}
}

type Zlib struct{}

func (z *Zlib) Marshal(v interface{}) ([]byte, error) {
	marshal, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}
	return z.ZlibCompress(marshal)
}

func (z *Zlib) Unmarshal(data []byte, v interface{}) error {
	compress, err := z.ZlibUnCompress(data)
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(compress, v)
}

func (z *Zlib) ZlibCompress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	_ = w.Close()
	return b.Bytes(), nil
}

func (z *Zlib) ZlibUnCompress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := bytes.NewReader(data)
	r, err := zlib.NewReader(w)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(&b, r)
	if err != nil {
		return nil, err
	}
	_ = r.Close()
	return b.Bytes(), nil
}

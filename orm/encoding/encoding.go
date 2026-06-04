package encoding

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/bytedance/sonic"
	"github.com/klauspost/compress/zlib"
	"github.com/vmihailenco/msgpack/v5"
)

// API 定义对象编解码器的统一接口。
type API interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// NewJSON 创建标准库 JSON 编解码器。
func NewJSON() *JSON {
	return &JSON{}
}

// JSON 使用标准库 encoding/json 进行编解码。
type JSON struct{}

// Marshal 将对象编码为 JSON 字节。
func (e *JSON) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal 将 JSON 字节解码到目标对象。
func (e *JSON) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewSonic 创建 Sonic JSON 编解码器。
func NewSonic() *Sonic {
	return &Sonic{}
}

// Sonic 使用 bytedance/sonic 进行 JSON 编解码。
type Sonic struct{}

// Marshal 将对象编码为 Sonic JSON 字节。
func (e *Sonic) Marshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

// Unmarshal 将 Sonic JSON 字节解码到目标对象。
func (e *Sonic) Unmarshal(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}

// NewMsgPack 创建 MsgPack 编解码器。
func NewMsgPack() *MsgPack {
	return &MsgPack{}
}

// MsgPack 使用 MessagePack 格式进行编解码。
type MsgPack struct{}

// Marshal 将对象编码为 MessagePack 字节。
func (e *MsgPack) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Unmarshal 将 MessagePack 字节解码到目标对象。
func (e *MsgPack) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

// NewZlib 创建 MessagePack 加 Zlib 压缩的编解码器。
func NewZlib() *Zlib {
	return &Zlib{}
}

// Zlib 使用 MessagePack 编码并通过 Zlib 压缩存储。
type Zlib struct{}

// Marshal 将对象编码为 MessagePack 后再进行 Zlib 压缩。
func (z *Zlib) Marshal(v interface{}) ([]byte, error) {
	marshal, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}
	return z.ZlibCompress(marshal)
}

// Unmarshal 解压 Zlib 数据并按 MessagePack 解码到目标对象。
func (z *Zlib) Unmarshal(data []byte, v interface{}) error {
	compress, err := z.ZlibUnCompress(data)
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(compress, v)
}

// ZlibCompress 对字节切片进行 Zlib 压缩。
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

// ZlibUnCompress 对 Zlib 压缩字节进行解压。
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

package bucket

import (
	"bytes"
	"encoding/gob"
)

type Codec interface {
	Encode(e interface{}) (*bytes.Buffer, error)
	Decode(b []byte, e interface{}) error
}

type codec struct{}

func NewCodec() Codec {
	return &codec{}
}

func (c *codec) Encode(e interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(&e)
	return buf, err
}

func (c *codec) Decode(raw []byte, e interface{}) (err error) {
	buf := BufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		BufPool.Put(buf)
	}()
	buf.Write(raw)
	dec := gob.NewDecoder(buf)
	return dec.Decode(e)
}

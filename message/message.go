package message

import (
	"context"
	"encoding/json"
)

const TypeHTTPGet = "httpget"

type Message struct {
	Type        string `json:"type"`
	HTTPRequest string `json:"http_request"`
}

type Encoder struct {
}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) Encode(ctx context.Context, message Message) ([]byte, error) {
	return json.Marshal(message)
}

type Decoder struct {
}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) Decode(ctx context.Context, b []byte) (Message, error) {
	m := Message{}
	err := json.Unmarshal(b, &m)

	return m, err
}

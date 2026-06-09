// Package msgpack provides a [encoding.Codec] implementation using
// github.com/vmihailenco/msgpack/v5.
//
// MessagePack is a compact binary serialization format that is
// cross-language and more efficient than JSON for structured data.
//
// The codec self-registers under the name "msgpack" via init().
package msgpack

import (
	"github.com/tx7do/go-wind-plugins/encoding"
	"github.com/vmihailenco/msgpack/v5"
)

// Name is the name registered for the MessagePack codec.
const Name = "msgpack"

func init() {
	encoding.RegisterCodec(codec{})
}

// codec implements encoding.Codec using vmihailenco/msgpack/v5.
type codec struct{}

// Marshal encodes v into MessagePack bytes.
func (codec) Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Unmarshal decodes MessagePack data into v.
func (codec) Unmarshal(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}

// Name returns the codec name.
func (codec) Name() string {
	return Name
}

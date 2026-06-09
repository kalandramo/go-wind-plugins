package proto

import (
	"fmt"

	"github.com/tx7do/go-wind-plugins/encoding"
	"google.golang.org/protobuf/proto"
)

// Name is the name registered for the protobuf codec.
const Name = "proto"

func init() {
	encoding.RegisterCodec(codec{})
}

// codec implements encoding.Codec using google.golang.org/protobuf.
type codec struct{}

// Marshal encodes v (which must be a proto.Message) into protobuf bytes.
func (codec) Marshal(v any) ([]byte, error) {
	m, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("proto: failed to marshal %T — not a proto.Message", v)
	}
	return proto.Marshal(m)
}

// Unmarshal decodes protobuf data into v (which must be a proto.Message).
func (codec) Unmarshal(data []byte, v any) error {
	m, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("proto: failed to unmarshal into %T — not a proto.Message", v)
	}
	return proto.Unmarshal(data, m)
}

// Name returns the codec name.
func (codec) Name() string {
	return Name
}

// Package xml provides a [encoding.Codec] implementation using Go's
// standard encoding/xml package.
//
// The codec self-registers under the name "xml" via init().
package xml

import (
	"encoding/xml"

	"github.com/tx7do/go-wind-plugins/encoding"
)

// Name is the name registered for the XML codec.
const Name = "xml"

func init() {
	encoding.RegisterCodec(codec{})
}

// codec implements encoding.Codec using encoding/xml.
type codec struct{}

// Marshal encodes v into XML bytes.
func (codec) Marshal(v any) ([]byte, error) {
	return xml.Marshal(v)
}

// Unmarshal decodes XML data into v.
func (codec) Unmarshal(data []byte, v any) error {
	return xml.Unmarshal(data, v)
}

// Name returns the codec name.
func (codec) Name() string {
	return Name
}

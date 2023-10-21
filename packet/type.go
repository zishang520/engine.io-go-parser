package packet

import (
	"io"

	"github.com/zishang520/engine.io-go-parser/types"
)

type (
	Type string

	Options struct {
		Compress bool `json:"compress" mapstructure:"compress" msgpack:"compress"`
	}

	Packet struct {
		Type         Type                  `json:"type" mapstructure:"type" msgpack:"type"`
		Data         io.Reader             `json:"data,omitempty" mapstructure:"data,omitempty" msgpack:"data,omitempty"`
		Options      *Options              `json:"options,omitempty" mapstructure:"options,omitempty" msgpack:"options,omitempty"`
		WsPreEncoded types.BufferInterface `json:"wsPreEncoded,omitempty" mapstructure:"wsPreEncoded,omitempty" msgpack:"wsPreEncoded,omitempty"`
	}
)

// Packet types.
const (
	OPEN    Type = "open"
	CLOSE   Type = "close"
	PING    Type = "ping"
	PONG    Type = "pong"
	MESSAGE Type = "message"
	UPGRADE Type = "upgrade"
	NOOP    Type = "noop"
	ERROR   Type = "error"
)

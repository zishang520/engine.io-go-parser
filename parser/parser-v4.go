package parser

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/zishang520/engine.io-go-parser/packet"
	"github.com/zishang520/engine.io-go-parser/types"
)

type parserv4 struct{}

var (
	defaultParserv4 Parser = &parserv4{}
)

func Parserv4() Parser {
	return defaultParserv4
}

// Current protocol version.
func (*parserv4) Protocol() int {
	return Protocol
}

func (p *parserv4) EncodePacket(data *packet.Packet, supportsBinary bool, _ ...bool) (types.BufferInterface, error) {
	if data == nil {
		return nil, ErrPacketNil
	}

	if c, ok := data.Data.(io.Closer); ok {
		defer c.Close()
	}

	_type, _type_ok := PACKET_TYPES[data.Type]
	if !_type_ok {
		return nil, ErrPacketType
	}

	switch v := data.Data.(type) {
	case *types.StringBuffer, *strings.Reader:
		encode := types.NewStringBuffer(nil)
		if err := encode.WriteByte(_type); err != nil {
			return nil, err
		}
		if _, err := io.Copy(encode, v); err != nil {
			return nil, err
		}

		return encode, nil
	case io.Reader:
		if !supportsBinary {
			// only 'message' packets can contain binary, so the type prefix is not needed
			encode := types.NewStringBuffer(nil)
			if err := encode.WriteByte('b'); err != nil {
				return nil, err
			}
			b64 := base64.NewEncoder(base64.StdEncoding, encode)
			defer b64.Close()
			if _, err := io.Copy(b64, v); err != nil {
				return nil, err
			}
			return encode, nil
		}
		// plain string
		encode := types.NewBytesBuffer(nil)
		if _, err := io.Copy(encode, v); err != nil {
			return nil, err
		}
		return encode, nil
	}
	encode := types.NewStringBuffer(nil)
	if err := encode.WriteByte(_type); err != nil {
		return nil, err
	}
	return encode, nil
}

func (p *parserv4) DecodePacket(data types.BufferInterface, _ ...bool) (*packet.Packet, error) {
	if data == nil {
		return ERROR_PACKET, ErrDataNil
	}

	// strings
	switch v := data.(type) {
	case *types.StringBuffer:
		msgType, err := v.ReadByte()
		if err != nil {
			return ERROR_PACKET, err
		}
		if msgType == 'b' {
			decode := types.NewBytesBuffer(nil)
			if _, err := decode.ReadFrom(base64.NewDecoder(base64.StdEncoding, v)); err != nil {
				return ERROR_PACKET, err
			}
			return &packet.Packet{Type: packet.MESSAGE, Data: decode}, nil
		}
		packetType, ok := PACKET_TYPES_REVERSE[msgType]
		if !ok {
			return ERROR_PACKET, fmt.Errorf(`%w, unknown data type [%c]`, ErrParser, msgType)
		}
		stringBuffer := types.NewStringBuffer(nil)
		if _, err := stringBuffer.ReadFrom(v); err != nil {
			return ERROR_PACKET, err
		}
		return &packet.Packet{Type: packetType, Data: stringBuffer}, nil
	}

	// binary
	decode := types.NewBytesBuffer(nil)
	if _, err := io.Copy(decode, data); err != nil {
		return ERROR_PACKET, err
	}
	return &packet.Packet{Type: packet.MESSAGE, Data: decode}, nil
}

func (p *parserv4) EncodePayload(packets []*packet.Packet, _ ...bool) (types.BufferInterface, error) {
	enPayload := types.NewStringBuffer(nil)

	for _, packet := range packets {
		if buf, err := p.EncodePacket(packet, false); err != nil {
			return nil, err
		} else {
			if enPayload.Len() > 0 {
				if err := enPayload.WriteByte(SEPARATOR); err != nil {
					return nil, err
				}
			}
			if _, err := buf.WriteTo(enPayload); err != nil {
				return nil, err
			}
		}
	}

	return enPayload, nil
}

func (p *parserv4) DecodePayload(data types.BufferInterface) (packets []*packet.Packet, _ error) {
	scanner := bufio.NewScanner(data)
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, SEPARATOR); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	for scanner.Scan() {
		if packet, err := p.DecodePacket(types.NewStringBuffer(scanner.Bytes())); err == nil {
			packets = append(packets, packet)
		} else {
			return packets, err
		}
	}
	return packets, scanner.Err()
}

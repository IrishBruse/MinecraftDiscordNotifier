package packet

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	MaxPacketLength int = 2097151
	MaxStringLength int = 32767
)

// OutboundPacket represents a packet to be sent over a network connection.
type OutboundPacket struct {
	payload []byte
}

// NewOutboundPacket creates a new OutboundPacket with a given id.
func NewOutboundPacket(id int32) *OutboundPacket {
	p := &OutboundPacket{}
	p.WriteVarInt(id)
	return p
}

// WriteInt writes a 32-bit integer to the packet.
func (p *OutboundPacket) WriteInt(n int32) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(n))
	p.WriteBytes(buf)
}

// WriteShort writes a 16-bit integer to the packet.
func (p *OutboundPacket) WriteShort(n int16) {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(n))
	p.WriteBytes(buf)
}

// WriteLong writes a 64-bit integer to the packet.
func (p *OutboundPacket) WriteLong(n int64) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n))
	p.WriteBytes(buf)
}

// WriteVarInt writes a variable-length 32-bit integer to the packet.
func (p *OutboundPacket) WriteVarInt(n int32) {
	buf := make([]byte, binary.MaxVarintLen32)
	size := binary.PutUvarint(buf, uint64(n))
	p.WriteBytes(buf[:size])
}

// WriteVarLong writes a variable-length 64-bit integer to the packet.
func (p *OutboundPacket) WriteVarLong(n int64) {
	buf := make([]byte, binary.MaxVarintLen64)
	size := binary.PutUvarint(buf, uint64(n))
	p.WriteBytes(buf[:size])
}

// WriteBool writes a boolean value to the packet.
func (p *OutboundPacket) WriteBool(value bool) {
	if value {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}
}

// WriteString writes a string to the packet.
func (p *OutboundPacket) WriteString(str string) error {
	length := len(str)
	if length > MaxStringLength {
		return fmt.Errorf("string is longer than %d", MaxStringLength)
	}

	p.WriteVarInt(int32(length))
	p.WriteBytes([]byte(str))

	return nil
}

// WriteByte writes a single byte to the packet.
func (p *OutboundPacket) WriteByte(b byte) {
	p.payload = append(p.payload, b)
}

// WriteBytes writes a byte slice to the packet.
func (p *OutboundPacket) WriteBytes(b []byte) {
	p.payload = append(p.payload, b...)
}

// Write sends the packet over the given network connection.
func (p *OutboundPacket) Write(conn net.Conn) error {
	payload := p.payload
	length := len(payload)

	if length > MaxPacketLength {
		return fmt.Errorf("packet exceeds max packet length of %d by %d bytes", MaxPacketLength, length-MaxPacketLength)
	}

	if _, err := conn.Write(encodeVarInt(int32(length))); err != nil {
		return fmt.Errorf("failed to write packet length: %w", err)
	}

	if _, err := conn.Write(payload); err != nil {
		return fmt.Errorf("failed to write packet payload: %w", err)
	}

	return nil
}

// encodeVarInt encodes an integer into a variable-length byte slice.
func encodeVarInt(value int32) []byte {
	buf := make([]byte, binary.MaxVarintLen32)
	size := binary.PutUvarint(buf, uint64(value))
	return buf[:size]
}

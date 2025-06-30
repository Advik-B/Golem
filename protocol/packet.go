package protocol

import "github.com/Advik-B/Golem/protocol/codec"

// Packet is the interface implemented by all network packets.
type Packet interface {
	// ID returns the constant packet ID for the packet's state and flow.
	ID() int32

	// ReadFrom deserializes the packet's data from a PacketBuffer.
	ReadFrom(r *codec.PacketBuffer) error

	// WriteTo serializes the packet's data to a PacketBuffer.
	WriteTo(w *codec.PacketBuffer) error
}

// Flow represents the direction of a packet (client-to-server or server-to-client).
type Flow int

const (
	Serverbound Flow = iota
	Clientbound
)

func (f Flow) String() string {
	if f == Serverbound {
		return "serverbound"
	}
	return "clientbound"
}

// State represents the current state of the connection (e.g., Handshaking, Play).
type State int

const (
	Handshaking State = iota
	Status
	Login
	Configuration
	Play
)

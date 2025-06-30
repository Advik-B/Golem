package handshake

import (
	"fmt"

	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec"
)

func init() {
	protocol.Registry[protocol.Handshaking][protocol.Serverbound].Register(
		ClientIntentionPacketID, func() protocol.Packet { return &ClientIntentionPacket{} },
	)
}

const (
	ClientIntentionPacketID = 0x00
)

// ClientIntent corresponds to Java's ClientIntent enum.
type ClientIntent int32

const (
	StatusIntent   ClientIntent = 1
	LoginIntent    ClientIntent = 2
	TransferIntent ClientIntent = 3
)

// ClientIntentionPacket is sent by the client to initiate a connection.
type ClientIntentionPacket struct {
	ProtocolVersion int32
	HostName        string
	Port            uint16
	Intention       ClientIntent
}

func (pk *ClientIntentionPacket) ID() int32 { return ClientIntentionPacketID }

func (pk *ClientIntentionPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	if pk.ProtocolVersion, err = r.ReadVarInt(); err != nil {
		return err
	}
	if pk.HostName, err = r.ReadString(255); err != nil {
		return err
	}
	var port int16
	if port, err = r.ReadShort(); err != nil {
		return err
	}
	pk.Port = uint16(port)
	var intent int32
	if intent, err = r.ReadVarInt(); err != nil {
		return err
	}
	if intent < 1 || intent > 3 {
		return fmt.Errorf("unknown connection intent: %d", intent)
	}
	pk.Intention = ClientIntent(intent)
	return nil
}

func (pk *ClientIntentionPacket) WriteTo(w *codec.PacketBuffer) error {
	w.WriteVarInt(pk.ProtocolVersion)
	w.WriteString(pk.HostName)
	w.WriteShort(int16(pk.Port))
	w.WriteVarInt(int32(pk.Intention))
	return nil
}

// ServerHandshakePacketListener defines the handler for handshake packets.
type ServerHandshakePacketListener interface {
	HandleIntention(pk *ClientIntentionPacket) error
}

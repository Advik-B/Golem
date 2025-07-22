package protocol

import (
	"github.com/google/uuid"
	"io"
)

// --- DisconnectPacket (Clientbound) ---
type LoginDisconnectPacket struct {
	Reason string // JSON encoded chat message
}

func (p *LoginDisconnectPacket) ID(version string) uint32 { return 0x00 }
func (p *LoginDisconnectPacket) State() State             { return StateLogin }
func (p *LoginDisconnectPacket) ReadFrom(r io.Reader, version string) (err error) {
	p.Reason, err = ReadString(r)
	return
}
func (p *LoginDisconnectPacket) WriteTo(w io.Writer, version string) (err error) {
	return WriteString(w, p.Reason)
}

// --- EncryptionBeginPacket (Clientbound) ---

type EncryptionBeginPacket struct {
	ServerID           string
	PublicKey          []byte
	VerifyToken        []byte
	ShouldAuthenticate bool
}

func (p *EncryptionBeginPacket) ID(version string) uint32 { return 0x01 }
func (p *EncryptionBeginPacket) State() State             { return StateLogin }
func (p *EncryptionBeginPacket) ReadFrom(r io.Reader, version string) (err error) {
	// This packet is only sent by the server, so reading is not needed for a server implementation.
	return nil
}
func (p *EncryptionBeginPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteString(w, p.ServerID); err != nil {
		return
	}
	if err = WriteVarInt(w, int32(len(p.PublicKey))); err != nil {
		return
	}
	if _, err = w.Write(p.PublicKey); err != nil {
		return
	}
	if err = WriteVarInt(w, int32(len(p.VerifyToken))); err != nil {
		return
	}
	if _, err = w.Write(p.VerifyToken); err != nil {
		return
	}
	err = WriteBool(w, p.ShouldAuthenticate)
	return
}

// --- SuccessPacket (Clientbound) ---

type LoginSuccessPacket struct {
	UUID       uuid.UUID
	Username   string
	Properties []struct {
		Name      string
		Value     string
		Signature *string
	}
}

func (p *LoginSuccessPacket) ID(version string) uint32 { return 0x02 }
func (p *LoginSuccessPacket) State() State             { return StateLogin }
func (p *LoginSuccessPacket) ReadFrom(r io.Reader, version string) (err error) {
	// This packet is only sent by the server.
	return nil
}
func (p *LoginSuccessPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteUUID(w, p.UUID); err != nil {
		return
	}
	if err = WriteString(w, p.Username); err != nil {
		return
	}
	if err = WriteVarInt(w, int32(len(p.Properties))); err != nil {
		return
	}
	for _, prop := range p.Properties {
		if err = WriteString(w, prop.Name); err != nil {
			return
		}
		if err = WriteString(w, prop.Value); err != nil {
			return
		}
		hasSignature := prop.Signature != nil
		if err = WriteBool(w, hasSignature); err != nil {
			return
		}
		if hasSignature {
			if err = WriteString(w, *prop.Signature); err != nil {
				return
			}
		}
	}
	return
}

// --- CompressPacket (Clientbound) ---

type CompressPacket struct {
	Threshold int32
}

func (p *CompressPacket) ID(version string) uint32 { return 0x03 }
func (p *CompressPacket) State() State             { return StateLogin }
func (p *CompressPacket) ReadFrom(r io.Reader, version string) (err error) {
	// Server only
	return nil
}
func (p *CompressPacket) WriteTo(w io.Writer, version string) (err error) {
	return WriteVarInt(w, p.Threshold)
}

// --- LoginPluginRequestPacket (Clientbound) ---

type LoginPluginRequestPacket struct {
	MessageID int32
	Channel   string
	Data      []byte
}

func (p *LoginPluginRequestPacket) ID(version string) uint32 { return 0x04 }
func (p *LoginPluginRequestPacket) State() State             { return StateLogin }
func (p *LoginPluginRequestPacket) ReadFrom(r io.Reader, version string) (err error) {
	// Server only
	return nil
}
func (p *LoginPluginRequestPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.MessageID); err != nil {
		return
	}
	if err = WriteString(w, p.Channel); err != nil {
		return
	}
	_, err = w.Write(p.Data)
	return
}

// --- CookieRequestPacket (Clientbound) ---
// This is a common packet, but its first appearance is in the Login state.
// We will define it fully here.

type CookieRequestPacket struct {
	Key string
}

func (p *CookieRequestPacket) ID(version string) uint32 { return 0x05 }
func (p *CookieRequestPacket) State() State             { return StateLogin }
func (p *CookieRequestPacket) ReadFrom(r io.Reader, version string) (err error) {
	// Server only
	return nil
}
func (p *CookieRequestPacket) WriteTo(w io.Writer, version string) (err error) {
	return WriteString(w, p.Key)
}

// --- LoginStartPacket (Serverbound) ---

type LoginStartPacket struct {
	Username   string
	PlayerUUID uuid.UUID
}

func (p *LoginStartPacket) ID(version string) uint32 { return 0x00 }
func (p *LoginStartPacket) State() State             { return StateLogin }
func (p *LoginStartPacket) ReadFrom(r io.Reader, version string) (err error) {
	if p.Username, err = ReadString(r); err != nil {
		return
	}
	p.PlayerUUID, err = ReadUUID(r)
	return
}
func (p *LoginStartPacket) WriteTo(w io.Writer, version string) (err error) {
	// Client only
	return nil
}

// --- EncryptionBeginResponsePacket (Serverbound) ---

type EncryptionBeginResponsePacket struct {
	SharedSecret []byte
	VerifyToken  []byte
}

func (p *EncryptionBeginResponsePacket) ID(version string) uint32 { return 0x01 }
func (p *EncryptionBeginResponsePacket) State() State             { return StateLogin }
func (p *EncryptionBeginResponsePacket) ReadFrom(r io.Reader, version string) (err error) {
	lenSecret, err := ReadVarInt(r)
	if err != nil {
		return
	}
	p.SharedSecret = make([]byte, lenSecret)
	if _, err = io.ReadFull(r, p.SharedSecret); err != nil {
		return
	}
	lenToken, err := ReadVarInt(r)
	if err != nil {
		return
	}
	p.VerifyToken = make([]byte, lenToken)
	_, err = io.ReadFull(r, p.VerifyToken)
	return
}
func (p *EncryptionBeginResponsePacket) WriteTo(w io.Writer, version string) (err error) {
	// Client only
	return nil
}

// --- LoginPluginResponsePacket (Serverbound) ---

type LoginPluginResponsePacket struct {
	MessageID int32
	Success   bool
	Data      []byte
}

func (p *LoginPluginResponsePacket) ID(version string) uint32 { return 0x02 }
func (p *LoginPluginResponsePacket) State() State             { return StateLogin }
func (p *LoginPluginResponsePacket) ReadFrom(r io.Reader, version string) (err error) {
	if p.MessageID, err = ReadVarInt(r); err != nil {
		return
	}
	if p.Success, err = ReadBool(r); err != nil {
		return
	}
	if p.Success {
		// Read remaining bytes as data
		p.Data, err = io.ReadAll(r)
	}
	return
}
func (p *LoginPluginResponsePacket) WriteTo(w io.Writer, version string) (err error) {
	// Client only
	return nil
}

// --- LoginAcknowledgedPacket (Serverbound) ---

type LoginAcknowledgedPacket struct {
	// No fields
}

func (p *LoginAcknowledgedPacket) ID(version string) uint32                         { return 0x03 }
func (p *LoginAcknowledgedPacket) State() State                                     { return StateLogin }
func (p *LoginAcknowledgedPacket) ReadFrom(r io.Reader, version string) (err error) { return nil }
func (p *LoginAcknowledgedPacket) WriteTo(w io.Writer, version string) (err error)  { return nil }

// --- CookieResponsePacket (Serverbound) ---
type CookieResponsePacket struct {
	Key      string
	HasValue bool
	Value    []byte
}

func (p *CookieResponsePacket) ID(version string) uint32 { return 0x04 }
func (p *CookieResponsePacket) State() State             { return StateLogin }
func (p *CookieResponsePacket) ReadFrom(r io.Reader, version string) (err error) {
	if p.Key, err = ReadString(r); err != nil {
		return
	}
	if p.HasValue, err = ReadBool(r); err != nil {
		return
	}
	if p.HasValue {
		var length int32
		if length, err = ReadVarInt(r); err != nil {
			return
		}
		p.Value = make([]byte, length)
		_, err = io.ReadFull(r, p.Value)
	}
	return
}
func (p *CookieResponsePacket) WriteTo(w io.Writer, version string) (err error) {
	// Client only
	return nil
}

// init registers the packets for the Login state.
func init() {
	// Clientbound (Server -> Client)
	RegisterPacket(StateLogin, Clientbound, 0x00, func() Packet { return &LoginDisconnectPacket{} })
	RegisterPacket(StateLogin, Clientbound, 0x01, func() Packet { return &EncryptionBeginPacket{} })
	RegisterPacket(StateLogin, Clientbound, 0x02, func() Packet { return &LoginSuccessPacket{} })
	RegisterPacket(StateLogin, Clientbound, 0x03, func() Packet { return &CompressPacket{} })
	RegisterPacket(StateLogin, Clientbound, 0x04, func() Packet { return &LoginPluginRequestPacket{} })
	RegisterPacket(StateLogin, Clientbound, 0x05, func() Packet { return &CookieRequestPacket{} })

	// Serverbound (Client -> Server)
	RegisterPacket(StateLogin, Serverbound, 0x00, func() Packet { return &LoginStartPacket{} })
	RegisterPacket(StateLogin, Serverbound, 0x01, func() Packet { return &EncryptionBeginResponsePacket{} })
	RegisterPacket(StateLogin, Serverbound, 0x02, func() Packet { return &LoginPluginResponsePacket{} })
	RegisterPacket(StateLogin, Serverbound, 0x03, func() Packet { return &LoginAcknowledgedPacket{} })
	RegisterPacket(StateLogin, Serverbound, 0x04, func() Packet { return &CookieResponsePacket{} })
}

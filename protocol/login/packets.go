package login

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"fmt"

	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec"
	"github.com/google/uuid"
)

// Packet IDs for the Login state.
const (
	// Serverbound
	ServerboundHelloPacketID             = 0x00
	ServerboundKeyPacketID               = 0x01
	ServerboundCustomQueryAnswerPacketID = 0x02
	ServerboundLoginAcknowledgedPacketID = 0x03

	// Clientbound
	ClientboundLoginDisconnectPacketID  = 0x00
	ClientboundHelloPacketID            = 0x01
	ClientboundLoginFinishedPacketID    = 0x02
	ClientboundLoginCompressionPacketID = 0x03
	ClientboundCustomQueryPacketID      = 0x04
)

// --- Serverbound ---

type ServerboundHelloPacket struct {
	Name      string
	ProfileID uuid.UUID
}

func (pk *ServerboundHelloPacket) ID() int32 { return ServerboundHelloPacketID }
func (pk *ServerboundHelloPacket) WriteTo(w *codec.PacketBuffer) error {
	w.WriteString(pk.Name)
	w.WriteUUID(pk.ProfileID)
	return nil
}
func (pk *ServerboundHelloPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	if pk.Name, err = r.ReadString(16); err != nil {
		return err
	}
	pk.ProfileID, err = r.ReadUUID()
	return err
}

type ServerboundKeyPacket struct {
	KeyBytes           []byte
	EncryptedChallenge []byte
}

func (pk *ServerboundKeyPacket) ID() int32 { return ServerboundKeyPacketID }
func (pk *ServerboundKeyPacket) WriteTo(w *codec.PacketBuffer) error {
	w.WriteByteArray(pk.KeyBytes)
	w.WriteByteArray(pk.EncryptedChallenge)
	return nil
}
func (pk *ServerboundKeyPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	if pk.KeyBytes, err = r.ReadByteArray(); err != nil {
		return err
	}
	pk.EncryptedChallenge, err = r.ReadByteArray()
	return err
}

// DecryptSharedSecret decrypts the shared secret using the server's private key.
func (pk *ServerboundKeyPacket) DecryptSharedSecret(priv *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, priv, pk.KeyBytes)
}

// DecryptChallenge decrypts the challenge using the server's private key.
func (pk *ServerboundKeyPacket) DecryptChallenge(priv *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, priv, pk.EncryptedChallenge)
}

type ServerboundCustomQueryAnswerPacket struct {
	TransactionID int32
	Payload       []byte // Opaque payload, can be nil
}

func (pk *ServerboundCustomQueryAnswerPacket) ID() int32 { return ServerboundCustomQueryAnswerPacketID }
func (pk *ServerboundCustomQueryAnswerPacket) WriteTo(w *codec.PacketBuffer) error {
	w.WriteVarInt(pk.TransactionID)
	hasPayload := pk.Payload != nil
	w.WriteBool(hasPayload)
	if hasPayload {
		w.Write(pk.Payload) // The Java impl has more complex logic here we simplify
	}
	return nil
}
func (pk *ServerboundCustomQueryAnswerPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	if pk.TransactionID, err = r.ReadVarInt(); err != nil {
		return err
	}
	hasPayload, err := r.ReadBool()
	if err != nil {
		return err
	}
	if hasPayload {
		// Read remaining bytes as payload
		pk.Payload = r.Bytes()
	}
	return nil
}

type ServerboundLoginAcknowledgedPacket struct{}

func (pk *ServerboundLoginAcknowledgedPacket) ID() int32                            { return ServerboundLoginAcknowledgedPacketID }
func (pk *ServerboundLoginAcknowledgedPacket) WriteTo(w *codec.PacketBuffer) error  { return nil }
func (pk *ServerboundLoginAcknowledgedPacket) ReadFrom(r *codec.PacketBuffer) error { return nil }

// --- Clientbound ---

type ClientboundLoginDisconnectPacket struct {
	Reason protocol.Component
}

func (pk *ClientboundLoginDisconnectPacket) ID() int32 { return ClientboundLoginDisconnectPacketID }
func (pk *ClientboundLoginDisconnectPacket) WriteTo(w *codec.PacketBuffer) error {
	return pk.Reason.WriteTo(w)
}
func (pk *ClientboundLoginDisconnectPacket) ReadFrom(r *codec.PacketBuffer) error {
	return pk.Reason.ReadFrom(r)
}

type ClientboundHelloPacket struct {
	ServerID     string
	PublicKey    []byte
	Challenge    []byte
	AuthRequired bool
}

func (pk *ClientboundHelloPacket) ID() int32 { return ClientboundHelloPacketID }
func (pk *ClientboundHelloPacket) WriteTo(w *codec.PacketBuffer) error {
	w.WriteString(pk.ServerID)
	w.WriteByteArray(pk.PublicKey)
	w.WriteByteArray(pk.Challenge)
	w.WriteBool(pk.AuthRequired)
	return nil
}
func (pk *ClientboundHelloPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	if pk.ServerID, err = r.ReadString(20); err != nil {
		return err
	}
	if pk.PublicKey, err = r.ReadByteArray(); err != nil {
		return err
	}
	if pk.Challenge, err = r.ReadByteArray(); err != nil {
		return err
	}
	pk.AuthRequired, err = r.ReadBool()
	return err
}

// NewClientboundHelloPacket creates a packet for initiating encryption.
func NewClientboundHelloPacket(serverID string, pubKey *rsa.PublicKey, challenge []byte, auth bool) (*ClientboundHelloPacket, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	return &ClientboundHelloPacket{
		ServerID:     serverID,
		PublicKey:    pubKeyBytes,
		Challenge:    challenge,
		AuthRequired: auth,
	}, nil
}

// Creates a hash for server authentication.
func (pk *ClientboundHelloPacket) ServerHash(sharedSecret []byte) string {
	h := sha1.New()
	h.Write([]byte(pk.ServerID))
	h.Write(sharedSecret)
	h.Write(pk.PublicKey)
	hash := h.Sum(nil)

	// Minecraft's hex digest is a little strange. It's a signed two's complement integer.
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		// Invert and add 1
		twosComplement := make([]byte, len(hash))
		for i, b := range hash {
			twosComplement[i] = ^b
		}
		for i := len(twosComplement) - 1; i >= 0; i-- {
			twosComplement[i]++
			if twosComplement[i] != 0 {
				break
			}
		}
		return "-" + fmt.Sprintf("%x", twosComplement)
	}
	return fmt.Sprintf("%x", hash)
}

type ClientboundLoginFinishedPacket struct {
	GameProfile protocol.GameProfile
}

func (pk *ClientboundLoginFinishedPacket) ID() int32 { return ClientboundLoginFinishedPacketID }
func (pk *ClientboundLoginFinishedPacket) WriteTo(w *codec.PacketBuffer) error {
	return pk.GameProfile.WriteTo(w)
}
func (pk *ClientboundLoginFinishedPacket) ReadFrom(r *codec.PacketBuffer) error {
	return pk.GameProfile.ReadFrom(r)
}

type ClientboundLoginCompressionPacket struct {
	CompressionThreshold int32
}

func (pk *ClientboundLoginCompressionPacket) ID() int32 { return ClientboundLoginCompressionPacketID }
func (pk *ClientboundLoginCompressionPacket) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteVarInt(pk.CompressionThreshold)
}
func (pk *ClientboundLoginCompressionPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.CompressionThreshold, err = r.ReadVarInt()
	return err
}

type ClientboundCustomQueryPacket struct {
	TransactionID int32
	Identifier    protocol.ResourceLocation
	Payload       []byte
}

func (pk *ClientboundCustomQueryPacket) ID() int32 { return ClientboundCustomQueryPacketID }
func (pk *ClientboundCustomQueryPacket) WriteTo(w *codec.PacketBuffer) error {
	w.WriteVarInt(pk.TransactionID)
	pk.Identifier.WriteTo(w)
	w.Write(pk.Payload)
	return nil
}
func (pk *ClientboundCustomQueryPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	if pk.TransactionID, err = r.ReadVarInt(); err != nil {
		return err
	}
	if err = pk.Identifier.ReadFrom(r); err != nil {
		return err
	}
	pk.Payload = r.Bytes() // Read the rest of the buffer
	return nil
}

// --- Listeners ---

type ServerLoginPacketListener interface {
	HandleHello(pk *ServerboundHelloPacket) error
	HandleKey(pk *ServerboundKeyPacket) error
	HandleCustomQueryAnswer(pk *ServerboundCustomQueryAnswerPacket) error
	HandleLoginAcknowledged(pk *ServerboundLoginAcknowledgedPacket) error
}

type ClientLoginPacketListener interface {
	HandleDisconnect(pk *ClientboundLoginDisconnectPacket) error
	HandleHello(pk *ClientboundHelloPacket) error
	HandleLoginFinished(pk *ClientboundLoginFinishedPacket) error
	HandleCompression(pk *ClientboundLoginCompressionPacket) error
	HandleCustomQuery(pk *ClientboundCustomQueryPacket) error
}

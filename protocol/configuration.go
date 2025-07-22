// File: protocol/configuration.go
package protocol

import (
	"io"

	"github.com/Advik-B/Golem/nbt"
	"github.com/google/uuid"
)

// --- Common Packets ---

// ClientSettingsPacket is sent by the client to inform the server of its settings.
type ClientSettingsPacket struct {
	Locale              string
	ViewDistance        int8
	ChatFlags           int32
	ChatColors          bool
	SkinParts           uint8
	MainHand            int32
	EnableTextFiltering bool
	EnableServerListing bool
	ParticleStatus      int32
}

func (p *ClientSettingsPacket) ID(version string) uint32 { return 0x00 }
func (p *ClientSettingsPacket) State() State             { return StateConfiguration }
func (p *ClientSettingsPacket) ReadFrom(r io.Reader, version string) (err error) {
	if p.Locale, err = ReadString(r); err != nil {
		return
	}
	if p.ViewDistance, err = ReadInt8(r); err != nil {
		return
	}
	if p.ChatFlags, err = ReadVarInt(r); err != nil {
		return
	}
	if p.ChatColors, err = ReadBool(r); err != nil {
		return
	}
	if p.SkinParts, err = ReadUint8(r); err != nil {
		return
	}
	if p.MainHand, err = ReadVarInt(r); err != nil {
		return
	}
	if p.EnableTextFiltering, err = ReadBool(r); err != nil {
		return
	}
	if p.EnableServerListing, err = ReadBool(r); err != nil {
		return
	}
	p.ParticleStatus, err = ReadVarInt(r)
	return
}
func (p *ClientSettingsPacket) WriteTo(w io.Writer, version string) (err error) {
	// Client only
	return nil
}

// StoreCookiePacket is sent by the server to ask the client to store a cookie.
type StoreCookiePacket struct {
	Key   string
	Value []byte
}

func (p *StoreCookiePacket) ID(version string) uint32 { return 0x0A }
func (p *StoreCookiePacket) State() State             { return StateConfiguration }
func (p *StoreCookiePacket) ReadFrom(r io.Reader, version string) (err error) {
	// Server only
	return nil
}
func (p *StoreCookiePacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteString(w, p.Key); err != nil {
		return
	}
	if err = WriteVarInt(w, int32(len(p.Value))); err != nil {
		return
	}
	_, err = w.Write(p.Value)
	return
}

// TransferPacket is sent by the server to transfer the client to another server.
type TransferPacket struct {
	Host string
	Port int32
}

func (p *TransferPacket) ID(version string) uint32 { return 0x0B }
func (p *TransferPacket) State() State             { return StateConfiguration }
func (p *TransferPacket) ReadFrom(r io.Reader, version string) (err error) {
	// Server only
	return nil
}
func (p *TransferPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteString(w, p.Host); err != nil {
		return
	}
	err = WriteVarInt(w, p.Port)
	return
}

// --- Clientbound (Server -> Client) ---

// CustomPayloadPacket for configuration state.
type ConfigCustomPayloadPacket struct {
	Channel string
	Data    []byte
}

func (p *ConfigCustomPayloadPacket) ID(version string) uint32 { return 0x01 }
func (p *ConfigCustomPayloadPacket) State() State             { return StateConfiguration }
func (p *ConfigCustomPayloadPacket) ReadFrom(r io.Reader, version string) (err error) {
	// Server only
	return nil
}
func (p *ConfigCustomPayloadPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteString(w, p.Channel); err != nil {
		return
	}
	_, err = w.Write(p.Data)
	return
}

// ConfigDisconnectPacket disconnects the client.
type ConfigDisconnectPacket struct {
	Reason nbt.Tag
}

func (p *ConfigDisconnectPacket) ID(version string) uint32 { return 0x02 }
func (p *ConfigDisconnectPacket) State() State             { return StateConfiguration }
func (p *ConfigDisconnectPacket) ReadFrom(r io.Reader, version string) (err error) {
	// Server only
	return nil
}
func (p *ConfigDisconnectPacket) WriteTo(w io.Writer, version string) (err error) {
	return WriteNBT(w, p.Reason)
}

// FinishConfigurationPacket signals the client to move to the Play state.
type FinishConfigurationPacket struct{}

func (p *FinishConfigurationPacket) ID(version string) uint32                         { return 0x03 }
func (p *FinishConfigurationPacket) State() State                                     { return StateConfiguration }
func (p *FinishConfigurationPacket) ReadFrom(r io.Reader, version string) (err error) { return nil }
func (p *FinishConfigurationPacket) WriteTo(w io.Writer, version string) (err error)  { return nil }

// ConfigKeepAlivePacket is for checking if the connection is still active.
type ConfigKeepAlivePacket struct {
	KeepAliveID int64
}

func (p *ConfigKeepAlivePacket) ID(version string) uint32 { return 0x04 }
func (p *ConfigKeepAlivePacket) State() State             { return StateConfiguration }
func (p *ConfigKeepAlivePacket) ReadFrom(r io.Reader, version string) (err error) {
	// Server only
	return nil
}
func (p *ConfigKeepAlivePacket) WriteTo(w io.Writer, version string) (err error) {
	return WriteInt64(w, p.KeepAliveID)
}

// PingPacket for configuration state.
type ConfigPingPacket struct {
	PayloadID int32
}

func (p *ConfigPingPacket) ID(version string) uint32 { return 0x05 }
func (p *ConfigPingPacket) State() State             { return StateConfiguration }
func (p *ConfigPingPacket) ReadFrom(r io.Reader, version string) (err error) {
	// Server only
	return nil
}
func (p *ConfigPingPacket) WriteTo(w io.Writer, version string) (err error) {
	return WriteInt32(w, p.PayloadID)
}

// ResetChatPacket is a signal to clear chat.
type ResetChatPacket struct{}

func (p *ResetChatPacket) ID(version string) uint32                         { return 0x06 }
func (p *ResetChatPacket) State() State                                     { return StateConfiguration }
func (p *ResetChatPacket) ReadFrom(r io.Reader, version string) (err error) { return nil }
func (p *ResetChatPacket) WriteTo(w io.Writer, version string) (err error)  { return nil }

// RegistryDataPacket sends registry information to the client.
type RegistryDataPacket struct {
	RegistryID string // FIX: Renamed from ID to avoid conflict with Packet.ID() method.
	Entries    []struct {
		Key   string
		Value *nbt.CompoundTag // Optional
	}
}

func (p *RegistryDataPacket) ID(version string) uint32 { return 0x07 }
func (p *RegistryDataPacket) State() State             { return StateConfiguration }
func (p *RegistryDataPacket) ReadFrom(r io.Reader, version string) (err error) {
	// Server only
	return nil
}
func (p *RegistryDataPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteString(w, p.RegistryID); err != nil {
		return
	}
	if err = WriteVarInt(w, int32(len(p.Entries))); err != nil {
		return
	}
	for _, entry := range p.Entries {
		if err = WriteString(w, entry.Key); err != nil {
			return
		}
		hasValue := entry.Value != nil
		if err = WriteBool(w, hasValue); err != nil {
			return
		}
		if hasValue {
			if err = WriteNBT(w, &*entry.Value); err != nil {
				return
			}
		}
	}
	return
}

// RemoveResourcePackPacket tells the client to remove a resource pack.
type RemoveResourcePackPacket struct {
	HasUUID bool
	UUID    uuid.UUID
}

func (p *RemoveResourcePackPacket) ID(version string) uint32                         { return 0x08 }
func (p *RemoveResourcePackPacket) State() State                                     { return StateConfiguration }
func (p *RemoveResourcePackPacket) ReadFrom(r io.Reader, version string) (err error) { return nil } // Server only
func (p *RemoveResourcePackPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteBool(w, p.HasUUID); err != nil {
		return
	}
	if p.HasUUID {
		err = WriteUUID(w, p.UUID)
	}
	return
}

// AddResourcePackPacket tells the client to add a resource pack.
type AddResourcePackPacket struct {
	UUID          uuid.UUID
	URL           string
	Hash          string
	Forced        bool
	PromptMessage *nbt.Tag
}

func (p *AddResourcePackPacket) ID(version string) uint32                         { return 0x09 }
func (p *AddResourcePackPacket) State() State                                     { return StateConfiguration }
func (p *AddResourcePackPacket) ReadFrom(r io.Reader, version string) (err error) { return nil } // Server only
func (p *AddResourcePackPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteUUID(w, p.UUID); err != nil {
		return
	}
	if err = WriteString(w, p.URL); err != nil {
		return
	}
	if err = WriteString(w, p.Hash); err != nil {
		return
	}
	if err = WriteBool(w, p.Forced); err != nil {
		return
	}
	hasPrompt := p.PromptMessage != nil
	if err = WriteBool(w, hasPrompt); err != nil {
		return
	}
	if hasPrompt {
		err = WriteNBT(w, *p.PromptMessage)
	}
	return
}

// FeatureFlagsPacket sends the enabled feature flags to the client.
type FeatureFlagsPacket struct {
	Features []string
}

func (p *FeatureFlagsPacket) ID(version string) uint32                         { return 0x0C }
func (p *FeatureFlagsPacket) State() State                                     { return StateConfiguration }
func (p *FeatureFlagsPacket) ReadFrom(r io.Reader, version string) (err error) { return nil } // Server only
func (p *FeatureFlagsPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, int32(len(p.Features))); err != nil {
		return
	}
	for _, feature := range p.Features {
		if err = WriteString(w, feature); err != nil {
			return
		}
	}
	return
}

// TagsPacket updates client-side tags.
type TagsPacket struct {
	Tags []struct {
		TagType string
		Tags    []struct {
			TagName string
			Entries []int32
		}
	}
}

func (p *TagsPacket) ID(version string) uint32                         { return 0x0D }
func (p *TagsPacket) State() State                                     { return StateConfiguration }
func (p *TagsPacket) ReadFrom(r io.Reader, version string) (err error) { return nil } // Server only
func (p *TagsPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, int32(len(p.Tags))); err != nil {
		return
	}
	for _, tagGroup := range p.Tags {
		if err = WriteString(w, tagGroup.TagType); err != nil {
			return
		}
		if err = WriteVarInt(w, int32(len(tagGroup.Tags))); err != nil {
			return
		}
		for _, tag := range tagGroup.Tags {
			if err = WriteString(w, tag.TagName); err != nil {
				return
			}
			if err = WriteVarInt(w, int32(len(tag.Entries))); err != nil {
				return
			}
			for _, entry := range tag.Entries {
				if err = WriteVarInt(w, entry); err != nil {
					return
				}
			}
		}
	}
	return
}

// --- Serverbound (Client -> Server) ---

// ConfigCustomPayloadPacket for configuration state.
type ServerConfigCustomPayloadPacket struct {
	Channel string
	Data    []byte
}

func (p *ServerConfigCustomPayloadPacket) ID(version string) uint32 { return 0x02 }
func (p *ServerConfigCustomPayloadPacket) State() State             { return StateConfiguration }
func (p *ServerConfigCustomPayloadPacket) ReadFrom(r io.Reader, version string) (err error) {
	if p.Channel, err = ReadString(r); err != nil {
		return
	}
	p.Data, err = io.ReadAll(r)
	return
}
func (p *ServerConfigCustomPayloadPacket) WriteTo(w io.Writer, version string) (err error) {
	// Client only
	return nil
}

// FinishConfigurationAckPacket is the client's acknowledgement.
type FinishConfigurationAckPacket struct{}

func (p *FinishConfigurationAckPacket) ID(version string) uint32                         { return 0x03 }
func (p *FinishConfigurationAckPacket) State() State                                     { return StateConfiguration }
func (p *FinishConfigurationAckPacket) ReadFrom(r io.Reader, version string) (err error) { return nil }
func (p *FinishConfigurationAckPacket) WriteTo(w io.Writer, version string) (err error)  { return nil }

// ServerConfigKeepAlivePacket is the client's response to the server's keep alive.
type ServerConfigKeepAlivePacket struct {
	KeepAliveID int64
}

func (p *ServerConfigKeepAlivePacket) ID(version string) uint32 { return 0x04 }
func (p *ServerConfigKeepAlivePacket) State() State             { return StateConfiguration }
func (p *ServerConfigKeepAlivePacket) ReadFrom(r io.Reader, version string) (err error) {
	p.KeepAliveID, err = ReadInt64(r)
	return
}
func (p *ServerConfigKeepAlivePacket) WriteTo(w io.Writer, version string) (err error) {
	// Client only
	return nil
}

// PongPacket for configuration state.
type ConfigPongPacket struct {
	PayloadID int32
}

func (p *ConfigPongPacket) ID(version string) uint32 { return 0x05 }
func (p *ConfigPongPacket) State() State             { return StateConfiguration }
func (p *ConfigPongPacket) ReadFrom(r io.Reader, version string) (err error) {
	p.PayloadID, err = ReadInt32(r)
	return
}
func (p *ConfigPongPacket) WriteTo(w io.Writer, version string) (err error) {
	// Client only
	return nil
}

// ResourcePackReceivePacket indicates the client's resource pack status.
type ResourcePackReceivePacket struct {
	UUID   uuid.UUID
	Result int32
}

func (p *ResourcePackReceivePacket) ID(version string) uint32 { return 0x06 }
func (p *ResourcePackReceivePacket) State() State             { return StateConfiguration }
func (p *ResourcePackReceivePacket) ReadFrom(r io.Reader, version string) (err error) {
	if p.UUID, err = ReadUUID(r); err != nil {
		return
	}
	p.Result, err = ReadVarInt(r)
	return
}
func (p *ResourcePackReceivePacket) WriteTo(w io.Writer, version string) (err error) {
	// Client only
	return nil
}

// init registers the packets for the Configuration state.
func init() {
	// Shared / Common Packets
	RegisterPacket(StateConfiguration, Clientbound, 0x00, func() Packet { return &CookieRequestPacket{} })
	RegisterPacket(StateConfiguration, Serverbound, 0x01, func() Packet { return &CookieResponsePacket{} })

	// Clientbound (Server -> Client)
	RegisterPacket(StateConfiguration, Clientbound, 0x01, func() Packet { return &ConfigCustomPayloadPacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x02, func() Packet { return &ConfigDisconnectPacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x03, func() Packet { return &FinishConfigurationPacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x04, func() Packet { return &ConfigKeepAlivePacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x05, func() Packet { return &ConfigPingPacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x06, func() Packet { return &ResetChatPacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x07, func() Packet { return &RegistryDataPacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x08, func() Packet { return &RemoveResourcePackPacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x09, func() Packet { return &AddResourcePackPacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x0A, func() Packet { return &StoreCookiePacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x0B, func() Packet { return &TransferPacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x0C, func() Packet { return &FeatureFlagsPacket{} })
	RegisterPacket(StateConfiguration, Clientbound, 0x0D, func() Packet { return &TagsPacket{} })

	// Serverbound (Client -> Server)
	RegisterPacket(StateConfiguration, Serverbound, 0x00, func() Packet { return &ClientSettingsPacket{} })
	RegisterPacket(StateConfiguration, Serverbound, 0x02, func() Packet { return &ServerConfigCustomPayloadPacket{} })
	RegisterPacket(StateConfiguration, Serverbound, 0x03, func() Packet { return &FinishConfigurationAckPacket{} })
	RegisterPacket(StateConfiguration, Serverbound, 0x04, func() Packet { return &ServerConfigKeepAlivePacket{} })
	RegisterPacket(StateConfiguration, Serverbound, 0x05, func() Packet { return &ConfigPongPacket{} })
	RegisterPacket(StateConfiguration, Serverbound, 0x06, func() Packet { return &ResourcePackReceivePacket{} })
}

package protocol

import (
    "bytes"
    "crypto/rsa"
    "fmt"
    "github.com/google/uuid"
    "io"
    "net"
)

// State represents the current protocol state of a connection.
type State int

const (
    StateHandshaking State = iota
    StateStatus
    StateLogin
    StateConfiguration
    StatePlay
)

// Direction indicates whether a packet is clientbound or serverbound.
type Direction int

const (
    Serverbound Direction = iota // Client to Server
    Clientbound                  // Server to Client
)

// Packet is the interface that all packet structs must implement.
type Packet interface {
    ID(version string) uint32
    State() State
    ReadFrom(r io.Reader, version string) error
    WriteTo(w io.Writer, version string) error
}

// Conn represents a single Minecraft protocol connection.
type Conn struct {
    conn      net.Conn
    direction Direction
    version   string
    protocol  int32

    // Player/Connection Info
    Username    string
    UUID        uuid.UUID
    ProfileKeys *rsa.PublicKey

    // Pipeline components
    splitter *Splitter
    framer   *Framer

    // These are wrappers around the raw conn, they are swapped out as the state changes.
    reader io.Reader
    writer io.Writer

    // State Management
    state                State
    compressionThreshold int32
}

// NewConn creates a new connection wrapper.
func NewConn(conn net.Conn, dir Direction, version string) *Conn {
    c := &Conn{
        conn:                 conn,
        direction:            dir,
        version:              version,
        state:                StateHandshaking,
        compressionThreshold: -1, // Disabled
        reader:               conn,
        writer:               conn,
    }
    c.splitter = NewSplitter(c.reader)
    c.framer = NewFramer(c.writer)
    return c
}

// State returns the current protocol state.
func (c *Conn) State() State {
    return c.state
}

// SetState transitions the connection to a new protocol state.
func (c *Conn) SetState(newState State) {
    c.state = newState
}

// SetVersion sets the protocol version for the connection.
func (c *Conn) SetVersion(version string, protocol int32) {
    c.version = version
    c.protocol = protocol
}

// EnableEncryption enables AES/CFB8 encryption on the connection.
func (c *Conn) EnableEncryption(sharedSecret []byte) error {
    decipher, err := NewDecipher(c.conn, sharedSecret)
    if err != nil {
        return fmt.Errorf("failed to create decipher: %w", err)
    }
    c.reader = decipher

    cipher, err := NewCipher(c.conn, sharedSecret)
    if err != nil {
        return fmt.Errorf("failed to create cipher: %w", err)
    }
    c.writer = cipher

    // The splitter and framer must now use the encrypted stream.
    c.splitter = NewSplitter(c.reader)
    c.framer = NewFramer(c.writer)
    return nil
}

// EnableCompression enables zlib compression on the connection.
func (c *Conn) EnableCompression(threshold int32) {
    if threshold < 0 {
        // Disable compression
        c.compressionThreshold = -1
        // Reset reader/writer back to the original stream (or the crypto stream if enabled)
        // This logic assumes encryption is enabled before compression.
        baseReader, baseWriter := c.getBaseStreams()
        c.reader = baseReader
        c.writer = baseWriter
    } else {
        // Enable compression
        c.compressionThreshold = threshold
        baseReader, baseWriter := c.getBaseStreams()
        c.reader = NewDecompressor(baseReader, threshold)
        c.writer = NewCompressor(baseWriter, threshold)
    }

    // The splitter and framer must now use the (de)compressed stream.
    c.splitter = NewSplitter(c.reader)
    c.framer = NewFramer(c.writer)
}

// getBaseStreams returns the underlying streams before compression is applied.
// This is important because compression wraps the encryption streams, not the raw conn.
func (c *Conn) getBaseStreams() (io.Reader, io.Writer) {
    // Check if the current reader is a Decompressor, if so, get its underlying reader.
    if dc, ok := c.reader.(*Decompressor); ok {
        return dc.r, c.writer.(*Compressor).w
    }
    // Otherwise, it's either the raw conn or the crypto stream.
    return c.reader, c.writer
}

// ReadPacket reads, decodes, and returns the next packet from the connection.
func (c *Conn) ReadPacket() (Packet, error) {
    payload, err := c.splitter.NextPacket()
    if err != nil {
        return nil, fmt.Errorf("failed to read next packet frame: %w", err)
    }

    r := bytes.NewReader(payload)

    packetID, err := ReadVarInt(r)
    if err != nil {
        return nil, fmt.Errorf("failed to read packet ID: %w", err)
    }

    factory := GetPacketFactory(c.state, c.direction, uint32(packetID))
    if factory == nil {
        return nil, fmt.Errorf("unknown packet ID 0x%X in state %d for direction %d", packetID, c.state, c.direction)
    }

    packet := factory()
    if err := packet.ReadFrom(r, c.version); err != nil {
        return nil, fmt.Errorf("failed to read packet payload for 0x%X: %w", packetID, err)
    }

    return packet, nil
}

// WritePacket encodes and writes a packet to the connection.
func (c *Conn) WritePacket(packet Packet) error {
    var payload bytes.Buffer

    if err := WriteVarInt(&payload, int32(packet.ID(c.version))); err != nil {
        return fmt.Errorf("failed to write packet ID: %w", err)
    }

    if err := packet.WriteTo(&payload, c.version); err != nil {
        return fmt.Errorf("failed to write packet payload: %w", err)
    }

    return c.framer.WritePacket(payload.Bytes())
}

// Close closes the underlying network connection.
func (c *Conn) Close() error {
    return c.conn.Close()
}

package net

import (
	"bytes"
	"errors"
	"github.com/Advik-B/Golem/internal/player"
	"io"
	"log"
	"net"
)

const (
	HandshakeState = iota
	StatusState
	LoginState
	ConfigurationState
	PlayState
)

type Connection struct {
	conn  net.Conn
	state int
	r     *Reader
	w     *Writer
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn:  conn,
		state: HandshakeState,
		r:     NewReader(conn),
		w:     NewWriter(conn),
	}
}

func (c *Connection) HandleLogin(addPlayerFunc func(*player.Player)) (*player.Player, error) {
	if err := c.handleHandshake(); err != nil {
		return nil, err
	}

	if c.state == StatusState {
		c.handleStatus()
		return nil, errors.New("status packet handled")
	}

	if c.state != LoginState {
		return nil, errors.New("connection not in login state")
	}

	return c.handleLoginStart(addPlayerFunc)
}

func (c *Connection) HandleConfiguration() error {
	pktLen, err := c.r.ReadVarInt()
	if err != nil {
		return err
	}
	data := make([]byte, pktLen)
	if _, err := c.r.Read(data); err != nil {
		return err
	}

	// Packet ID should be 0x02 ServerboundFinishConfigurationPacket
	c.w.WritePacket(0x03) // ClientboundFinishConfigurationPacket
	c.state = PlayState
	return nil
}

func (c *Connection) HandlePlay(p *player.Player) {
	for {
		pktLen, err := c.r.ReadVarInt()
		if err != nil {
			if err != io.EOF {
				log.Printf("HandlePlay error for %s: %v", p.Username, err)
			}
			return
		}

		data := make([]byte, pktLen)
		if _, err := c.r.Read(data); err != nil {
			return
		}

		packetReader := NewReader(bytes.NewReader(data))
		pktID, _ := packetReader.ReadVarInt()

		// ServerboundKeepAlivePacket for 1.21 is 0x15
		if pktID == 0x15 {
			id, _ := packetReader.ReadLong()
			if id != p.LastKeepAliveID {
				log.Printf("Invalid keep-alive ID from %s. Got %d, expected %d.", p.Username, id, p.LastKeepAliveID)
			}
		}
	}
}

func (c *Connection) handleHandshake() error {
	pktLen, err := c.r.ReadVarInt()
	if err != nil {
		return err
	}
	data := make([]byte, pktLen)
	if _, err := c.r.Read(data); err != nil {
		return err
	}
	packetReader := NewReader(bytes.NewReader(data))
	_, err = packetReader.ReadVarInt() // Packet ID
	if err != nil {
		return err
	}

	_, _ = packetReader.ReadVarInt() // protocol
	_, _ = packetReader.ReadString() // host
	_, _ = packetReader.r.ReadByte() // port
	_, _ = packetReader.r.ReadByte() // port
	nextState, _ := packetReader.ReadVarInt()
	c.state = nextState
	return nil
}

func (c *Connection) handleStatus() {
	pktLen, _ := c.r.ReadVarInt()
	data := make([]byte, pktLen)
	c.r.Read(data)

	resp := `{"version":{"name":"Golem 1.21","protocol":767},"players":{"max":100,"online":0,"sample":[]},"description":{"text":"A Golem Server 🗿"}}`
	c.w.WritePacket(0x00, WriteString(resp))

	pktLen, _ = c.r.ReadVarInt()
	data = make([]byte, pktLen)
	c.r.Read(data)
	pingData := data[1:]
	c.w.WritePacket(0x01, pingData)
}

func (c *Connection) handleLoginStart(addPlayerFunc func(*player.Player)) (*player.Player, error) {
	pktLen, err := c.r.ReadVarInt()
	if err != nil {
		return nil, err
	}
	data := make([]byte, pktLen)
	if _, err := c.r.Read(data); err != nil {
		return nil, err
	}
	packetReader := NewReader(bytes.NewReader(data))
	_, _ = packetReader.ReadVarInt()

	username, _ := packetReader.ReadString()
	p := player.New(c.conn, username)

	// ClientboundLoginSuccessPacket (ID 0x02)
	c.w.WritePacket(0x02,
		WriteString("00000000-0000-0000-0000-000000000000"),
		WriteString(username),
		WriteVarInt(0),
	)

	addPlayerFunc(p)
	c.state = ConfigurationState // Transition connection state
	return p, nil
}

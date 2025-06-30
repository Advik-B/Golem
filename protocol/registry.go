package protocol

import "fmt"

// Protocol maps packet IDs to packet constructors for a given State and Flow.
type Protocol struct {
	// packetConstructors maps a packet ID to a function that returns a new, empty packet struct.
	packetConstructors map[int32]func() Packet
}

// New creates a new, empty protocol definition.
func New() *Protocol {
	return &Protocol{
		packetConstructors: make(map[int32]func() Packet),
	}
}

// Register registers a packet constructor for a given ID.
func (p *Protocol) Register(id int32, constructor func() Packet) {
	if _, exists := p.packetConstructors[id]; exists {
		panic(fmt.Sprintf("Packet with ID %#x is already registered", id))
	}
	p.packetConstructors[id] = constructor
}

// NewPacket creates a new instance of a packet based on its ID.
// It returns nil if the packet ID is not registered for this protocol.
func (p *Protocol) NewPacket(id int32) Packet {
	constructor, ok := p.packetConstructors[id]
	if !ok {
		return nil
	}
	return constructor()
}

// --- Global Protocol Registry ---

// Registry holds all defined protocols, indexed by State and then Flow.
// It is populated by the init() functions in the sub-packages.
var Registry map[State]map[Flow]*Protocol

func init() {
	Registry = make(map[State]map[Flow]*Protocol)
	for s := Handshaking; s <= Play; s++ {
		Registry[s] = make(map[Flow]*Protocol)
		Registry[s][Serverbound] = New()
		Registry[s][Clientbound] = New()
	}
	// NO PACKET REGISTRATIONS HERE
}

// GetProtocol returns the protocol definition for a given state and flow.
// It returns nil if no protocol is defined for the combination.
func GetProtocol(state State, flow Flow) *Protocol {
	if stateMap, ok := Registry[state]; ok {
		if proto, ok := stateMap[flow]; ok {
			return proto
		}
	}
	return nil
}

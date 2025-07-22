package protocol

import "fmt"

// PacketFactory is a function that creates a new, empty packet struct.
type PacketFactory func() Packet

// To avoid circular dependencies, packet structs will be in their own state-specific packages.
// We will define the maps here and populate them as we create the packet files.
var (
    serverboundFactories = make(map[State]map[uint32]PacketFactory)
    clientboundFactories = make(map[State]map[uint32]PacketFactory)
)

// RegisterPacket registers a packet's factory function for a given state, direction, and ID.
// This function will be called from the `init()` function of each packet file.
func RegisterPacket(state State, dir Direction, id uint32, factory PacketFactory) {
    var factoryMap map[State]map[uint32]PacketFactory
    if dir == Serverbound {
        factoryMap = serverboundFactories
    } else {
        factoryMap = clientboundFactories
    }

    if _, ok := factoryMap[state]; !ok {
        factoryMap[state] = make(map[uint32]PacketFactory)
    }
    factoryMap[state][id] = factory
}

// GetPacketFactory retrieves the correct PacketFactory for a given state, direction, and packet ID.
// This resolves the first unresolved reference in conn.go.
func GetPacketFactory(state State, dir Direction, id uint32) PacketFactory {
    var factoryMap map[State]map[uint32]PacketFactory
    if dir == Serverbound {
        factoryMap = serverboundFactories
    } else {
        factoryMap = clientboundFactories
    }

    stateFactories, ok := factoryMap[state]
    if !ok {
        return nil // No packets registered for this state
    }

    factory, ok := stateFactories[id]
    if !ok {
        return nil // No packet registered for this ID in this state
    }
    return factory
}

// For debugging purposes
func PrintRegistry() {
    fmt.Println("--- Serverbound Packet Registry ---")
    for state, packets := range serverboundFactories {
        fmt.Printf("State %d:\n", state)
        for id := range packets {
            fmt.Printf("  - 0x%02X\n", id)
        }
    }
    fmt.Println("--- Clientbound Packet Registry ---")
    for state, packets := range clientboundFactories {
        fmt.Printf("State %d:\n", state)
        for id := range packets {
            fmt.Printf("  - 0x%02X\n", id)
        }
    }
}

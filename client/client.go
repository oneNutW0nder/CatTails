package main

import (
	"fmt"
	"log"
	"net"

	"github.com/mdlayher/ethernet"
	"github.com/mdlayher/raw"
)

func newEthernetFrame(payload string) *ethernet.Frame {
	// The frame to be sent over the network.
	f := &ethernet.Frame{
		// Broadcast frame to all machines on same network segment.
		Destination: ethernet.Broadcast,
		// Identify our machine as the sender.
		Source: net.HardwareAddr{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad},
		// Identify frame with an unused EtherType.
		EtherType: 0xcccc,
		// Send a simple message.
		Payload: []byte(payload),
	}

	return f
}

func main() {
	// Select the eth0 interface to use for Ethernet traffic.
	ifi, err := net.InterfaceByName("lo")
	fmt.Println(net.Interfaces())
	if err != nil {
		log.Fatalf("failed to open interface: %v", err)
	}

	// Open a raw socket using same EtherType as our frame.
	c, err := raw.ListenPacket(ifi, 0xcccc, nil)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer c.Close()

	// Marshal a frame to its binary format.
	f := newEthernetFrame("hello world")
	b, err := f.MarshalBinary()
	if err != nil {
		log.Fatalf("failed to marshal frame: %v", err)
	}

	// Broadcast the frame to all devices on our network segment.
	addr := &raw.Addr{HardwareAddr: ethernet.Broadcast}
	if _, err := c.WriteTo(b, addr); err != nil {
		log.Fatalf("failed to write frame: %v", err)
	}
}

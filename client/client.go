package main

import (
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/mdlayher/ethernet"
	"github.com/mdlayher/raw"
)

func newEthernetFrame() *ethernet.Frame {
	// The frame to be sent over the network.
	f := &ethernet.Frame{
		// Broadcast frame to all machines on same network segment.
		Destination: ethernet.Broadcast,
		// Identify our machine as the sender.
		Source: net.HardwareAddr{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad},
		// Identify frame with an unused EtherType.
		EtherType: 0xcccc,
		// Data is going to be layers 3-4
		Payload: []byte(createPacket()),
	}

	return f
}

func createPacket() []byte {
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	gopacket.SerializeLayers(buf, opts,
		// Ethernet layer
		&layers.Ethernet{
			EthernetType: layers.EthernetTypeIPv4,
			SrcMAC: net.HardwareAddr{
				0xde, 0xad, 0xbe, 0xef, 0xde, 0xad,
			},
			DstMAC: net.HardwareAddr{
				0xde, 0xad, 0xbe, 0xef, 0xde, 0xad,
			},
			Length: 0,
		},
		// IPv4 layer
		&layers.IPv4{
			Version:    0x4,
			IHL:        5,
			Length:     28,
			TTL:        255,
			Flags:      0x40,
			FragOffset: 0,
			Checksum:   0,
			Protocol:   syscall.IPPROTO_UDP,
			DstIP:      net.IPv4(127, 0, 0, 1),
			SrcIP:      net.IPv4(127, 0, 0, 1),
		},
		// UDP layer
		&layers.UDP{
			SrcPort:  6969,
			DstPort:  9696,
			Length:   8,
			Checksum: 0,
		},
		gopacket.Payload("Encapsulation work"))
	packetData := buf.Bytes()

	return packetData
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
	//f := newEthernetFrame()
	//b, err := f.MarshalBinary()
	if err != nil {
		log.Fatalf("failed to marshal frame: %v", err)
	}

	// Broadcast the frame to all devices on our network segment.
	addr := &raw.Addr{HardwareAddr: ethernet.Broadcast}
	if _, err := c.WriteTo(createPacket(), addr); err != nil {
		log.Fatalf("failed to write frame: %v", err)
	}
}

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

func createPacket() []byte {
	// Create a new seriablized buffer
	buf := gopacket.NewSerializeBuffer()
	// Generate options
	opts := gopacket.SerializeOptions{}
	// Serialize layers
	// This builds/encapsulates the layers of a packet properly
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
			Length:     46,
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
			Length:   26,
			Checksum: 0, // TODO
		},
		// Set the payload
		gopacket.Payload("Encapsulation work"),
	)
	// Save the newly formed packet and return it
	packetData := buf.Bytes()

	return packetData
}

func main() {
	// Select the interface from your device (loopback for testing)
	ifi, err := net.InterfaceByName("lo")
	if err != nil {
		log.Fatalf("failed to open interface: %v", err)
	}
	// Prints out info about all interfaces on device
	fmt.Println(net.Interfaces())

	// Open a raw socket using same EtherType as our frame.
	c, err := raw.ListenPacket(ifi, 0xcccc, nil)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer c.Close()

	// Broadcast the frame to all devices on our network segment.
	addr := &raw.Addr{HardwareAddr: ethernet.Broadcast}
	if _, err := c.WriteTo(createPacket(), addr); err != nil {
		log.Fatalf("failed to write frame: %v", err)
	}
}

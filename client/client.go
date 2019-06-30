package main

import (
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Function to do this err checking repeatedly
func checkEr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// htons converts a short (uint16) from host-to-network byte order.
// #Stackoverflow
func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}

// This function takes a net.Interface pointer to access
// 	things like the MAC Address... and yeah... the MAC Address
//
// ifaceInfo	--> pointer to a net.Interface
//
// Returns		--> Byte array that is a properly formed/serialized packet
func createPacket(ifaceInfo *net.Interface) []byte {

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
			SrcMAC:       ifaceInfo.HardwareAddr,
			DstMAC: net.HardwareAddr{
				0x88, 0xb1, 0x11, 0x58, 0xf7, 0x3c,
			},
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
			Protocol:   syscall.IPPROTO_UDP, // Sending a UDP Packet
			DstIP:      net.IPv4(192, 168, 1, 57),
			SrcIP:      net.IPv4(192, 168, 1, 57),
		},
		// UDP layer
		&layers.UDP{
			SrcPort:  6969,
			DstPort:  layers.UDPPort(1337), // Saw this used in some code @github... seems legit
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

// This function sends a packet using a provided
//	socket file descriptor (fd)
//
// fd 		--> The file descriptor for the socket to use
//
// Returns 	--> None
func sendPacket(fd int) {

	// Get *Interface struct for the interface that we are using
	ifaceInfo, _ := net.InterfaceByName("wlp4s0")

	// Create a byte array for the MAC Addr
	var haddr [8]byte

	// Copy the MAC from the interface struct in the new array
	copy(haddr[0:7], ifaceInfo.HardwareAddr[0:7])

	// Initialize the Sockaddr struct
	addr := syscall.SockaddrLinklayer{
		Protocol: syscall.ETH_P_IP,
		Ifindex:  ifaceInfo.Index,
		Halen:    uint8(len(ifaceInfo.HardwareAddr)),
		Addr:     haddr,
	}

	// Bind the socket
	checkEr(syscall.Bind(fd, &addr))

	// Set promiscuous mode = true
	checkEr(syscall.SetLsfPromisc("wlp4s0", true))

	// Send a packet using our socket
	// n --> number of bytes sent
	n, err := syscall.Write(fd, createPacket(ifaceInfo))
	checkEr(err)

	// Debug shenanigans
	fmt.Printf("Number of bytes written: %d", n)
}

func main() {

	// Create the socket
	// AF_PACKET 	--> Low level packet interface access
	// SOCK_RAW		--> Socket type is RAWe
	// ETH_P_ALL	--> Handle all Ethernet frames that come
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	checkEr(err)

	// REEEEEEEEEEEEEEEEEEEEE
	for {
		sendPacket(fd)
	}
}

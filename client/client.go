package main

import (
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func checkEr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// htons converts a short (uint16) from host-to-network byte order.
func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}

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
			Protocol:   syscall.IPPROTO_UDP,
			DstIP:      net.IPv4(192, 168, 1, 57),
			SrcIP:      net.IPv4(192, 168, 1, 57),
		},
		// UDP layer
		&layers.UDP{
			SrcPort:  6969,
			DstPort:  layers.UDPPort(1337),
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

func sendPacket(fd int) {
	// Get *Interface struct for the interface that we are using
	info, _ := net.InterfaceByName("wlp4s0")

	var haddr [8]byte
	copy(haddr[0:7], info.HardwareAddr[0:7])
	addr := syscall.SockaddrLinklayer{
		Protocol: syscall.ETH_P_IP,
		Ifindex:  info.Index,
		Halen:    uint8(len(info.HardwareAddr)),
		Addr:     haddr,
	}

	checkEr(syscall.Bind(fd, &addr))
	checkEr(syscall.SetLsfPromisc("wlp4s0", true))
	n, err := syscall.Write(fd, createPacket(info))
	checkEr(err)

	fmt.Printf("Number of bytes written: %d", n)

	//fmt.Println(info)
	//fmt.Println(info.HardwareAddr)
}

func main() {

	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	checkEr(err)
	for {
		sendPacket(fd)
	}
}

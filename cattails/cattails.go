package cattails

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/bpf"
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

// ReadPacket reads packets from a socket file descriptor (fd)
//
// fd  	--> file descriptor that relates to the socket created in main
// vm 	--> BPF VM that contains the BPF Program
//
// Returns 	--> None
func ReadPacket(fd int, vm *bpf.VM) {

	// Buffer for packet data that is read in
	buf := make([]byte, 1500)

	for {
		// Read in the packets
		// num 		--> number of bytes
		// sockaddr --> the sockaddr struct that the packet was read from
		// err 		--> was there an error?
		_, _, err := syscall.Recvfrom(fd, buf, 0)
		checkEr(err)

		// Filter packet?
		// numBytes	--> Number of bytes
		// err	--> Error you say?
		numBytes, err := vm.Run(buf)
		checkEr(err)
		if numBytes == 0 {
			continue // 0 means that the packet should be dropped
			// Here we are just "ignoring" the packet and moving on to the next one
		}
		fmt.Println(numBytes)

		// Parse packet... hopefully
		packet := gopacket.NewPacket(buf, layers.LayerTypeEthernet, gopacket.Default)
		if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			fmt.Printf("Data in UDP packet is: %d", udp.Payload)
		}
	}
}

// SendPacket sends a packet using a provided
//	socket file descriptor (fd)
//
// fd 		--> The file descriptor for the socket to use
//
// Returns 	--> None
func SendPacket(fd int, ifaceInfo *net.Interface, packetData []byte) {

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
	checkEr(syscall.SetLsfPromisc(ifaceInfo.Name, true))

	// Send a packet using our socket
	// n --> number of bytes sent
	_, err := syscall.Write(fd, packetData)
	checkEr(err)
}

// CreatePacket takes a net.Interface pointer to access
// 	things like the MAC Address... and yeah... the MAC Address
//
// ifaceInfo	--> pointer to a net.Interface
//
// Returns		--> Byte array that is a properly formed/serialized packet
func CreatePacket(ifaceInfo *net.Interface, srcIp net.IP,
	dstIP net.IP, dstMAC net.HardwareAddr, payload string) (packetData []byte) {

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
			DstMAC:       dstMAC,
		},
		// IPv4 layer
		&layers.IPv4{
			Version:    0x4,
			IHL:        5,
			Length:     uint16(20 + 8 + len(payload)), // 20 = IP header; 8 = UDP header; payload = ?
			TTL:        255,
			Flags:      0x40,
			FragOffset: 0,
			Checksum:   0,                   // Wireshark does not blatantly complain about this so it is on the backlog...
			Protocol:   syscall.IPPROTO_UDP, // Sending a UDP Packet
			DstIP:      dstIP,               //net.IPv4(192, 168, 1, 57),
			SrcIP:      srcIp,               //net.IPv4(192, 168, 1, 57),
		},
		// UDP layer
		&layers.UDP{
			SrcPort:  6969,
			DstPort:  layers.UDPPort(1337), // Saw this used in some code @github... seems legit
			Length:   uint16(8 + len(payload)),
			Checksum: 0, // TODO
		},
		// Set the payload
		gopacket.Payload(payload),
	)

	// Save the newly formed packet and return it
	packetData = buf.Bytes()

	return packetData
}

// CreateBPFVM creates a BPF VM that contains a BPF program
// 	given by the user in the form of "[]bpf.RawInstruction".
// You can create this by using "tcpdump -dd [your filter here]"
//
// filter	--> Raw BPF instructions generated from tcpdump
//
// Returns	--> Pointer to a BPF VM containing the filter/program
func CreateBPFVM(filter []bpf.RawInstruction) (vm *bpf.VM) {

	insts, allDecoded := bpf.Disassemble(filter)
	if allDecoded != true {
		log.Fatal("Error decoding BPF instructions...")
	}

	vm, err := bpf.NewVM(insts)
	checkEr(err)

	return vm
}

// NewSocket creates a new RAW socket and returns the file descriptor
//
// Returns --> File descriptor for the raw socket
func NewSocket() (fd int) {

	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	checkEr(err)

	return fd
}

// GetOutboundIP finds the outbound IP addr for the machine
//
// Returns	--> IP address in form "XXX.XXX.XXX.XXX"
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	checkEr(err)

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// GetOutwardIface determines the interface associated with
// sending traffic out on the wire and returns a *net.Interface struct
//
// Returns	--> *net.Interface struct of outward interface
//			--> net.IP used for creating a packet
func GetOutwardIface() (byNameiface *net.Interface, ip net.IP) {
	outboundIP := getOutboundIP()

	ifaces, err := net.Interfaces()
	checkEr(err)

	for _, i := range ifaces {

		byNameiface, err := net.InterfaceByName(i.Name)
		checkEr(err)

		addrs, _ := byNameiface.Addrs()

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if bytes.Compare(outboundIP, ipnet.IP.To4()) == 0 {
					ip := ipnet.IP.To4()
					return byNameiface, ip
				}
			}
		}
	}

	return
}

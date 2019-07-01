package main

import (
	"fmt"
	"log"
	"syscall"

	"golang.org/x/net/bpf"

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

// Reads the packet in from a socket file descriptor (fd)
//
// fd  	--> file descriptor that relates to the socket created in main
// vm 	--> BPF VM that contains the BPF Program
//
// Returns 	--> None
func readPacket(fd int, vm *bpf.VM) {

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
	// Debug shenanigans
	// fmt.Println(num, "bytes from", sockaddr)
	// fmt.Println(buf)
}

func main() {

	// Create the socket
	// AF_PACKET 	--> Low level packet interface access
	// SOCK_RAW		--> Socket type is RAWe
	// ETH_P_ALL	--> Handle all Ethernet frames that com
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	checkEr(err)

	// sudo tcpdump -dd udp and port 6969
	filterRaw := []bpf.RawInstruction{
		{0x28, 0, 0, 0x0000000c},
		{0x15, 0, 6, 0x000086dd},
		{0x30, 0, 0, 0x00000014},
		{0x15, 0, 15, 0x00000011},
		{0x28, 0, 0, 0x00000036},
		{0x15, 12, 0, 0x00001b39},
		{0x28, 0, 0, 0x00000038},
		{0x15, 10, 11, 0x00001b39},
		{0x15, 0, 10, 0x00000800},
		{0x30, 0, 0, 0x00000017},
		{0x15, 0, 8, 0x00000011},
		{0x28, 0, 0, 0x00000014},
		{0x45, 6, 0, 0x00001fff},
		{0xb1, 0, 0, 0x0000000e},
		{0x48, 0, 0, 0x0000000e},
		{0x15, 2, 0, 0x00001b39},
		{0x48, 0, 0, 0x00000010},
		{0x15, 0, 1, 0x00001b39},
		{0x6, 0, 0, 0x00040000},
		{0x6, 0, 0, 0x00000000},
	}

	// Creates the ASM instructions that will allow passing to BPF VM
	insts, allDecoded := bpf.Disassemble(filterRaw)

	// Debug stuffs
	fmt.Println(insts)      // Should be ASM instructions for filter
	fmt.Println(allDecoded) // Should be true

	vm, err := bpf.NewVM(insts)
	checkEr(err)

	// REEEEEEEEEEEEEEEEEEEEEEEEEE
	readPacket(fd, vm)

}

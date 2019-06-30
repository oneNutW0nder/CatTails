package main

import (
	"fmt"
	"log"
	"syscall"
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
// fd int 	--> file descriptor that relates to the socket created in main
//
// Returns 	--> None
func readPacket(fd int) {

	// Buffer for packet data that is read in
	buf := make([]byte, 1500)

	// Read in the packets
	// num 		--> number of bytes
	// sockaddr --> the sockaddr struct that the packet was read from
	// err 		--> was there an error?
	num, sockaddr, err := syscall.Recvfrom(fd, buf, 0)
	checkEr(err)

	// Debug shenanigans
	fmt.Println(num, "bytes from", sockaddr)
	fmt.Println(buf)
}

func main() {

	// Create the socket
	// AF_PACKET 	--> Low level packet interface access
	// SOCK_RAW		--> Socket type is RAWe
	// ETH_P_ALL	--> Handle all Ethernet frames that com
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	checkEr(err)

	for {
		// REEEEEEEEEEEEEEEEEEEEEEEEEE
		readPacket(fd)
	}

}

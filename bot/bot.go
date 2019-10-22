package main

import (
	"net"
	"syscall"

	"github.com/oneNutW0nder/CatTails/cattails"
)

func sendHello(fd int, iface *net.Interface, src net.IP, dst net.IP, dstMAC net.HardwareAddr) {
	packet := cattails.CreatePacket(iface, src, dst, dstMAC, cattails.CreateHello(iface.HardwareAddr, src))

	addr := cattails.CreateAddrStruct(iface)

	cattails.SendPacket(fd, iface, addr, packet)
}

func main() {

	fd := cattails.NewSocket()
	defer syscall.Close(fd)

	iface, src := cattails.GetOutwardIface("8.8.8.8:80")

	//vm := cattails.CreateBPFVM(filterRaw)

	dstMAC, _ := cattails.GetRouterMAC()

	sendHello(fd, iface, src, net.IPv4(129, 21, 117, 56), dstMAC)

}

package main

import (
	"fmt"
	"net"
	"syscall"
	"time"

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
	for x := 1; x < 5; x++ {
		fmt.Println("Sending number:", x)
		time.Sleep(time.Second * 5)
		sendHello(fd, iface, src, net.IPv4(192, 168, 68, 129), dstMAC)
	}
	//fmt.Println("Gateway MAC used:", dstMAC)

	//fmt.Print(cattails.CreateHello(iface.HardwareAddr, srcIp))

	//packet := cattails.ReadPacket(fd, vm)
}

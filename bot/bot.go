package main

import (
	"net"

	"github.com/oneNutW0nder/CatTails/cattails"
)

func main() {

	fd := cattails.NewSocket()
	iface, srcIp := cattails.GetOutwardIface("8.8.8.8:80")

	//vm := cattails.CreateBPFVM(filterRaw)

	dstMAC, _ := cattails.GetRouterMAC()
	//fmt.Println("Gateway MAC used:", dstMAC)

	//fmt.Print(cattails.CreateHello(iface.HardwareAddr, srcIp))

	packet := cattails.CreatePacket(iface, srcIp, net.IPv4(192, 168, 68, 129), dstMAC, cattails.CreateHello(iface.HardwareAddr, srcIp))
	addr := cattails.CreateAddrStruct(iface)

	cattails.SendPacket(fd, iface, addr, packet)
	//packet := cattails.ReadPacket(fd, vm)
}

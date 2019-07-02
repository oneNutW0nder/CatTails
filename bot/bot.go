package main

import (
	"net"

	"github.com/oneNutW0nder/CatTails/cattails"
	"golang.org/x/net/bpf"
)

func main() {

	filterRaw := []bpf.RawInstruction{
		{0x28, 0, 0, 0x0000000c},
		{0x15, 0, 6, 0x000086dd},
		{0x30, 0, 0, 0x00000014},
		{0x15, 0, 15, 0x00000011},
		{0x28, 0, 0, 0x00000036},
		{0x15, 12, 0, 0x00001b39},
		{0x28, 0, 0, 0x00000038},
		{0x15, 10, 11, 0x00001b39},
		{0x15, 0, 10, 0x00000800},
		{0x30, 0, 0, 0x00000017},
		{0x15, 0, 8, 0x00000011},
		{0x28, 0, 0, 0x00000014},
		{0x45, 6, 0, 0x00001fff},
		{0xb1, 0, 0, 0x0000000e},
		{0x48, 0, 0, 0x0000000e},
		{0x15, 2, 0, 0x00001b39},
		{0x48, 0, 0, 0x00000010},
		{0x15, 0, 1, 0x00001b39},
		{0x6, 0, 0, 0x00040000},
		{0x6, 0, 0, 0x00000000},
	}

	fd := cattails.NewSocket()
	iface, srcIp := cattails.GetOutwardIface()

	//vm := cattails.CreateBPFVM(filterRaw)

	dstMAC := net.HardwareAddr{
		0x88, 0xb1, 0x11, 0x58, 0xf7, 0x3c,
	}

	packet := cattails.CreatePacket(iface, srcIp, net.IPv4(192, 168, 68, 120), dstMAC, "REEEEEEEEEEEEEEE")

	cattails.SendPacket(fd, iface, packet)

}

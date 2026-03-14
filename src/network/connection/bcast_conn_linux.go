//go:build linux
// +build linux

package connection

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

func DialBroadcastUDP(port int) net.PacketConn {
	socket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil { fmt.Println("Error: Socket:", err) }

	syscall.SetsockoptInt(socket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil { fmt.Println("Error: SetSockOpt REUSEADDR:", err) }

	syscall.SetsockoptInt(socket, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil { fmt.Println("Error: SetSockOpt BROADCAST:", err) }

	syscall.Bind(socket, &syscall.SockaddrInet4{Port: port})
	if err != nil { fmt.Println("Error: Bind:", err) }

	file := os.NewFile(uintptr(socket), "")
	connection, err := net.FilePacketConn(file)
	if err != nil { fmt.Println("Error: FilePacketConn:", err) }
	file.Close()

	return connection
}

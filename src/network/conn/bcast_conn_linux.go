// +build linux

package conn

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

func DialBroadcastUDP(port int) net.PacketConn {
	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil { fmt.Println("Error: Socket:", err) }
	syscall.SetsockoptInt(sock, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil { fmt.Println("Error: SetSockOpt REUSEADDR:", err) }
	syscall.SetsockoptInt(sock, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil { fmt.Println("Error: SetSockOpt BROADCAST:", err) }
	syscall.Bind(sock, &syscall.SockaddrInet4{Port: port})
	if err != nil { fmt.Println("Error: Bind:", err) }

	f := os.NewFile(uintptr(sock), "")
	conn, err := net.FilePacketConn(f)
	if err != nil { fmt.Println("Error: FilePacketConn:", err) }
	f.Close()

	return conn
}

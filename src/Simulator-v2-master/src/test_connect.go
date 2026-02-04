package main 
/* 
To communicate between 2 programs running in different terminals using UDP write this in your terminal

Skriv dette i terminal 1: go run test_connect.go 30001 30002
og dette i terminal 2: go run test_connect.go 30002 30001


*/


import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run udp_node.go <myPort> <peerPort>")
		return
	}

	myPort, _ := strconv.Atoi(os.Args[1])
	peerPort, _ := strconv.Atoi(os.Args[2])

	// --- Listen socket ---
	localAddr := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: myPort,
	}

	conn, err := net.ListenUDP("udp", &localAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Printf("Listening on port %d\n", myPort)

	// --- Goroutine: receive messages ---
	go func() {
		buf := make([]byte, 1024)
		for {
			n, from, err := conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Read error:", err)
				continue
			}
			fmt.Printf("Received from %d: %s\n", from.Port, string(buf[:n]))
		}
	}()

	// --- Send socket ---
	peerAddr := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: peerPort,
	}

	counter := 0
	for {
		msg := fmt.Sprintf("Hello from %d (%d)", myPort, counter)
		_, err := conn.WriteToUDP([]byte(msg), &peerAddr)
		if err != nil {
			fmt.Println("Send error:", err)
		} else {
			fmt.Printf("Sent to %d: %s\n", peerPort, msg)
		}
		counter++
		time.Sleep(1 * time.Second)
	}
}

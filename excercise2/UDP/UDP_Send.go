package main

import (
	"net"
	"time"
)

func send() {
	remoteIP := net.ParseIP("10.100.23.11")
	remotePort := 20003

	addr := &net.UDPAddr{
		IP:   remoteIP,
		Port: remotePort,
	}

	// "connect" i Go = net.DialUDP
	sock, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer sock.Close()

	message := []byte("hello from station 3")

	// send()
	_, err = sock.Write(message)
	if err != nil {
		panic(err)
	}
	time.Sleep(100 * time.Millisecond)
}
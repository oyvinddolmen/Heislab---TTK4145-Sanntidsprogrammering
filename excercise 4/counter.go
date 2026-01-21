package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
)

//hei

func backup() {

	current_number := string(0)

	// the address we are listening for messages on
	// we have no choice in IP, so use 0.0.0.0, INADDR_ANY, or leave the IP field empty
	// the port should be whatever the sender sends to
	// alternate names: sockaddr, resolve(udp)addr,
	addr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 20003,
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	buffer := make([]byte, 1024)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	// ReadFromUDP blokkerer, for løkken sammen med ReadFromUDP gjør at vi fortsetter å lytte dersom vi ikke får en pakke
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		if err != nil {
			//fmt.Println(err)

			conn.Close()
			primary(current_number)
			continue
		}

		current_number = string(buffer[:n])

		//fmt.Println("Mottatt melding:", current_number)
		//fmt.Println("Server IP:", fromWho.IP.String())

	}

}

func primary(str string) {

	cmd := exec.Command(
		"gnome-terminal",
		"--",
		"bash",
		"-c",
		"go run counter.go; exec bash",
	)

	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	time.Sleep(400 * time.Millisecond)
	remoteIP := net.ParseIP("127.0.0.1")
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

	current_value, err := strconv.Atoi(str)
	for i := current_value + 1; i <= 100; i++ {

		message := []byte(strconv.Itoa(i))

		fmt.Println(i)
		// send()
		_, err = sock.Write(message)
		if err != nil {
			panic(err)
		}
		time.Sleep(500 * time.Millisecond)
	}

}

func main() {
	fmt.Println("PC 1 starting")
	backup()

}

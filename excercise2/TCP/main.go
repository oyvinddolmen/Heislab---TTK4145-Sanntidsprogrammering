package main

import (
	"fmt"
	"net"
	"time"
)

// 10.1.197.192 vår ip

func readTCP(conn net.Conn) {
	// ReadsM SERVER - TCP

	fmt.Println(("Lytter..."))

	// a buffer where the received network data is stored
	buffer := make([]byte, 1024)

	// You may not be able to use the same port twice when you restart the program, unless you set this option

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Feil:", err)
			continue
		}

		message := string(buffer[:n])

		fmt.Println("Melding fra server:", message)
		//fmt.Println("Server IP:", fromWho.IP.String())

		time.Sleep(400 * time.Millisecond)
		// Når du har funnet IP-en, kan du avslutte
		break
	}

}

func sendTCP(conn net.Conn) {
	// WRITE TO SERVER - TCP

	fmt.Println(("Skriver til server"))

	_, err2 := conn.Write([]byte("Hello from station 3\x00"))

	if err2 != nil {
		panic(err2)
	}

	time.Sleep(100 * time.Millisecond)

}

func createConn() net.Conn {
	conn, err := net.Dial("tcp", "10.100.23.11:33546")
	if err != nil {
		panic(err)
	}

	return conn
}

func main() {

	conn := createConn()
	defer conn.Close() // closes the connection at the end of the main func

	ready := make(chan struct{})
	
	go func() {
		readTCP(conn)
		ready <- struct{}{} // signaler at lesing er ferdig
	}()

	<-ready // vent til readTCP har startet eller gjort noe
	
	
	go func() {
		sendTCP(conn)
		ready <- struct{}{} // signaler at lesing er ferdig
	}()

	<-ready

	go readTCP(conn)

	select {}
}

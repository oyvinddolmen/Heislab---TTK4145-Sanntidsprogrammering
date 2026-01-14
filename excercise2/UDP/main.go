package main

func main() {
	go send()
	go receive()

	// Hindrer at main avslutter
	select {}
}

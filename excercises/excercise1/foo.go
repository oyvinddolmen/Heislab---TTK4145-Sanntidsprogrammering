// Use `go run foo.go` to run your program

package main

import (
	//"fmt"
	. "fmt"
	"runtime"
	//"time"
)

var i = 0

func incrementing(c chan int, done chan int) {
	for range 1000000 {
		c <- 1
	}

	done <- 1
}

func decrementing(c chan int, done chan int) {
	for range 1000001 {
		c <- -1
	}

	done <- 1
}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1? ---- antall prosessorer som kjører parallellt
	runtime.GOMAXPROCS(2)

	c := make(chan int)
	done := make(chan int)

	// TODO: Spawn both functions as goroutines
	go incrementing(c, done)
	go decrementing(c, done)

	// venter på done signaler fra inkrementeringen og dekrementeringen før den lukker c
	go func() {
		<-done
        <-done

		close(c)
	}()

	// looper over antall sendinger til kanalen
	for v := range c {
		i += v
		// fmt.Println(v)
	}

	// We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
	// We will do it properly with channels soon. For now: Sleep.
	//time.Sleep(500 * time.Millisecond)
	Println("The magic number is:", i)
}

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"time"
)

// sleep simulates each thread taking a longer time to process the input.
func sleep(id int) {
	if id == 0 {
		time.Sleep(200 * time.Millisecond)
	}
	if id == 1 {
		time.Sleep(800 * time.Millisecond)
	}
	if id == 2 {
		time.Sleep(2 * time.Second)
	}
}

// processInput simulates each thread processing the perceived keyboard
// signals.
//
// Each thread writes the output to a file in order to avoid cluttering the
// standard input (and also note the main thread reads keyboard input there).
func processInput(id int, c <-chan byte, wg *sync.WaitGroup, w *os.File) {

	defer w.Close()
	defer wg.Done()

	q := []byte{}

	stillListening := true
	for {
		select {
		case character, ok := <-c:
			if !ok {
				c = nil
				stillListening = false
			} else {
				q = append(q, character)
			}
		default:
			if len(q) != 0 {
				top := q[:1]
				q = q[1:]
				sleep(id)
				w.Write(top)
			} else {
				if !stillListening {
					fmt.Printf("thread %d done\n", id)
					return
				}
			}
		}
	}
}

func initthreads() (channels [3]chan byte, writers [3]*os.File) {

	// basically unlimited size channels
	c0 := make(chan byte, 100)
	c1 := make(chan byte, 100)
	c2 := make(chan byte, 100)

	f0, err0 := os.OpenFile("0.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	f1, err1 := os.OpenFile("1.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	f2, err2 := os.OpenFile("2.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)

	if err0 != nil || err1 != nil || err2 != nil {
		panic("error opening files")
	}

	channels[0] = c0
	channels[1] = c1
	channels[2] = c2

	writers[0] = f0
	writers[1] = f1
	writers[2] = f2
	return
}

func main() {
	wg := &sync.WaitGroup{}
	wg.Add(3)

	c, w := initthreads()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	go processInput(0, c[0], wg, w[0])
	go processInput(1, c[1], wg, w[1])
	go processInput(2, c[2], wg, w[2])
	r := bufio.NewReader(os.Stdin)
	for {
		chrs, _, err := r.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				panic(err)
			}
		}
		for _, chr := range chrs {
			c[0] <- chr
			c[1] <- chr
			c[2] <- chr
		}
	}

	close(c[0])
	close(c[1])
	close(c[2])

	fmt.Println("waiting for other threads to complete...")
	wg.Wait()

	fmt.Println("done")
}

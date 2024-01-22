package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mr-joshcrane/siphon"
)

func main() {
	q, err := siphon.NewAWSQueue("testEvents")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	r := bufio.NewReader(os.Stdin)
	input := make(chan string, 1_000_000)
	go func() {
		for {
			text, err := r.ReadString('\n')
			if err == io.EOF {
				input <- text
				os.Exit(0)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			input <- text
		}
	}()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			select {
			case message := <-input:
				err := q.Enqueue(message)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				fmt.Println("sent message")
			default:
				// Do nothing if no message is available
			}
		}
	}
}

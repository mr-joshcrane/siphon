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
	input := make(chan string)
	go func() {
		for {
			text, err := r.ReadString('\n')
			if err == io.EOF {
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
	for range ticker.C {
		message := <-input
		err := q.Enqueue(message)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}
}

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
	input := make(chan string, 1000)
	go func() {
		for {
			text, err := r.ReadString('\n')
			if err == io.EOF {
				input <- text
				return
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			input <- text
		}
	}()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	var messages []string
	for {
		select {
		case m := <-input:
			messages = append(messages, m)
		case <-ticker.C:
			err := q.Enqueue(messages...)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			messages = messages[:0]
		}
	}
}

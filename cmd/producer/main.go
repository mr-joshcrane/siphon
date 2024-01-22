package main

import (
	"bufio"
	"fmt"
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

	input := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			if text == "" {
				continue
			}
			err := q.Enqueue(text)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}()
	select {
	case data := <-input:
		err := q.Enqueue(data)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case <-time.After(10 * time.Second):
		fmt.Println("Nothing recieved for 10 seconds")
	}
}

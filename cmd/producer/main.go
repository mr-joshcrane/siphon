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

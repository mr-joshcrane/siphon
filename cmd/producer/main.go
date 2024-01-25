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
	for {
		var message string
		select {
		case <-ticker.C:
			for {
				select {
				case m := <-input:
					message += m
					fmt.Println(message)
					continue
				default:
					break
				}
			
			err := q.Enqueue(message)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				break
			}
		}
	}
}

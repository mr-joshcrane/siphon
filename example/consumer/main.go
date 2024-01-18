package main

import (
	"fmt"
	"os"

	"github.com/mr-joshcrane/siphon"
)

func main() {
	q, err := siphon.GetQueue("localhost:8000", "testEvents")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for {
		msg, err := q.Receive()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(msg)
	}
}

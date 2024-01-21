package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mr-joshcrane/siphon"
)

func main() {
	q, err := siphon.NewAWSQueue("testEvents")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		err := q.Enqueue(text)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

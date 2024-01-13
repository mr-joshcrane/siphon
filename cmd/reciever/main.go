package main

import (
	"os"

	"github.com/mr-joshcrane/siphon"
)

func main() {
	err := siphon.Recieve()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

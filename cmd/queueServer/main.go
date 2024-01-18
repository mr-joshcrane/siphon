package main

import (
	"fmt"
	"os"

	"github.com/mr-joshcrane/siphon"
)

func main() {
	err := siphon.ListenAndServe(":8000")
	if err != siphon.ErrServerClosed {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

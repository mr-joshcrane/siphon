package main

import (
	"fmt"
	"os"

	"github.com/mr-joshcrane/siphon"
)

func main() {
	err := siphon.Recieve(os.Stdout)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

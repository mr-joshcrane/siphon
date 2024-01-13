package siphon

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
)

func Main() int {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	stdoutMultiWriter := io.MultiWriter(os.Stdout, conn)
	stderrMultiWriter := io.MultiWriter(os.Stderr, conn)
	defer conn.Close()
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return 1
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
		return 1
	}
	command := fmt.Sprintf(">> %s", strings.Join(os.Args[1:], " "))
	fmt.Fprintln(conn, command)
	go io.Copy(stdoutMultiWriter, stdoutPipe)
	go io.Copy(stderrMultiWriter, stderrPipe)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
		return 1
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return 1
	}
	return 0
}

func Recieve() error {
	fmt.Println("Listening on port 8080")
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			log.Fatal(err)
		}
	}
}

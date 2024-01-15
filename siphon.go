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
	defer conn.Close()
	q := NetworkQueueTarget(conn)
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return 1
	}
	command := fmt.Sprintf(">> %s", strings.Join(os.Args[1:], " "))
	producer := NewProducer(q)
	combinedOutput := io.MultiWriter(os.Stdout, producer)
	fmt.Fprintf(producer, ">> %s\n", command)
	done := make(chan error)
	go func() {
		err := producer.ProduceStream()
		if err != nil {
			log.Fatal(err)
		}
		done <- err

	}()
	go func() {
		io.Copy(combinedOutput, stdoutPipe)
		producer.Done()
	}()
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
		return 1
	}
	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		log.Fatal(err)
		exitCode = 1
	}
	<-done
	return exitCode
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

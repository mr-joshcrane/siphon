package siphon

import (
	"bufio"
	"bytes"
	"encoding/base64"
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

func DoWork() {
	q := NewMemoryQueue(new(bytes.Buffer))
	fmt.Println("Listening on port 8080 for enqueue")
	qListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			conn, err := qListener.Accept()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Accepted connection on 8080")
			handleEnqueue(q, conn)
			if err != nil {
				fmt.Println(err)
			}

		}
	}()
	fmt.Println("Listening on port 8081 for dequeue")
	bListener, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}
	defer bListener.Close()
	go func() {
		for {
			conn, err := bListener.Accept()
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println("Accepted connection on 8081")
			handleDequeue(q, conn)
		}
	}()
	select {}
}

func handleEnqueue(q Queue, conn net.Conn) {
	defer conn.Close()
	fmt.Println("enqueuing")
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		encoded := scanner.Text()
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(decoded))
		err = q.Enqueue(string(decoded))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func handleDequeue(q Queue, conn net.Conn) {
	defer conn.Close()
	fmt.Println("dequeuing")
	for {
		if q.Size() > 0 {
			s, err := q.Dequeue()
			fmt.Println(s)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Fprintln(conn, s)
		}
	}
}

func Recieve(r io.Writer) error {
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = io.Copy(r, conn)
	return err
}

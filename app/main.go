package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	store := NewStore()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, store)
	}
}

func handleConnection(conn net.Conn, store *Store) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		args, err := decodeCommand(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error decoding command:", err)
			}
			return
		}
		if len(args) == 0 {
			continue
		}

		response := dispatch(args, store)
		if _, err := conn.Write(response); err != nil {
			return
		}
	}
}

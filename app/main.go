package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	consumeListner(l)

}

func consumeListner(l net.Listener) {

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			os.Exit(1)
		}
		data := buf[:n]
		decoded, err := decodeCommand(data)
		if err != nil {
			fmt.Println("Error decoding command: ", err.Error())
			continue
		}
		switch decoded[0] {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			if len(decoded) < 2 {
				fmt.Println("ECHO command requires an argument")
			}
			conn.Write(encodeCommand(decoded[1:]))
		}
	}
}

func decodeCommand(data []byte) ([]string, error) {
	if len(data) < 5 || data[0] != '*' {
		return nil, fmt.Errorf("invalid command format")
	}

	parts := strings.Split(string(data[1:]), "\r\n")

	var res []string
	if len(parts) < 2 {
		return res, nil
	}

	for i := 2; i < len(parts); i += 2 {
		res = append(res, parts[i])
	}

	return res, nil
}

func encodeCommand(cmd []string) []byte {
	var res = []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(cmd), cmd))
	return res
}

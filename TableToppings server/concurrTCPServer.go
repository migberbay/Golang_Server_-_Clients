package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	// "encoding/json"
	// "crypto/sha1"
	// "encoding/base64"
)

var connectedClients = 0 // number of currently connected clients to the server.

func handleConnection(c net.Conn) {
	fmt.Print(".")
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			break
		}
		fmt.Println(temp)
		counter := strconv.Itoa(connectedClients) + "\n"
		c.Write([]byte(string(counter)))
	}
	c.Close()
}

func main() {
	config := LoadConfig()
	fmt.Println(config)
	// fmt.Println(config.Users)
	// fmt.Println(config.Worlds)

	PORT := ":" + config.Port
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Server is up and running in port ", config.Port, ", ready for master connection.")
	}

	// defer makes it so the server closes if the main function is exited.
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
		connectedClients++
	}
}

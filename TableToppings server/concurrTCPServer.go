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

type user_conn struct {
	connection net.Conn
	user       User
}

var connections []user_conn // current connections

func handleConnection(c net.Conn) {
	fmt.Print("new connection accepted, current connected clients are:\n")
	for _, c := range connections {
		fmt.Print(c.connection.RemoteAddr())
		fmt.Print("\n")
	}

	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		// message handling.
		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			break
		}
		fmt.Println(temp)
		counter := strconv.Itoa(len(connections)) + "\n"
		c.Write([]byte(string(counter)))
	}
	c.Close()
}

func addressIsConnected(address string) bool {
	for _, e := range connections {
		if address == strings.Split(e.connection.RemoteAddr().String(), ":")[0] {
			return true
		}
	}
	return false
}

func main() {

	//Load configuration.
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

	for { // for with nothing else attached acts like a while(true)
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		ip := strings.Split(c.RemoteAddr().String(), ":")[0]
		// check the remote adress is not any of the currently
		// connected users.
		if addressIsConnected(ip) {
			return
		}

		var conn user_conn
		conn.connection = c

		//check credentials for initial connection

		connections = append(connections, conn)
		go handleConnection(c)
	}
}

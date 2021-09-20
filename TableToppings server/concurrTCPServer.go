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

func handleConnection(c net.Conn, config Config) {
	fmt.Print("new connection accepted, current connected clients are:\n")
	for _, c := range connections {
		fmt.Print(c.connection.RemoteAddr())
		fmt.Print("\n")
	}

	authed := false

	for !authed {
		fmt.Println("awaiting login attempt")
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		temp := strings.TrimSpace(string(netData))
		fmt.Println(temp)
		log_in_attempt := strings.Split(temp, ":")

		if log_in_attempt[0] != "Login" {
			c.Write([]byte(string("401:")))
			return
		}

		usr_pass := strings.Split(log_in_attempt[1], ",")
		usr := usr_pass[0]
		pass := usr_pass[1]

		for _, e := range config.Users {
			if e.Username == usr {
				if e.Password == pass { // peferably add some type of cyper to this.
					c.Write([]byte(string("001:accepted;" + strconv.Itoa(e.ID) + "," + e.Type + "," + e.Username)))
					fmt.Println("accepted login attempt")
					authed = true
				}
			}
		}
		if !authed {
			c.Write([]byte(string("001:rejected")))
			fmt.Println("rejected login attempt")
		}
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
		c.Write([]byte("401:error"))
	}
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

		connections = append(connections, conn)
		go handleConnection(c, config)
	}
}

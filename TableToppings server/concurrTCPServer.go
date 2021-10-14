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

func handleConnection(conn user_conn, config Config) {
	current_connection := conn.connection
	fmt.Print("new connection accepted\n")

	// AUTHENTICATION PROCESS
	authed, err := AuthenticationHandler(conn, config)

	if err != nil || !authed {
		fmt.Println(err)
		current_connection.Write([]byte("401:Fatal error during authentication."))
		return
	}

	// MANAGEMENT.
	for {
		//read new petition from client
		netData, err := bufio.NewReader(current_connection).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		// message handling.
		msg := strings.TrimSpace(string(netData))
		if msg == "STOP" {
			break
		}

		fmt.Println("Message recieved from server it is:" + msg)
		// fmt.Println(temp)

		current_connection.Write([]byte("401:unknown error"))
	}
}

func AuthenticationHandler(conn user_conn, config Config) (bool, error) {
	current_connection := conn.connection

	authed := false
	for !authed {
		fmt.Println("awaiting login attempt for connection:")
		fmt.Println(current_connection.RemoteAddr())

		cont := true

		netData, err := bufio.NewReader(current_connection).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			current_connection.Write([]byte(string("400: Server reading error.\n")))
			cont = false
		}

		temp := strings.TrimSpace(string(netData))
		fmt.Println(temp)
		log_in_attempt := strings.Split(temp, ":")

		if log_in_attempt[0] != "Login" {
			current_connection.Write([]byte(string("400: Invalid login attempt.\n")))
			cont = false
		}

		if cont {
			usr_pass := strings.Split(log_in_attempt[1], ",")
			usr := usr_pass[0]
			pass := usr_pass[1]

			for _, e := range config.Users {
				if e.Username == usr {
					if e.Password == pass { // peferably add some type of cyper to this.

						if addressIsConnected(current_connection.RemoteAddr().String()) {
							current_connection.Write([]byte(string("400: User is already connected, something went wrong.")))
							break
						}

						current_connection.Write([]byte(string("001:accepted;" + strconv.Itoa(e.ID) + "," + e.Type + "," + e.Username)))
						authed = true

						fmt.Print("Authentication accepted, current connections are:")
						for _, c := range connections {
							if c.connection.RemoteAddr() == conn.connection.RemoteAddr() {
								c.user.ID = e.ID
								c.user.Username = e.Username
								c.user.Type = e.Type
							}

							fmt.Printf("%s - %s (%s)", c.connection.RemoteAddr(), c.user.Username, c.user.Type)
							fmt.Print("\n")
						}
					}
					break
				}
			}

			if !authed {
				current_connection.Write([]byte(string("001:rejected")))
				fmt.Println("rejected login attempt")
			}
		}
	}
	return authed, nil
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
	// fmt.Println(config)
	fmt.Println(config.Users)
	fmt.Println(config.Worlds)

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
		go handleConnection(conn, config)
	}
}

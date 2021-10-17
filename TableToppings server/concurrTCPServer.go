package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"gitlab.com/NebulousLabs/go-upnp"
	// "encoding/json"
	// "crypto/sha1"
	// "encoding/base64"
)

type user_conn struct {
	connection net.Conn
	user       User
}

var connections []user_conn // current connections
var external_ip string

func logger_message(conn net.Conn, message string) {
	t := time.Now()
	formatted := fmt.Sprintf("%02d-%02d-%d %02d:%02d:%02d",
		t.Day(), t.Month(), t.Year(),
		t.Hour(), t.Minute(), t.Second())
	fmt.Println("[", conn.RemoteAddr(), formatted, "]: ", message)
}

func handleConnection(conn user_conn, config Config) {
	current_connection := conn.connection

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
		logger_message(current_connection, "awaiting login attempt")
		cont := true

		netData, err := bufio.NewReader(current_connection).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			current_connection.Write([]byte(string("400: Server reading error.\n")))
			cont = false
		}

		temp := strings.TrimSpace(string(netData))

		log_in_attempt := strings.Split(temp, ":")
		msg := "Login attempt => "
		msg = msg + temp
		logger_message(current_connection, msg)

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
						current_connection.Write([]byte(string("001:accepted;" + strconv.Itoa(e.ID) + "," + e.Type + "," + e.Username)))
						authed = true

						logger_message(current_connection, "Authentication accepted, current connections are:")
						for _, c := range connections {
							if c.connection.RemoteAddr() == conn.connection.RemoteAddr() {
								c.user.ID = e.ID
								c.user.Username = e.Username
								c.user.Type = e.Type
							}

							fmt.Printf("%s - %s (%s)\n", c.connection.RemoteAddr(), c.user.Username, c.user.Type)
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
		if address == e.connection.RemoteAddr().String() {
			return true
		}
	}
	return false
}

func forwardConfigPort(config Config) {
	d, err := upnp.Discover()
	if err != nil {
		fmt.Println("Discovering router failed, network failure?")
	}

	external_ip, err = d.ExternalIP()
	if err != nil {
		fmt.Println("Extracting external ID failed, is UPnP allowed by your router?")
	} else {
		fmt.Println("External IP is ", external_ip)
	}

	port_i64, err := strconv.ParseUint(config.Port, 10, 16)
	port_i16 := uint16(port_i64)

	//Example of port forwarding, this probably neesd to be moved to someplace else.
	err = d.Forward(port_i16, "TableToppings Server")
	if err != nil {
		fmt.Println("Error Forwarding port")
	} else {
		fmt.Println("Forwarded port ", port_i16, " ready for external connections.")
	}
}

func main() {
	//Load configuration.
	config := LoadConfig()
	fmt.Println("Users: ", config.Users)
	fmt.Println("Worlds: ", config.Worlds)

	PORT := ":" + config.Port

	listener, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Server is up and running in port ", config.Port, ", ready for master connection.")
	}

	// defer makes it so the server closes if the main function is exited.
	defer listener.Close()

	for { // for with nothing else attached acts like a while(true)
		c, err := listener.Accept()

		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Print("Accepted connection ", c.RemoteAddr(), "\n")
		}

		// check the remote adress is not any of the currently connected sockets. (ip:port)
		if addressIsConnected(c.RemoteAddr().String()) {
			return
		}

		var conn user_conn
		conn.connection = c
		connections = append(connections, conn)
		fmt.Println("Ready for next connection")

		go handleConnection(conn, config)
	}
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"gitlab.com/NebulousLabs/go-upnp"
	// "crypto/sha1"
	// "encoding/base64"
)

type user_conn struct {
	connection net.Conn
	user       User
}

var connections []user_conn // current connections
var external_ip string
var config Config

// Main Function for handling connections, meant to be paralelized for 1 thread per connection.
func handleConnection(conn user_conn) {
	current_connection := conn.connection

	// AUTHENTICATION PROCESS
	authed, err := AuthenticationHandler(conn)

	if err != nil || !authed {
		fmt.Println(err)
		current_connection.Write([]byte("401:Fatal error during authentication."))
		return
	}

	displayAvaliableConnections()

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
		logger_message(current_connection, "Message recieved from server => "+msg)
		stopConnFlag := messageHandler(msg, current_connection)

		if stopConnFlag {
			current_connection.Write([]byte("002")) // Send logout back to client.
			break
		}
	}
}

func messageHandler(msg string, conn net.Conn) bool {
	info := strings.Split(msg, ":")
	main_code := info[0]

	if info[0] == "Logout" { // special logout case
		logout_user_id, _ := strconv.Atoi(info[1])
		for i, c := range connections {
			println("user id: ", c.user.ID, "user type:  ", c.user.Type, "username: ", c.user.Username, "logged user id: ", logout_user_id, logout_user_id == i)
			if c.user.ID == logout_user_id {
				logger_message(c.connection, "Removing from connections...")
				connections = append(connections[:i], connections[i+1:]...)
				return true
			}
		}
		conn.Write([]byte("400: Error loggin out user..."))
		logger_message(conn, "Could not logout requested user...")
		return false
	}

	// codes:
	// 0XX -> connection codes.
	// 1XX -> actions (in game operations).
	// 2XX -> audio.
	// 3XX -> chat codes.
	// 400 -> error code.
	switch main_code[0] {
	case '0':
		ConnectionSubcodeHandler(main_code[1:], info[1], conn)
	case '1':

		break
	case '2':

		break
	default:
		m := "Code not understood, message was: " + msg
		logger_message(conn, m)
	}

	return false
}

// Awaits a login attempt if incorrect, returns false and exits.
func AuthenticationHandler(conn user_conn) (bool, error) {
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
						for i, c := range connections {
							if c.connection.RemoteAddr() == conn.connection.RemoteAddr() {
								var u User
								u.ID = e.ID
								u.Type = e.Type
								u.Username = e.Username

								connections[i].user = u
							}

							fmt.Printf("%s - %s (%s)\n", c.connection.RemoteAddr(), e.Username, e.Type)
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

// ENTRY FUNCTION.
func main() {
	//Load configuration.
	config = LoadConfig()
	fmt.Println("Users: ", config.Users)
	fmt.Println("Worlds: ", config.Worlds)

	PORT := ":" + config.Port

	// Open the port in the config file via UPnP
	forwardConfigPort()

	listener, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Server is up and running in port ", config.Port, ", ready for master connection.")
	}

	// defer Executes the function id we exit the current one (when no more clients exist on the file.)
	defer onClose(listener)

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

		go handleConnection(conn)
	}
}

// SUBCODE HANDLERS
func ConnectionSubcodeHandler(subcode string, info string, conn net.Conn) {
	switch subcode {
	case "03": // Client asking for worlds belonging to user id
		user_id, _ := strconv.Atoi(info)
		toSend := ""
		for _, w := range config.Worlds {
			if w.Owner == user_id {
				json_world, _ := json.Marshal(w)
				toSend += string(json_world)
			}
		}
		conn.Write([]byte("003:" + toSend))

	default:
		m := "Subcode not handled message was: 0" + subcode + ":" + info
		logger_message(conn, m)
	}
}

// AUXILIARY FUNCTIONS

// Logs a message to the console.
func logger_message(conn net.Conn, message string) {
	t := time.Now()
	formatted := fmt.Sprintf("%02d-%02d-%d %02d:%02d:%02d",
		t.Day(), t.Month(), t.Year(),
		t.Hour(), t.Minute(), t.Second())
	fmt.Println("[", conn.RemoteAddr(), formatted, "]: ", message)
}

//Check if the given address (IP:port) is on the connected client list.
func addressIsConnected(address string) bool {
	for _, e := range connections {
		if address == e.connection.RemoteAddr().String() {
			return true
		}
	}
	return false
}

// Forwards the port given by the config file if UPnP is enabled in the router.
func forwardConfigPort() {
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

// Shows all connections in on the server currently
func displayAvaliableConnections() {
	for _, c := range connections {
		fmt.Println("IP: ", c.connection.RemoteAddr(), "user ID:", c.user.ID, ", User type: ", c.user.Type, ", User value: ", c.user.Username)
	}
}

// Runs on defer, meaning the program has no current connections in the listener, currently not working.
func onClose(listener net.Listener) {
	listener.Close()
	println("Server has stopped due to lack of connected clients, relaunching...")
	main()
}

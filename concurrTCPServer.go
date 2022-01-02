package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	upnp "github.com/jcuga/go-upnp"
	// "crypto/sha1"
	// "encoding/base64"
)

// TODO: READ THIS!
// https://johnpili.com/how-to-parse-json-data-without-struct-in-golang/

type user_conn struct {
	connection net.Conn
	user       User
}

var connections []user_conn // current connections
var external_ip string
var config Config
var activeWorld World
var messageCount int = 0

// Main Function for handling connections, meant to be paralelized for 1 thread per connection.
func handleConnection(conn user_conn) {
	current_connection := conn.connection

	// AUTHENTICATION PROCESS
	authed, err := AuthenticationHandler(conn)

	if err != nil || !authed {
		fmt.Println(err)
		// this is done outside the send MessageToClient cuz threading ordering.
		SendMessageToClient(current_connection, []byte("401:Fatal error during authentication."))
		return
	}

	displayAvaliableConnections()

	// MANAGEMENT.
	for {
		//read new petition from client
		reader := bufio.NewReader(current_connection)

		result := ""

		var endCond bool = false

		for !endCond {
			line, isPrefix, err := reader.ReadLine()
			result += string(line)

			if err != nil || !isPrefix {
				fmt.Println(err)
				endCond = true
			}
		}

		// message handling.
		msg := strings.TrimSpace(string(result))
		logger_message(current_connection, "Message recieved from client => "+msg)
		stopConnFlag := messageHandler(msg, conn)

		if stopConnFlag {

			SendMessageToClient(current_connection, []byte("002:Logout processed"))
			break
		}
	}
}

func messageHandler(msg string, connection user_conn) bool {
	info := strings.SplitN(msg, ":", 2)
	main_code := info[0]

	conn := connection.connection

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

		SendMessageToClient(conn, []byte("400: Error loggin out user..."))
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
		ConnectionSubcodeHandler(main_code[1:], info[1], connection)
	case '1':

		break
	case '2':

		break
	case '3':

		break

	case '4':

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

			SendMessageToClient(current_connection, []byte("400: Server reading error."))
			cont = false
		}

		temp := strings.TrimSpace(string(netData))

		log_in_attempt := strings.Split(temp, ":")
		msg := "Login attempt => "
		msg = msg + temp
		logger_message(current_connection, msg)

		if log_in_attempt[0] != "Login" {

			SendMessageToClient(current_connection, []byte("400: Invalid login attempt"))
			cont = false
		}

		if cont {
			usr_pass := strings.Split(log_in_attempt[1], ",")
			usr := usr_pass[0]
			pass := usr_pass[1]

			for _, e := range config.Users {
				if e.Username == usr {
					if e.Password == pass { // peferably add some type of cyper to this.

						SendMessageToClient(current_connection, []byte("001:accepted;"+strconv.Itoa(e.ID)+","+e.Type+","+e.Username))
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

				SendMessageToClient(current_connection, []byte("001:rejected"))
				fmt.Println("rejected login attempt")
			}
		}
	}
	return authed, nil
}

//the client expects a 1024 byte message maximum ,and its not a good idea to send
//bigger packets anyways so we divide possible packets into smaller ones before sending them
func SendMessageToClient(c net.Conn, msg []byte) {
	messageCount++
	res := DivideArrayAndAddIDs(msg)
	res = AddPaddingToMessage(res)

	for _, p := range res {
		fmt.Printf("sending message with length: %v , contents: %v \n", len(p), string(p))
		c.Write(p)
	}
}

// o split byte[] into chunks of < 16384 bytes and add an ID number followed by t if its the last chunk
func DivideArrayAndAddIDs(msg []byte) [][]byte {
	strid := strconv.Itoa(messageCount)
	res := make([][]byte, 0, 16384)
	parts := (len(msg) / 16000) + 1

	for p := 0; p < parts; p++ {
		i := p * 16000
		toadd := strconv.Itoa(p) + ":" + strid
		if p == parts-1 {
			toadd += "t;"
			res = append(res, []byte(toadd+string(msg[i:])))
		} else {
			toadd += ";"
			res = append(res, []byte(toadd+string(msg[i:i+16000])))
		}
	}

	return res
}

func AddPaddingToMessage(toSend [][]byte) [][]byte {
	var err error
	for i, p := range toSend {
		toSend[i], err = PadByteArray(p, 16384)
		errCheck(err)
	}
	return toSend
}

// pkcs7Pad right-pads the given byte slice with 1 to n bytes, where
// n is the block size. The size of the result is x times n, where x
// is at least 1.
func PadByteArray(b []byte, blocksize int) ([]byte, error) {
	// ErrInvalidBlockSize indicates hash blocksize <= 0.
	ErrInvalidBlockSize := errors.New("invalid blocksize")
	// ErrInvalidPKCS7Data indicates bad input
	ErrInvalidData := errors.New("invalid data (empty, null, full or over)")

	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if len(b) == 0 || len(b) >= blocksize {
		return nil, ErrInvalidData
	}

	// padding bytes to add initially
	toPad := blocksize - len(b)
	initial := iterativeDigitsCount(toPad)
	// we aremove from the padding ammount the number of bytes we are appending in front:
	summed := iterativeDigitsCount(toPad - initial - 1)
	// when we add the padding info we change the padding needed if we go to a smaller ammount.
	toPad -= summed + 1
	// padding infostring
	ab := strconv.Itoa(toPad) + ":"

	aux := make([][]byte, 0)
	aux = append(aux, []byte(ab))
	aux = append(aux, b)
	b = bytes.Join(aux, make([]byte, 0))

	// toPad = toPad - len(ab)

	pb := make([]byte, blocksize)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(' ')}, toPad))
	return pb, nil
}

//Count the number of digits of an element
func iterativeDigitsCount(number int) int {
	count := 0
	for number != 0 {
		number /= 10
		count += 1
	}
	return count
}

// ENTRY FUNCTION.
func main() {
	//Load configuration.
	config = LoadConfig()
	fmt.Println("Users: ", config.Users)
	fmt.Println("Worlds: ", config.Worlds)

	PORT := ":" + config.Port

	// Open the port in the config file via UPnP
	err := forwardConfigPort()
	if err != nil {
		fmt.Println("Couldn't open a port in your router, make sure UPnP protocol is enabled. Proceeding in local only mode.")
	}

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

// -----------SUBCODE HANDLERS--------------
// CONNECTION SUBCODES
func ConnectionSubcodeHandler(subcode string, info string, connection user_conn) {
	conn := connection.connection
	switch subcode {
	case "03": // Client asking for worlds belonging to user id
		user_id, _ := strconv.Atoi(info)
		msg := "Sending world info for user with id: " + info
		logger_message(conn, msg)
		toSend := "{\"worlds\":["
		for _, w := range config.Worlds {
			if w.Owner == user_id {
				json_world, _ := json.Marshal(w)
				toSend += string(json_world)
				toSend += ","
			}
		}
		toSend += "]}"

		SendMessageToClient(conn, []byte("003:"+toSend))

	case "04": //Client asking for world to be loaded & updated
		w_id, _ := strconv.Atoi(info)
		for _, w := range config.Worlds {
			if w.ID == w_id {
				activeWorld = w
				msg := activeWorld.Name + " has been activated."
				logger_message(conn, msg)
				break
			}
		}

		json_world, _ := json.Marshal(activeWorld)
		toSend := string(json_world)

		msg := "004:" + toSend

		SendMessageToClient(conn, []byte(msg))

	case "05": //Client sending file information for file syncing.
		CompareAndUpdateClientFiles(connection, info, false)

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
func forwardConfigPort() error {

	d, err := upnp.Discover()
	if err != nil {
		fmt.Println("Failed discovery; ", err)
		return err
	}

	// discover external IP
	external_ip, err = d.ExternalIP()
	if err != nil {
		fmt.Println("Error fetching external IP address; ", err)
		return err
	}
	fmt.Println("Your external IP is:", external_ip)

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		fmt.Printf("Error converting port from config file: %v\n", err)
		return err
	}

	fmt.Println("Opening port forward for TCP port ", port)
	errT := d.Forward(uint16(port), "dndserver", "TCP")
	// errU := d.Forward(uint16(port), "dndserver", "UDP")
	if errT != nil { //|| errU != nil
		fmt.Println("Error opening port forward: ", errT) //errU
		return err
	}

	return nil
}

// Shows all connections in on the server currently
func displayAvaliableConnections() {
	for _, c := range connections {
		fmt.Println("IP: ", c.connection.RemoteAddr(), "user ID:", c.user.ID, ", User type: ", c.user.Type, ", User value: ", c.user.Username)
	}
}

//TODO: fix this.
// Runs on defer, meaning the program has no current connections in the listener, currently not working.
func onClose(listener net.Listener) {
	listener.Close()
	println("Server has stopped due to lack of connected clients, relaunching...")
	main()
}

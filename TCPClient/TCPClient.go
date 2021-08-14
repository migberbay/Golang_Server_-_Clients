package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// from https://www.linode.com/docs/guides/developing-udp-and-tcp-clients-and-servers-in-go/

func main() {
	args := os.Args // list containing the arguments of the package
	if len(args) == 1 {
		fmt.Println("Please provide host:port")
		return
	}

	CONNECT := args[1]
	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}

	for { // this is an endless while loop in golang.
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(c, text+"\n")

		msg, _ := bufio.NewReader(c).ReadString('\n')
		fmt.Print("->: " + msg)
		if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}

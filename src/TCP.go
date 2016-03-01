package main

import (
	"net"
	"bufio"
	"os"
	"fmt"
	"time"
	"log"
)

func TCP_receive(conn net.Conn) string {
	//Wait for message ending in '\0'
	msg, _ := bufio.NewReader(conn).ReadString(byte('\x00'))
	return msg
}

func TCP_send(conn net.Conn, msg string) {
	conn.Write([]byte(msg + string('\x00')))
}

func main() {
	serverAddr := "129.241.187.23:33546"

	fmt.Println("Launching TCP server...")
	input := bufio.NewReader(os.Stdin)
	
	//Get the servers TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		log.Fatal("ResolveTCPAddr failed: ", err.Error())
	}
	
	//Connect to the TCP server
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal("DialTCP failed: ", err.Error())
	}

	//Receive welcome message
	msg_rcpt := TCP_receive(conn)
	fmt.Println(msg_rcpt)

	//Close connection when the script ends
	defer conn.Close()
	
	for {
		//Send message to the server
		fmt.Print("Enter message to send to server: ")
		msg_send, _ := input.ReadString('\n')
		TCP_send(conn, msg_send)
		
		msg_rcpt := TCP_receive(conn)
		
		//Print received message to console
		fmt.Print("Message received: ", string(msg_rcpt))
		
		//Prevent spamming the network
		time.Sleep(1)
	}
}


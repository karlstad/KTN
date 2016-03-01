package main

import(
	"config"
	"fmt"
	"strings"
	"encoding/json"
	"net"
	"log"
)

func TCP_receive(conn net.Conn/*, ch_receive chan<- config.ClientMessage*/) {
	for{
		//Wait for message ending in '\0'
		//received, _ := bufio.NewReader(conn).ReadString(byte('\x00'))
		received_byte := make([]byte, 1024)
		
		conn.Read(received_byte)
		
		var msg config.ServerMessage
		err := json.Unmarshal(received_byte, &msg)
		if err != nil {
			log.Printf("TCP_receive: json error:", err)
		}
		//fmt.Printf("%s, %s, %s, %s", msg.timestamp, msg.sender, msg.response, msg.content)
	}
}

func TCP_send(conn net.Conn, ch_send <-chan config.ClientMessage) {
	for{
		msg := ch_send
		json_msg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("TCP_send: json error:", err)
		}
		conn.Write([]byte(json_msg))
	}
}

//Assuming ip contains the port and address
func TCP_init(ip string) *net.TCPConn {
	serverAddr := ip
	
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
	
	return conn
}

func main(){
	var logged_in bool
	var ch_send = make(chan config.ClientMessage)
	//var ch_receive = make(chan config.ClientMessage)

	conn := TCP_init("pella")
	defer conn.Close()
	
	go TCP_receive(conn)
	go TCP_send(conn, ch_send)

	var input string
	var splitted []string
	fmt.Print("Enter command: \n login username \n logout \n names \n help \n All other inputs will be treated as messages")
	fmt.Printf("Please log in to be able to chat")
	for{
		fmt.Scanln(&input)
		
		splitted = strings.Split(input, " ")
		fmt.Printf("%q\n", splitted)
		msg := config.ClientMessage{Request: splitted[0], Content: splitted[1]}
		if msg.Request == "login"{
			logged_in = true
			ch_send <- msg
		}
		if logged_in == true{
			switch msg.Request{
				case "logout":
					logged_in = false
					ch_send <- msg
				case "names":
					ch_send <- msg
				case "help":
					ch_send <- msg
				default:
					msg.Request = "msg"
					msg.Content = input
					ch_send <- msg
			}
		}
	}
}

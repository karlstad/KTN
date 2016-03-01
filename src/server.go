package main

import (
	"log"
	"encoding/json"	
	"net"
	"time"
	"strings"
)

type Client struct {
	conn *net.TCPConn
	username string
}

type ServerMessage struct {
	timestamp string
	sender string
	response string
	content string
	conn *net.TCPConn `json="-"`
}

type ClientMessage struct{
	request string
	content string
	conn *net.TCPConn `json="-"` //Denne vil ikke bli med i json encoding --> Client sender ikke denne, den opprettes server-side
}

const TCPPORT = ""

var Connections = make([]Client,0)
var ReceivedChan = make(chan ClientMessage, 5)
var SendChan = make(chan ServerMessage)
var BroadcastChan = make(chan ServerMessage)
var History = make([]string, 0)

func TCPListen() {
	TcpPort, _ := net.ResolveTCPAddr("tcp", "78.91.15.53:30000")
	TcpListener, _ := net.ListenTCP("tcp", TcpPort)
	for {
		connection,_ := TcpListener.AcceptTCP()
		Connections = append(Connections, Client{conn: connection, username: ""})
		go TCPReceive(connection)
	}
}

//Unødvendig?
/*func IsLoggedIn(username struct) bool {
	for client := range Connections {
		if client.username == username {
			return true
		}
	}
	return false
}*/

func TCPSend(chSend <-chan ServerMessage){
	for{
		msg := <- chSend
		json_msg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("TCP_send: json error:", err)
		}
		msg.conn.Write([]byte(json_msg))
	}
}

func TCPBroadcast(chMsg <-chan ServerMessage){
	for{
		msg := <- chMsg
		json_msg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("TCP_send: json error:", err)
		}
		for clients := range Connections{
			Connections[clients].conn.Write([]byte(json_msg))
		}
	}
}

func TCPReceive(conn *net.TCPConn){
	received := make([]byte, 1024)
	for{
		conn.Read(received)
		
		var msg ClientMessage
		err := json.Unmarshal(received, &msg)
		if err != nil {
			log.Printf("TCP_receive: json error:", err)
		}
		log.Printf("%q\n", msg)
		ReceivedChan <- msg
	}
}



func main() {
	go TCPListen()
	go TCPSend(SendChan)
	go TCPBroadcast(BroadcastChan)
	
	for {
		msg := <- ReceivedChan
		switch msg.request {
			case "login":
				for client := range Connections {
					if msg.conn == Connections[client].conn {
						if msg.content != Connections[client].username {
							Connections[client].username = msg.content
							SendChan <- ServerMessage{timestamp: time.Now().Format(time.RFC850), sender: "server", response: "info", content: "Login successful"}
							SendChan <- ServerMessage{timestamp: time.Now().Format(time.RFC850), sender: "server", response: "history", content: strings.Join(History, "\n")}
						} else {
							SendChan <- ServerMessage{timestamp: time.Now().Format(time.RFC850), sender: "server", response: "error", content: "User already logged in!"}
						}
						break
					}
				}		
			case "logout":
				for client := range Connections {
					if Connections[client].conn == msg.conn {
						Connections[client].username = ""
						SendChan <- ServerMessage{timestamp: time.Now().Format(time.RFC850), sender: "server", response: "info", content: "Logout successful"}
					}
				}
			case "msg":
				History = append(History, msg.content)
				BroadcastChan <- ServerMessage{timestamp: time.Now().Format(time.RFC850), sender: "server", response : "message", content: msg.content}
			case "names":
				usernames := make([]string, 0)
				for client := range Connections{
					usernames = append(usernames, Connections[client].username)
				}
					
				SendChan <- ServerMessage{timestamp: time.Now().Format(time.RFC850), sender: "server", response : "message", content: strings.Join(usernames, "\n")}
				
			//case "help":
		

		}
	}
}















package main

import {
	"log"
	"encoding/json"
	
}

type Client struct {
	conn *net.TCPConn
	username string
}

type ServerMessage struct {
	timestamp string
	sender string
	response string
	content string
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

func TCPListen() {
	TcpPort, _ := net.ResolveTCPAddr(SERVERADDR, SERVERPORT)
	TcpListener, _ := net.ListenTCP("tcp", tcp_port)
	for {
		conn,_ := TcpListener.Accept()
		Connections = append(Connections, Client{Conn: conn, Authenticated: false})
	}
}

//Unødvendig?
func IsLoggedIn(username struct) bool {
	for client := range Connections {
		if client.username == username {
			return true
		}
	}
	return false
}

func main() {
	for {
		msg := <- ReceivedChan
		switch msg.request {
			case "login":
				for client := range Connections {
					if msg.conn == client.conn {
						if msg.username != client.username {
							client.username = msg.content
						} else {
							SendChan <- ServerMessage{timestamp: time.Now().Local(), sender: "server", response: "error", content: "User already logged in!"}
						}
						break
					}
				}		
			case "logout":
				for client := range Connections {
					if client.conn == msg.conn {
						client.username = ""
					}
				}
			case "msg":
				
			case "names":
				
			case "help":
		}
	}
}















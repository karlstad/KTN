package main

import (
	"log"
	"encoding/json"	
	"net"
	"time"
	"strings"
	"bufio"
	"bytes"
	"regexp"
)

type Client struct {
	conn *net.TCPConn
	username string
}

type ServerMessage struct {
	Timestamp string `json:"timestamp"`
	Sender string `json:"sender"`
	Response string	`json:"response"`
	Content string `json:"content"`
	conn *net.TCPConn `json:"-"`
}

type ClientMessage struct{
	Request string `json:"request"`
	Content string `json:"content"`
	conn *net.TCPConn `json:"-"`
}

const TCPPORT = ""

var Connections = make([]*Client,0)
var ReceivedChan = make(chan ClientMessage, 5)
var SendChan = make(chan ServerMessage)
var BroadcastChan = make(chan ServerMessage)
var History = make([]string, 0)

func TCPListen() {
	TcpPort, _ := net.ResolveTCPAddr("tcp", ":30000")
	TcpListener, _ := net.ListenTCP("tcp", TcpPort)
	for {
		connection,_ := TcpListener.AcceptTCP()
		log.Printf("Connection made to %s!\n", connection.RemoteAddr().String())
		go TCPReceive(connection)
		Connections = append(Connections, &Client{conn: connection, username: ""})
		time.Sleep(100*time.Millisecond)
	}
}

func TCPSend(chSend <-chan ServerMessage){
	for{
		msg := <- chSend
		json_msg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("TCP_send: json error:", err)
		}
		msg.conn.Write(append([]byte(json_msg),byte('\x00')))
		time.Sleep(100*time.Millisecond)
	}
}

func TCPBroadcast(chMsg <-chan ServerMessage){
	for{
		msg := <- chMsg
		json_msg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("TCP_broadcast: json error:", err)
		}
		for _, client := range Connections{
			if client.conn != msg.conn {
				client.conn.Write(append([]byte(json_msg),byte('\x00')))
			}
		}
		time.Sleep(100*time.Millisecond)
	}
}

func TCPReceive(conn *net.TCPConn){
	for{
		rec, _ := bufio.NewReader(conn).ReadString(byte('\x00'))
		rec = strings.Trim(rec, "\x00")
		received := []byte(rec)
		var msg ClientMessage
		err := json.Unmarshal(received, &msg)
		if err != nil {
			log.Printf("TCP_receive: json error! Shutting down thread. (IP: %s", conn.RemoteAddr().String())
			KillConnection(conn)
			return
		}
		msg.conn = conn
		ReceivedChan <- msg
		time.Sleep(100*time.Millisecond)
	}
}

func IsLoggedIn(conn *net.TCPConn) bool {
	for _, client := range Connections {
		if client.conn == conn && client.username != "" {
			return true
		}
	}
	return false
}

func KillConnection(conn *net.TCPConn) {
	for key, client := range Connections {
		if client.conn == conn {
			Connections = append(Connections[:key], Connections[key+1:]...)
			conn.Close()
		}
	}
}

func GetUsername(conn *net.TCPConn) string {
	for _,client := range Connections {
		if client.conn == conn {
			return client.username
		}
	}
	return ""
}

func AddHistoryEntry(msg *ClientMessage) {
	var buf bytes.Buffer
	buf.WriteString("<")
	buf.WriteString(time.Now().Format(time.RFC850))
	buf.WriteString("> ")
	buf.WriteString(GetUsername(msg.conn))
	buf.WriteString(": ")
	buf.WriteString(msg.Content)
	History = append(History, buf.String())
}

func main() {
	helpTextNotLoggedIn := "\nCommands:\nlogin <username>\nhelp"
	helpTextLoggedIn := "\nCommands:\nlogin <username>\nlogout\nnames\nhelp"
	nameCheck := regexp.MustCompile("^[a-zA-Z0-9]*$")

	go TCPListen()
	go TCPSend(SendChan)
	go TCPBroadcast(BroadcastChan)
	
	for {
		msg := <- ReceivedChan
		if IsLoggedIn(msg.conn) {
			switch msg.Request {
				case "login":
					SendChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: "server", Response: "error", Content: "You are already logged in", conn: msg.conn}
				case "logout":
					for _, client := range Connections {
						if client.conn == msg.conn {
							client.username = ""
							SendChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: "server", Response: "info", Content: "Logout successful", conn: msg.conn}
						}
					}
				case "msg":
					AddHistoryEntry(&msg)
					BroadcastChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: GetUsername(msg.conn), Response : "message", Content: msg.Content, conn: msg.conn}
				case "names":
					usernames := make([]string, 0)
					for _, client := range Connections{
						usernames = append(usernames, client.username)
					}			
					SendChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: "server", Response : "message", Content: strings.Join(usernames, "\n"), conn: msg.conn}
				case "help":
					SendChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: "server", Response: "message", Content: helpTextLoggedIn, conn: msg.conn}
				default:
					SendChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: "server", Response: "error", Content: "Invalid request", conn: msg.conn}
			}
		} else {
			switch msg.Request {
				case "login":
					for _, client := range Connections {
						if msg.conn == client.conn {
							if msg.Content != client.username {
								if nameCheck.MatchString(msg.Content){
									client.username = msg.Content
									SendChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: "server", Response: "info", Content: "Login successful", conn: msg.conn}
									SendChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: "server", Response: "history", Content: strings.Join(History, "\n"), conn: msg.conn}
								} else {
									SendChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: "server", Response: "error", Content: "Only alphanumerical usernames are acceptable.", conn: msg.conn}
								}
							} else {
								SendChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: "server", Response: "error", Content: "User already logged in!", conn: msg.conn}
							}
							break
						}
					}
				case "help":
					SendChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: "server", Response: "message", Content: helpTextNotLoggedIn, conn: msg.conn}
				default:
					SendChan <- ServerMessage{Timestamp: time.Now().Format(time.RFC850), Sender: "server", Response: "error", Content: "You are not logged in! Please use <help> for a list of commands.", conn: msg.conn}
			}	
		}
		time.Sleep(100*time.Millisecond)
	}
}















package main

import(
	"fmt"
	"os"
	"strings"
	"encoding/json"
	"net"
	"log"
	"time"
	"bufio"
	//"reflect"
)

type HistoryMessage struct {
	Timestamp string `json:"timestamp"`
	Sender string `json:"sender"`
	Response string	`json:"response"`
	Content []string `json:"content"`
	conn *net.TCPConn `json:"-"`
}

type ServerMessage struct {
	Timestamp string `json:"timestamp"`
	Sender string `json:"sender"`
	Response string	`json:"response"`
	Content string `json:"content"`
}

type ClientMessage struct{
	Request string `json:"request"`
	Content string `json:"content"`
}

var HistoryChan = make(chan HistoryMessage)

func TCP_receive(conn net.Conn, ch_receive chan<- ServerMessage) {
	for{
		/*rec, _ := bufio.NewReader(conn).ReadString(byte('}'))
		received := []byte(rec)
		log.Printf("Received %s\n", rec)
		
		var msg ServerMessage
		err := json.Unmarshal(received, &msg)
		if err != nil {
			log.Fatal("TCP_receive: json error:", err)
		}*/
		d := json.NewDecoder(conn)
		log.Printf("Received: %s", d)
		var msg ServerMessage
		//var hist HistoryMessage
		d.Decode(&msg)
		//d.Decode(&hist)
		//log.Printf("Received %s\n", reflect.TypeOf(msg.Content))
		//log.Printf("Received %s\n", msg.Content)
		//if msg.Content != "" {
		//	ch_receive <- msg
		//}
		//if hist.Content != nil {
		//	HistoryChan <- hist
		//}
		time.Sleep(100*time.Millisecond)
	}
}

func TCP_send(conn net.Conn, ch_send <-chan ClientMessage) {
	for{
		msg := <- ch_send
		json_msg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("TCP_send: json error:", err)
		}
		conn.Write([]byte(json_msg))
		time.Sleep(100*time.Millisecond)
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

//{"content": ["{\"content\": \"magnus logged in\", \"timestamp\": \"02-03-2016 14:18:34\", \"sender\": \"system\", \"response\": \"info\"}", 
func Print(ch_receive <-chan ServerMessage) {
	for {
		select {
			case msg := <- ch_receive:
				switch msg.Response {
					case "error":
						fmt.Printf("<%s> ERROR: %s\n", msg.Timestamp, msg.Content)
					case "info":
						fmt.Printf("INFO: %s\n", msg.Content)
					case "message":
						fmt.Printf("<%s> %s said: %s\n", msg.Timestamp, msg.Sender, msg.Content)
					default:
						log.Fatal("Invalid response from server!")
					}
			case msg := <- HistoryChan:
				//var msg ServerMessage
				fmt.Printf("HISTORY: %s", msg.Content)
				/*for _, elem := range msg.Content {
					err := json.Unmarshal(elem, &msg)
					if err != nil {
						log.Fatal("History could not be decoded: ", err)
					}
				}*/		
		}
		time.Sleep(100*time.Millisecond)
	}
}

func main(){
	ch_send := make(chan ClientMessage)
	ch_receive := make(chan ServerMessage, 5)

	conn := TCP_init("10.20.70.103:30000")
	defer conn.Close()
	
	go TCP_receive(conn, ch_receive)
	go TCP_send(conn, ch_send)
	go Print(ch_receive)

	var splitted []string
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		splitted = strings.Fields(text)
		var msg ClientMessage
		req := splitted[0]
		if req == "login" {
			msg = ClientMessage{Request: splitted[0], Content: splitted[1]}
			ch_send <- msg
		} else if req == "logout" || req == "names" || req == "help" {
			msg = ClientMessage{Request: splitted[0]}
			ch_send <- msg
		} else {
			msg.Request = "msg"
			msg.Content = text
			ch_send <- msg
		}
		time.Sleep(100*time.Millisecond)
	}
}

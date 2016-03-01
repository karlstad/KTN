package config

type ClientMessage struct {
	Request string
	Content string
}

type ServerMessage struct{
	Timestamp string
	Sender string
	Response string
	Content string
}

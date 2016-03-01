package config

type ClientMessage struct {
	request string
	content string
}

type ServerMessage struct{
	timestamp string
	sender string
	response string
	content string
}

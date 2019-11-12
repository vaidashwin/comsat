package model

var StreamerMap = make(map[int]Streamer)

type Streamer interface {
	GetLogin() string
	GetAlias() string
	GetLink() string
	GetID() int
}

type StreamerMessage struct {
	Name string
}
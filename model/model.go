package model

import "net/url"

var StreamerMap = make(map[int]Streamer)

type Streamer interface {
	GetName() string
	GetLink() url.URL
	GetID() int
}

type StreamerMessage struct {
	Name string
}

func getStreamerMessage(streamer *Streamer) string {

}

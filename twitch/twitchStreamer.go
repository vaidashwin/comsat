package twitch

import "net/url"

type TwitchStreamer struct {
	name string
	link url.URL
	id int
}

func (streamer TwitchStreamer) GetName() string {
	return streamer.name
}

func (streamer TwitchStreamer) GetLink() url.URL {
	return streamer.link
}

func (streamer TwitchStreamer) GetID() int {
	return streamer.id
}
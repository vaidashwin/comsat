package storage

import (
	"encoding/json"
	"errors"
	"github.com/vaidashwin/comsat/configuration"
	"log"
	"net/http"
	"strconv"
)

type TwitchStreamer struct {
	login string
	alias string
	link string
	id int
}

func NewStreamer(login string, alias string, link string, id int) TwitchStreamer {
	return TwitchStreamer{
		login: login,
		alias: alias,
		link: link,
		id:   id,
	}
}

func (streamer TwitchStreamer) GetLogin() string {
	return streamer.login
}

func (streamer TwitchStreamer) GetAlias() string {
	return streamer.alias
}

func (streamer TwitchStreamer) GetLink() string {
	return streamer.link
}

func (streamer TwitchStreamer) GetID() int {
	return streamer.id
}

var client = &http.Client { }

func GetTwitchStreamers() []TwitchStreamer {
	streams, err := GetStreams("twitch")
	if err != nil {
		return nil
	}
	result := make([]TwitchStreamer, 0)
	for _, stream := range streams {
		switch twitchStreamer := stream.(type) {
		case TwitchStreamer:
			result = append(result, twitchStreamer)
		default:
			continue
		}
	}
	return result
}

func AddTwitchStreamers(streamNames map[string]string) error {
	if len(streamNames) == 0 {
		return errors.New("No streamers added.")
	}
	queryString := ""
	for streamName, _ := range streamNames {
		queryString += streamName + "&"
	}

	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/users?login=" + queryString, nil)
	if err != nil {
		log.Fatal("Failed to get user data.", err)
	}
	req.Header.Add("Client-ID", configuration.Get().TwitchClientID)
	resp, err := client.Do(req)

	if err != nil || resp.Status != "200 OK" {
		log.Fatal("Failed to get user data.", err)
	}
	responseBytes := make([]byte, resp.ContentLength)

	defer resp.Body.Close()
	_, err = resp.Body.Read(responseBytes)
	if err != nil {
		log.Fatal("Failed to get user data.", err)
	}

	log.Println("Got user data:", string(responseBytes))
	type Streamer struct {
		ID string `json:"id"`
		Login string `json:"login"`
	}
	type userList struct {
		Data []Streamer `json:"data"`
	}
	response := userList{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		log.Fatal("Failed to parse user data.", err)
	}

	for _, datum := range response.Data {
		id, err := strconv.Atoi(datum.ID)
		if err != nil {
			log.Println("Error adding stream:", err)
			continue
		}
		alias := streamNames[datum.Login]
		err = AddStream(id, "twitch", datum.Login, alias, "https://twitch.tv/" + datum.Login)
		if err != nil {
			log.Println("Error adding stream:", err)
			continue
		}
	}
	return nil
}
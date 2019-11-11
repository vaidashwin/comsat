package twitch

import (
	"encoding/json"
	"github.com/vaidashwin/comsat/configuration"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var client = &http.Client { }

func InitTwitch(callback func([]StreamData)) {
	log.Println("Initializing Twitch listener.")
	// get the initial streamer list

	// init http server
	streamers := getStreamers()
	pollStatus(streamers, callback)
}

func getStreamers() []Streamer {
	queryString := ""
	for _, streamName := range configuration.Get().Streams {
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
	type userList struct {
		Data []Streamer `json:"data"`
	}
	response := userList{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		log.Fatal("Failed to parse user data.", err)
	}

	result := make([]Streamer, len(configuration.Get().Streams))
	for _, user := range response.Data {
		result = append(result, user)
	}

	return result
}

func pollStatus(streamers []Streamer, callback func([]StreamData)) {
	log.Println("Scanning for streams for IDs:", streamers)
	queryString := ""
	if len(streamers) > 0 {
		for _, streamer := range streamers {
			if streamer.getID() > 0 {
				queryString += "user_id=" + strconv.Itoa(streamer.getID()) + "&"
			}
		}

		log.Println("query string", queryString)

		req, err := http.NewRequest("GET",
			"https://api.twitch.tv/helix/streams/?" + queryString, nil)
		if err != nil {
			log.Fatal("Failed to get stream data.", err)
		}
		req.Header.Add("Client-ID", configuration.Get().TwitchClientID)
		resp, err := client.Do(req)
		if err != nil || resp.Status != "200 OK" {
			log.Fatal("Failed to get user data.", err, resp)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Failed to get stream data.", err)
		}

		type payload struct {
			Data []StreamData `json:"data"`
		}

		respData := &payload{}
		err = json.Unmarshal(body, respData)
		if err != nil {
			log.Fatal("Failed to get stream data.", err)
		}

		streamerMap := make(map[string]string)
		for _, streamer := range streamers {
			streamerMap[streamer.ID] = streamer.Login
		}

		result := respData.Data
		for _, streamer := range result {
			streamer.URL = "https://twitch.tv/" + streamerMap[streamer.UserID]
		}

		log.Println("Scan completed successfully:", result)
		callback(result)
		time.Sleep(5 * time.Minute)
	}
	go pollStatus(streamers, callback)
}

type Streamer struct {
	ID string `json:"id"`
	Login string `json:"login"`
}

func (streamer *Streamer) getID() int {
	id, err := strconv.Atoi(streamer.ID)
	if err != nil {
		return -1
	} else {
		return id
	}
}

type StreamData struct {
	URL string
	UserName string `json:"user_name"`
	UserID string `json:"user_id"`
	Type string `json:"type"`
	Title string `json:"title"`
	Viewers int `json:"viewer_count"`
}
package twitch

import (
	"encoding/json"
	"github.com/vaidashwin/comsat/configuration"
	"github.com/vaidashwin/comsat/storage"
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
	streamers := storage.GetTwitchStreamers()
	pollStatus(streamers, callback)
}

func pollStatus(streamers []storage.TwitchStreamer, callback func([]StreamData)) {
	log.Println("Scanning for streams for IDs:", streamers)
	queryString := ""
	if len(streamers) > 0 {
		for _, streamer := range streamers {
			if streamer.GetID() > 0 {
				queryString += "user_id=" + strconv.Itoa(streamer.GetID()) + "&"
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

		result := respData.Data
		for idx, stream := range result {
			userID, err := strconv.Atoi(stream.UserID)
			if err != nil {
				log.Fatal("Problem decoding response from Twitch.", err)
			}
			// TODO refactor
			var thisStreamer *storage.TwitchStreamer
			for _, streamer := range streamers {
				if streamer.GetID() == userID {
					thisStreamer = &streamer
					break
				}
			}
			result[idx] = StreamData{
				Streamer: thisStreamer,
				UserName: stream.UserName,
				UserID:   stream.UserID,
				Type:     stream.Type,
				Title:    stream.Title,
				Viewers:  stream.Viewers,
			}
		}

		log.Println("Scan completed successfully:", result)
		callback(result)
		time.Sleep(5 * time.Minute)
	} else {
		time.Sleep(30 * time.Second)
	}
	go pollStatus(storage.GetTwitchStreamers(), callback)
}

type StreamData struct {
	Streamer *storage.TwitchStreamer
	UserName string `json:"user_name"`
	UserID string `json:"user_id"`
	Type string `json:"type"`
	Title string `json:"title"`
	Viewers int `json:"viewer_count"`
}
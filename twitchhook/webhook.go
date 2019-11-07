package twitchhook

import (
	"bytes"
	"encoding/json"
	"github.com/vaidashwin/comsat/configuration"
	"log"
	"net/http"
	"strconv"
)

func InitTwitchhooks(callback func([]string)) {
	// init http server
	InitTwitchhookServer(callback)
	for _, streamID := range getStreamIDs() {
		listenToStream(streamID)
	}
}

func getStreamIDs() []int {
	queryString := ""
	for _, streamName := range configuration.Get().Streams {
		queryString += "login=" + streamName + "&"
	}
	client := &http.Client { }

	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/users?" + queryString, nil)
	if err != nil {
		log.Fatal("Failed to get user data.", err)
	}
	req.Header.Add("Client-ID", configuration.Get().TwitchClientID)
	resp, err := client.Do(req)

	if err != nil || resp.Status != "200 OK" {
		log.Fatal("Failed to get user data.")
	}
	responseBytes := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(responseBytes)
	if err != nil {
		log.Fatal("Failed to get user data.")
	}

	type userList struct {
		users []map[string]string
	}
	response := &userList{}
	err = json.Unmarshal(responseBytes, response)
	if err != nil {
		log.Fatal("Failed to get user data.")
	}

	result := make([]int, len(configuration.Get().Streams))
	for _, user := range response.users {
		id, err := strconv.Atoi(user["id"])
		if err != nil {
			log.Fatal("Failed to get user data.")
		}
		result = append(result, id)
	}

	return result
}

func listenToStream(streamID int) {
	bodyBytes, err := json.Marshal(attachToWebhook{
		Callback:     configuration.Get().TwitchCallbackURL,
		Mode:         "subscribe",
		Topic:        "https://api.twitch.tv/helix/streams?user_id=" + string(streamID),
		LeaseSeconds: 864000,
		Secret:       configuration.Get().TwitchSecret,
	})
	if err != nil {
		log.Println("Failed to marshall subscribe message.", err)
		return
	}
	reader := bytes.NewReader(bodyBytes)
	resp, err := http.Post("https://api.twitch.tv/helix/webhooks/hub",
		"application/json",
		reader)
	if err != nil {
		log.Println("Failed to subscribe to stream ", streamID, err)
		return
	}
	if resp.Status != "200 OK" {
		log.Println("Failed to subscribe to stream ", streamID, resp)
	} else {
		log.Println("Subscribed to user ", streamID)
		// TODO: schedule resubscribe
	}
}

type attachToWebhook struct {
	Callback     string `json:"hub.callback"`
	Mode         string `json:"hub.mode"`
	Topic        string `json:"hub.topic"`
	LeaseSeconds int    `json:"hub.lease_seconds"`
	Secret       string `json:"hub.secret"`
}

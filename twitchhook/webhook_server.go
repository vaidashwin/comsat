package twitchhook

import (
	"encoding/json"
	"log"
	"net/http"
)

func InitTwitchhookServer(callback func([]string)) {
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes := make([]byte, r.ContentLength)
		_, err := r.Body.Read(bodyBytes)
		if err != nil {
			log.Println("Failed to get bytes from twitch callback.")
		} else {
			body := twitchCallback{}
			err := json.Unmarshal(bodyBytes, &body)
			if err != nil {
				log.Println("Failed to unmarshal twitch callback.")
			}
			result := make([]string, len(body.Data))
			for idx, datum := range body.Data {
				result[idx] = datum["username"]
			}
			callback(result)
		}
		w.Write([]byte("OK"))
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error setting up twitch webhooks:", err)
	}
}

type twitchCallback struct {
	Data []map[string]string `json:"data"`
}
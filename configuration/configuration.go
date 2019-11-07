package configuration

import (
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
	ServerToChannel   map[string]string
	Streams           []string
	StorageFile       string
	TwitchClientID    string
	TwitchCallbackURL string
	TwitchSecret      string
}

var config *Configuration = &Configuration{}

func InitConfig(configfile string) {
	file, err := os.Open(configfile)
	if err != nil {
		log.Fatal("Unable to open Configuration file!", err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		log.Fatal("Unable to open Configuration file!", err)
	}
}

func Get() *Configuration {
	return config
}

package configuration

import (
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
	ServerToChannel   map[string]string
	Streams           []string
	Logfile			  string
	TwitchClientID    string
	ComsatSecret	  string
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
	result, err := json.Marshal(config)
	log.Println(string(result))
}

func Get() *Configuration {
	return config
}

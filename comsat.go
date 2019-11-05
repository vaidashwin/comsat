package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os/signal"
	"syscall"

	// "gopkg.in/go-playground/webhooks.v3"
	"log"
	"os"
)

var logfile *string
var token *string
var channel *string

func init() {
	// cli params
	token = flag.String("token", "", "Bot token from Discord API.")
	logfile = flag.String("logfile", "", "Logfile.")
	channel = flag.String("channel", "general", "TEMP channel to post in.")
	flag.Parse()
}

func main() {

	if *logfile != "" {
		// init logfile
		f, err := os.OpenFile(*logfile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	dg, err := discordgo.New("Bot " + *token)
	if err != nil {
		log.Fatal("Failed to connect to Discord: ", err)
		return
	}

	// add handlers
	dg.AddHandler(ready)

	// connect!
	err = dg.Open()
	if err != nil {
		log.Fatal("Failed to connect to Discord: ", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	defer dg.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("ComSat Station online!")
	log.Println("Joining servers: ", event.Guilds)

	for _, server := range event.Guilds {
		for _, targetChan := range server.Channels {
			if targetChan.Name == *channel {
				message, err := s.ChannelMessageSend(targetChan.ID, "Bleep bleep bleep!")
				if err != nil {
					log.Println("Error sending message: ", err)
				} else {
					log.Println("Sent message ID: " + message.ID)
				}
			}
		}
	}

	err := s.UpdateStatus(0, "hard to get...")
	if err != nil {
		log.Println("Unexpected error: ", err)
	}
}

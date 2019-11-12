package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/vaidashwin/comsat/configuration"
	"github.com/vaidashwin/comsat/storage"
	"github.com/vaidashwin/comsat/twitch"
	"math/rand"
	"os/signal"
	"strings"
	"syscall"

	"log"
	"os"
)

var token *string
var Config *string

func init() {
	// cli params
	token = flag.String("token", "", "Bot token from Discord API.")
	Config = flag.String("config", "", "Configuration of servers/channels/streams.")
	flag.Parse()

	if *token == "" || *Config == "" {
		panic("Token and config must be nonempty.")
	}

	configuration.InitConfig(*Config)
	storage.InitStorage()
}

var messageObject = discordgo.MessageEmbed{
	Title:       "ComSat Station",
	Description: "v0.1 - Stream monitoring bot",
	Color:       3036836,
	Footer: &discordgo.MessageEmbedFooter{
		IconURL: "https://i.imgur.com/i42Svek.png",
	},
	Image: &discordgo.MessageEmbedImage{
		URL: "https://i.imgur.com/UbdQ8nj.png",
	},
	Thumbnail: &discordgo.MessageEmbedThumbnail{
		URL: "https://media1.tenor.com/images/24318597c7ce1477f54587dd1616f3e8/tenor.gif",
	},
	Author: &discordgo.MessageEmbedAuthor{
		URL:     "https://github.com/vaidashwin/comsat",
		Name:    "Created by des",
		IconURL: "https://i.imgur.com/i42Svek.png",
	},
	Fields: nil,
}

func main() {
	if configuration.Get().Logfile != "" {
		// init logfile
		f, err := os.OpenFile(configuration.Get().Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
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
	dg.AddHandler(onGuildCreate)
	dg.AddHandler(onMessage)

	// connect!
	err = dg.Open()
	if err != nil {
		log.Fatal("Failed to connect to Discord: ", err)
		return
	}
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("ComSat Station online!")
	err := s.UpdateStatus(0, "hard to get...")
	if err != nil {
		log.Println("Unexpected error: ", err)
	}
	twitch.InitTwitch(createWebhookCallback(s))
}

func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	log.Println("Joining server:", event.Name)
	if channel, ok := configuration.Get().ServerToChannel[event.ID]; ok {
		log.Println("Found channel", channel, "for this server.")
		updateMessage(nil)
		if messageToEdit := storage.GetMessageIdForGuild(event.ID); messageToEdit != "" {
			s.ChannelMessageEditEmbed(channel, messageToEdit, &messageObject)
		} else {
			message, err := s.ChannelMessageSendEmbed(channel, &messageObject)
			if err == nil {
				storage.SetMessageIdForGuild(event.ID, message.ID)
			}
		}
	} else {
		log.Println("No channel defined for server", event.ID, event.Name)
	}
}

func onMessage(s *discordgo.Session, event *discordgo.MessageCreate) {
	if configuration.Get().ComsatSecret != "" &&
		strings.HasPrefix(event.Message.Content, configuration.Get().ComsatSecret) {
		tokens := strings.Split(strings.TrimPrefix(event.Message.Content, configuration.Get().ComsatSecret+" "), " ")
		log.Println(tokens)
		if len(tokens) <= 0 {
			return
		}
		switch tokens[0] {
		case "addTwitchStreamers":
			log.Println("Adding streamers...")
			streamers := make(map[string]string)
			for alias := 1; alias < len(tokens)-1; alias += 2 {
				streamerAlias := tokens[alias]
				streamName := tokens[alias+1]
				streamers[streamName] = streamerAlias
			}
			log.Println("New streamers:", streamers)
			err := storage.AddTwitchStreamers(streamers)
			if err != nil {
				log.Println("Error adding twitch streamers:", err)
			}
		}
		log.Println("Bot message created:", event.ID)
	}
}

func updateMessage(casters []twitch.StreamData) {
	var scanSounds = []string{
		"Activating my legal maphack!",
		"I'm in ur base lookin at ur doods.",
		"Tastosis calls it 'starsense' but I call it hacking.",
	}
	messageObject.Footer.Text = scanSounds[rand.Intn(len(scanSounds))]
	streamLinks := make([]*discordgo.MessageEmbedField, len(casters))
	for index, caster := range casters {
		streamLinks[index] = &discordgo.MessageEmbedField{
			Name:  caster.Streamer.GetAlias(),
			Value: fmt.Sprintf("%v - %v [%d viewers]", caster.Streamer.GetLink(), caster.Title, caster.Viewers),
		}
	}

	messageObject.Fields = streamLinks
}

func createWebhookCallback(s *discordgo.Session) func([]twitch.StreamData) {
	return func(liveCasters []twitch.StreamData) {
		log.Println("Firing callback for casters:", liveCasters)
		for guildID, messageID := range storage.GetMessages() {
			channelID := configuration.Get().ServerToChannel[guildID]
			updateMessage(liveCasters)
			_, err := s.ChannelMessageEditEmbed(channelID, messageID, &messageObject)
			if err != nil {
				log.Println("Got an error updating message?", err)
			}
		}
	}
}

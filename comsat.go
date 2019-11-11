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

func main() {
	if configuration.Get().Logfile != "" {
		// init logfile
		f, err := os.OpenFile(configuration.Get().Logfile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
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
		if messageToEdit := storage.GetMessageIdForGuild(event.ID); messageToEdit != "" {
			s.ChannelMessageEdit(channel, messageToEdit, getMessage(nil))
		} else {
			message, err := s.ChannelMessageSend(channel, getMessage(nil))
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
		log.Println("Bot message created:", event.ID)
	}
}

func getMessage(casters []twitch.StreamData) string {
	var scanSounds = []string {
		"Activating my legal maphack!",
		"I'm in ur base lookin at ur doods.",
		"Tastosis calls it 'starsense' but I call it hacking.",
	}

	result := "**" + scanSounds[rand.Intn(len(scanSounds))] + "**\n\n" +
		"*Active streams*:\n"
	if len(casters) == 0 {
		result += "No one is casting, banking up another 50 energy."
	} else {
		for _, caster := range casters {
			result += fmt.Sprintf("* %v (%v) - %v [%d viewers]",
				caster.UserName,
				caster.URL,
				caster.Title,
				caster.Viewers)
		}
	}
	return result
}

func createWebhookCallback(s *discordgo.Session) func([]twitch.StreamData) {
	return func(liveCasters []twitch.StreamData) {
		log.Println("Firing callback for casters:", liveCasters)
		for guildID, messageID := range storage.GetMessages() {
			channelID := configuration.Get().ServerToChannel[guildID]
			_, err := s.ChannelMessageEdit(channelID, messageID, getMessage(liveCasters))
			if err != nil {
				log.Println("Got an error updating message?", err)
			}
		}
	}
}
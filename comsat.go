package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/vaidashwin/comsat/configuration"
	"github.com/vaidashwin/comsat/storage"
	"github.com/vaidashwin/comsat/twitchhook"
	"math/rand"
	"os/signal"
	"syscall"

	"log"
	"os"
)

var logfile *string
var token *string
var Config *string

func init() {
	// cli params
	token = flag.String("token", "", "Bot token from Discord API.")
	logfile = flag.String("logfile", "", "Logfile.")
	Config = flag.String("config", "", "Configuration of servers/channels/streams.")
	flag.Parse()

	if *token == "" || *Config == "" {
		panic("Token and config must be nonempty.")
	}

	configuration.InitConfig(*Config)
	storage.InitStorage()
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
	dg.AddHandler(onGuildCreate)

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
	twitchhook.InitTwitchhooks(createWebhookCallback(s))
}

func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	log.Println("Joining server:", event.Name)
	if channel, ok := configuration.Get().ServerToChannel[event.ID]; ok {
		log.Println("Found channel", channel, "for this server.")
		casters := []string{"caster"}
		if messageToEdit := storage.GetMessageIdForGuild(event.ID); messageToEdit != "" {
			s.ChannelMessageEdit(channel, messageToEdit, getMessage(casters))
		} else {
			message, err := s.ChannelMessageSend(channel, getMessage(casters))
			if err == nil {
				storage.SetMessageIdForGuild(event.ID, message.ID)
			}
		}
	} else {
		log.Println("No channel defined for server", event.ID, event.Name)
	}
}

func getMessage(casters []string) string {
	var scanSounds = []string {
		"Activating my legal maphack!",
	}

	result := "**" + scanSounds[rand.Intn(len(scanSounds))] + "**\n\n" +
		"*Active streams*:\n"
	for _, caster := range casters {
		result += "* " + caster + "\n"
	}
	return result
}

func createWebhookCallback(s *discordgo.Session) func([]string) {
	return func(liveCasters []string) {
		for channelID, messageID := range storage.GetMessages() {
			s.ChannelMessageEdit(channelID, messageID, getMessage(liveCasters))
		}
	}
}
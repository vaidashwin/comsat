package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vaidashwin/comsat/configuration"
	"log"
	"os"
)

var db *sql.DB

func InitStorage() {
	var err error
	if _, err := os.Stat(configuration.Get().StorageDirectory + "/comsat.db"); os.IsNotExist(err) {
		err = os.MkdirAll(configuration.Get().StorageDirectory, 0666)
		if err != nil {
			log.Fatal(err)
		}
		_, err = os.Create(configuration.Get().StorageDirectory + "/comsat.db")
		if err != nil {
			log.Fatal(err)
		}
	}
	db, err = sql.Open("sqlite3", configuration.Get().StorageDirectory + "/comsat.db")
	if err != nil {
		log.Fatal(err)
	}
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS guild_messages (guild_id varchar(64) primary key, message_id varchar(64))")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
}

func GetMessageIdForGuild(guildID string) string {
	rows, err := db.Query("SELECT message_id FROM guild_messages WHERE guild_id = " + guildID)
	if err != nil {
		log.Println(err)
		return ""
	}
	var result string = ""
	rows.Next()
	rows.Scan(&result)
	return result
}

func GetMessages() map[string]string {
	result := make(map[string]string)
	rows, err := db.Query("SELECT guild_id, message_id FROM guild_messages")
	if err != nil {
		log.Println(err)
		return result
	}
	for rows.Next() {
		var guildID string
		var messageID string
		rows.Scan(&guildID, &messageID)
		result[guildID] = messageID
	}
	return result
}

func SetMessageIdForGuild(guildID string, messageID string) {
	statement, err := db.Prepare("INSERT INTO guild_messages VALUES (?,?)" +
		"ON CONFLICT(guild_id) DO UPDATE SET message_id = ? WHERE guild_id = ?")
	if err != nil {
		log.Println(err)
		return
	}
	statement.Exec(guildID, messageID, messageID, guildID)
}
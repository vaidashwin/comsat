package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/vaidashwin/comsat/configuration"
)

var db *sql.DB

func InitStorage() {
	var err error
	db, err = sql.Open("sqlite3", configuration.Get().StorageFile)
	if err != nil {
		log.Fatal(err)
	}
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS guild_channels (guild_id varchar(64) primary key, channel_id varchar(64))")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
}

func GetMessageIdForGuild(guildID string) string {
	rows, err := db.Query("SELECT channel_id FROM guild_channels WHERE guild_id = " + guildID)
	if err != nil {
		log.Println(err)
		return ""
	}
	var result string = ""
	rows.Next()
	rows.Scan(&result)
	return result
}

func SetMessageIdForGuild(guildID string, channelID string) {
	statement, err := db.Prepare("INSERT INTO guild_channels VALUES (?,?)" +
		"ON CONFLICT(guild_id) DO UPDATE SET channel_id = ? WHERE guild_id = ?")
	if err != nil {
		log.Println(err)
		return
	}
	statement.Exec(guildID, channelID, channelID, guildID)
}
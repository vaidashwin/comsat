package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vaidashwin/comsat/model"
	"log"
	"os"
)

var db *sql.DB

func InitStorage() {
	var err error
	if _, err := os.Stat("/etc/storage/comsat.db"); os.IsNotExist(err) {
		err = os.MkdirAll("/etc/storage", 0666)
		if err != nil {
			log.Fatal(err)
		}
		_, err = os.Create("/etc/storage/comsat.db")
		if err != nil {
			log.Fatal(err)
		}
	}
	db, err = sql.Open("sqlite3", "/etc/storage/comsat.db")
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

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS streams (stream_id int primary key, service text, display_name text, alias text, url varchar(64))")
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

func SetMessageIdForGuild(guildID string, messageID string) error {
	statement, err := db.Prepare("INSERT INTO guild_messages VALUES (?,?)" +
		"ON CONFLICT(guild_id) DO UPDATE SET message_id = ? WHERE guild_id = ?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(guildID, messageID, messageID, guildID)
	return err
}

func AddStream(id int, service string, displayName string, alias string, url string) error {
	statement, err := db.Prepare("INSERT INTO streams VALUES(?, ?, ?, ?, ?)" +
		"ON CONFLICT(stream_id) DO UPDATE SET service = ?, display_name = ?, alias = ?, url = ? WHERE stream_id = ?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(id, service, displayName, alias, url, service, displayName, alias, url, id)
	return err
}

func GetStreams(service string) ([]model.Streamer, error) {
	result := make([]model.Streamer, 0)
	statement, err := db.Prepare("SELECT stream_id, display_name, alias, url FROM streams WHERE service = ?")
	if err != nil {
		return result, err
	}
	rows, err := statement.Query(service)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var id int
		var displayName string
		var alias string
		var url string
		err = rows.Scan(&id, &displayName, &alias, &url)
		if err != nil {
			continue
		}
		var streamer model.Streamer
		switch service {
		case "twitch":
			streamer = NewStreamer(displayName, alias, url, id)
		default:
			continue
		}
		result = append(result, streamer)
	}
	return result, nil
}
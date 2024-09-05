package main

import (
	"log"
	"slugbot/db"
	"slugbot/helix"
	"slugbot/models"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gempir/go-twitch-irc/v4"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	var config models.Config
	_, err := toml.DecodeFile("./config.toml", &config)
	if err != nil {
		log.Fatalf("Could not load config: %s", err)
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{config.Clickhouse.Host},
		Auth: clickhouse.Auth{
			Database: config.Clickhouse.Database,
			Username: config.Clickhouse.User,
			Password: config.Clickhouse.Password,
		},
	})
	if err != nil {
		log.Fatalf("Could not open clickhouse connection: %s", err)
	}
	if err := db.CreateTable(conn); err != nil {
		log.Fatalf("Could not create tables: %s", err)
	}

	helix, err := helix.New(config.Helix.ClientID, config.Helix.ClientSecret, config.Channels, config.Debug)
	if err != nil {
		log.Fatalf("Could not create Helix client: %s", err)
	}
	log.Println("Created Helix client")

	irc := twitch.NewAnonymousClient()
	irc.OnConnect(func() { log.Printf("Connected to IRC") })
	irc.OnPrivateMessage(func(msg twitch.PrivateMessage) {
		channelID, err := strconv.Atoi(msg.RoomID)
		if err != nil {
			log.Printf("Could not convert to string: %s\n", err)
			return
		}
		senderID, err := strconv.Atoi(msg.User.ID)
		if err != nil {
			log.Printf("Could not convert to string: %s\n", err)
			return
		}
		// VIP is specified by the existance of the vip tag, not the value
		_, vip := msg.Tags["vip"]

		message := db.Log{
			ID:                msg.ID,
			Channel:           msg.Channel,
			ChannelID:         int32(channelID),
			SenderDisplayName: msg.User.DisplayName,
			SenderUsername:    msg.User.Name,
			SenderID:          int32(senderID),
			SenderColor:       msg.User.Color,
			Message:           msg.Message,
			SentAt:            msg.Time,
			Live:              helix.LiveChannels[msg.Channel],
			Vip:               vip,
			Mod:               msg.Tags["mod"] == "1",
			Sub:               msg.Tags["subscriber"] == "1",
		}
		if config.Debug {
			log.Printf("[#%s] %s: %s", message.Channel, message.SenderUsername, message.Message)
		}
		if err := db.InsertMessage(conn, message); err != nil {
			log.Printf("Could not insert message: %s", err)
		}
	})
	log.Printf("Joining %d channels", len(config.Channels))
	irc.Join(config.Channels...)
	if err := irc.Connect(); err != nil {
		log.Printf("Could not connect to IRC: %s", err)
	}
}

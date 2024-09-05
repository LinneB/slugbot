package db

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type Log struct {
	ID                string
	Channel           string
	ChannelID         int32
	SenderDisplayName string
	SenderUsername    string
	SenderID          int32
	SenderColor       string
	Message           string
	SentAt            time.Time
	Live              bool
	Vip               bool
	Mod               bool
	Sub               bool
}

func CreateTable(conn clickhouse.Conn) error {
	query := `CREATE TABLE IF NOT EXISTS logs (
    id UUID DEFAULT generateUUIDv4(),
    channel String,
    channel_id Int32,
    sender_displayname String,
    sender_username String,
    sender_userid Int32,
    sender_color String,
    message String,
    sent_at DateTime64(3),
    live Bool,
    vip Bool,
    mod Bool,
    sub Bool
)
ENGINE = MergeTree
ORDER BY sent_at
PARTITION BY channel;`

	return conn.Exec(context.Background(), query)
}

func InsertMessage(conn clickhouse.Conn, m Log) error {
	query := `
INSERT INTO logs (id, channel, channel_id, sender_displayname, sender_username, sender_userid, sender_color, message, sent_at, live, vip, mod, sub)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	return conn.Exec(context.Background(), query,
		m.ID,
		m.Channel,
		m.ChannelID,
		m.SenderDisplayName,
		m.SenderUsername,
		m.SenderID,
		m.SenderColor,
		m.Message,
		m.SentAt.UnixMilli(),
		m.Live,
		m.Vip,
		m.Mod,
		m.Sub,
	)
}

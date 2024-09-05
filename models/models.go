package models

type Config struct {
	Debug      bool
	Channels   []string
	Clickhouse struct {
		Host     string
		Database string
		User     string
		Password string
	}
	Helix struct {
		ClientID     string
		ClientSecret string
	}
}

# Slugbot

Slugbot logs Twitch chat messages into a Clickhouse database.

This project is a rewrite of [logbot](https://github.com/LinneB/logbot). If you have an instance of the logbot running, see [Migrating](#migrating-from-postgresql).

## Setup

### Configuration

```toml
# Debug log output
debug = true
# Channels to log
channels = ["forsen", "psp1g"]

# Helix application details
[helix]
clientid = "myclientid"
clientsecret = "freshclientsecret"

# Clickhouse DB details
[clickhouse]
# If you are using Docker Compose, you can use the container name as the hostname (slugbot-clickhouse by default).
host = "127.0.0.1:9000"
database = "slugbot"
user = "slugbot"
password = "slugbot"
```

### Docker Compose

There is a `docker-compose.yml` file included that features everything you will need:

```yaml
services:
  slugbot:
    build: .
    container_name: slugbot-bot
    volumes:
      - "./config.toml:/app/config.toml"
    depends_on:
      clickhouse:
        condition: service_healthy
  clickhouse:
    image: clickhouse/clickhouse-server:latest
    container_name: slugbot-clickhouse
    environment:
      CLICKHOUSE_DB: "slugbot"
      CLICKHOUSE_USER: "slugbot"
      CLICKHOUSE_PASSWORD: "slugbot"
    volumes:
      - "./db-data:/var/lib/clickhouse"
    healthcheck:
      test: wget --no-verbose --tries=1 --spider http://localhost:8123/ping || exit 1
      interval: 5s
```

## Migrating from PostgreSQL

The SQL schema is more or less the same, only that the Twitch message ID is now also stored.
You can still import old logs from PostgreSQL, **but it will generate a NEW random UUID**.

```sql
INSERT INTO logs (channel, channel_id, sender_displayname, sender_username, sender_userid, sender_color, message, sent_at, live, vip, mod, sub)
SELECT channel, channel_id, sender_displayname, sender_username, sender_userid, sender_color, message, sent_at, live, vip, mod, sub
FROM postgresql('host:port', 'database', 'table', 'user', 'password')
```

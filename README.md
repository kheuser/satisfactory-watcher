# satisfactory-watcher

A lightweight Linux daemon written in Go that tails a Satisfactory Dedicated Server log file and posts a Discord webhook notification whenever a player connects.

## Features

- **Log rotation resilient** — Automatically reopens `FactoryGame.log` when Unreal Engine rotates logs on server restarts.
- **No spam on startup** — Seeks to the end of the log file so only new connections trigger alerts.
- **Graceful shutdown** — Handles `SIGINT` and `SIGTERM` cleanly for `systemd` integration.

## Prerequisites

- Go 1.21+
- A Discord channel webhook URL
- Read access to your Satisfactory server's `FactoryGame.log`

## Installation

```bash
git clone [https://github.com/kheuser/satisfactory-watcher.git](https://github.com/kheuser/satisfactory-watcher.git)
cd satisfactory-watcher
go build -o satisfactory-watcher .
sudo mv satisfactory-watcher /usr/local/bin/
```

## Usage

```bash
satisfactory-watcher -log /path/to/FactoryGame/Saved/Logs/FactoryGame.log -webhook [https://discord.api/webhooks/YOUR_ID/YOUR_TOKEN](https://discord.api/webhooks/YOUR_ID/YOUR_TOKEN)
```

## Running with systemd

Create a new service at `/etc/systemd/system/satisfactory-watcher.service`

```bash
[Unit]
Description=Satisfactory Log Watcher for Discord Alerts
After=network-online.target

[Service]
Type=simple
User=steam ## Or whatever you named the user for the Satisfactory Dedicated Server
Environment="DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/YOUR_ID/YOUR_TOKEN"
ExecStart=/usr/local/bin/satisfactory-watcher -log /path/to/Factory/Saved/Logs/FactoryGame.log
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```
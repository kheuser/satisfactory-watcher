package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/nxadm/tail"
)

var joinRegex = regexp.MustCompile(`LogNet:\s+Join succeeded:\s+(.+)$`)

type discordPayload struct {
	Content string `json:"content"`
}

func main() {
	logPath := flag.String("log", "", "Path to Satisfactory FactoryGame.log")
	webhookURL := flag.String("webhook", os.Getenv("DISCORD_WEBHOOK_URL"), "Discord webhook URL (or set DISCORD_WEBHOOK_URL)")
	flag.Parse()

	if *logPath == "" || *webhookURL == "" {
		log.Fatal("Usage: satisfactory-watcher -log /path/to/FactoryGame.log -webhook https://discord.com/api/webhooks/...")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	t, err := tail.TailFile(*logPath, tail.Config{
		Follow:    true,
		ReOpen:    true, // Handles log rotation on server restart
		MustExist: false,
		Location:  &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}, // Only alert on new logins
		Logger:    tail.DiscardingLogger,
	})
	if err != nil {
		log.Fatalf("Failed to tail file: %v", err)
	}
	defer t.Cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	log.Printf("Watching %s for player logins...", *logPath)

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")
			_ = t.Stop()
			return
		case line, ok := <-t.Lines:
			if !ok {
				return
			}
			if line.Err != nil {
				log.Printf("Tail error: %v", line.Err)
				continue
			}

			if matches := joinRegex.FindStringSubmatch(line.Text); len(matches) == 2 {
				player := strings.TrimSpace(matches[1])
				log.Printf("Player connected: %s", player)
				if err := sendDiscordAlert(ctx, client, *webhookURL, player); err != nil {
					log.Printf("Failed to send webhook: %v", err)
				}
			}
		}
	}
}

func sendDiscordAlert(ctx context.Context, client *http.Client, webhookURL, player string) error {
	payload := discordPayload{
		Content: fmt.Sprintf("🟢 **%s** has connected to the Satisfactory server.", player),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
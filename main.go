package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"runtime/debug"
	"slices"
	"strconv"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
	"github.com/mcstatus-io/mcutil/v4/query"
	"github.com/mcstatus-io/mcutil/v4/response"
)

var ip = ""
var port uint16 = 0
var webhookUrl = ""

var oldStatus *response.QueryFull = &response.QueryFull{}

func main() {
	loadEnvironmentVars()

	// create a scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}

	fmt.Println("Querying Server at " + ip + ":" + fmt.Sprint(port))

	sendMessage("Bot", "", "**Restarted!**")

	app()

	// add a job to the scheduler
	_, err = s.NewJob(
		gocron.DurationJob(
			30*time.Second,
		),
		gocron.NewTask(
			func() {
				app()
			},
		),
	)
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}

	// start the scheduler
	s.Start()

	// block forever
	select {}
}

func app() {

	status := getServerStatus()

	oldPlayers := oldStatus.Players
	newPlayers := status.Players

	fmt.Println("Online", newPlayers)

	allPlayers := append(oldPlayers, newPlayers...)

	for _, player := range allPlayers {
		inNew := slices.Contains(newPlayers, player)
		inOld := slices.Contains(oldPlayers, player)

		if inOld && !inNew {
			if rand.Intn(30) == 0 {
				switch player {
				case "Nocnava_":
					sendPlayerMessage(player, "**Crashed**")
				case "Ryanosaurus":
					sendPlayerMessage(player, "**Nuked his PC**")
				}
			} else {
				sendPlayerMessage(player, "**Left the game**")
			}
		}

		if !inOld && inNew {
			sendPlayerMessage(player, "**Joined the game**")
		}
	}

	if rand.Intn(3400) == 0 {
		sendPlayerMessage("Your_Da", "**Joined the game**")
	}

	oldStatus = status
}

func loadEnvironmentVars() {
	godotenv.Load()

	ip = os.Getenv("MC_DISCORD_IP")

	parsedPort, err := strconv.Atoi(os.Getenv("MC_DISCORD_PORT"))
	if err == nil {
		port = uint16(parsedPort)
	}

	url, found := os.LookupEnv("MC_DISCORD_WEBHOOK")
	if !found {
		panic("Env Var not set for MC_DISCORD_WEBHOOK")
	}
	webhookUrl = url
}

type Profile struct {
	Id string `json:"id"`
}

func getMinecraftProfile(player string) (*Profile, error) {

	resp, err := http.Get("https://api.mojang.com/users/profiles/minecraft/" + player)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var profile Profile

	err = json.Unmarshal(body, &profile)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return &profile, nil
}

func sendPlayerMessage(player string, message string) {

	profile, err := getMinecraftProfile(player)
	uuid := profile.Id

	if err != nil {
		debug.PrintStack()
		slog.Error("Error", err)

		uuid = "60e832dbfeb745e2b92fe82e5132bf03"
	}

	webhook := DiscordWebhook{
		Username:  player,
		AvatarURL: fmt.Sprintf("https://api.mineatar.io/face/" + uuid),
		Content:   message,
		Flags:     4096, // Silent Messages
	}

	jsonData, _ := json.Marshal(webhook)

	_, err = http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		debug.PrintStack()
		slog.Error("Error", err)
	}
}

func sendMessage(username string, avatarUrl string, content string) {

	webhook := DiscordWebhook{
		Username:  username,
		AvatarURL: avatarUrl,
		Content:   content,
		Flags:     4096, // Silent Messages
	}

	jsonData, _ := json.Marshal(webhook)

	_, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		debug.PrintStack()
		slog.Error("Error", err)
	}
}

func getServerStatus() *response.QueryFull {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	response, err := query.Full(ctx, ip, port)

	if err != nil {
		panic(err)
	}

	return response

}

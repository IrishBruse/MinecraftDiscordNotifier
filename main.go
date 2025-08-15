package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
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
var port uint16 = 26585
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

	fmt.Println("Querying " + ip + ":" + fmt.Sprint(port))

	status := getServerStatus()

	oldPlayers := oldStatus.Players
	newPlayers := status.Players

	fmt.Println("Online", newPlayers)

	allPlayers := append(oldPlayers, newPlayers...)

	for _, player := range allPlayers {
		inNew := slices.Contains(newPlayers, player)
		inOld := slices.Contains(oldPlayers, player)

		if inOld && !inNew {
			if rand.Intn(50) == 0 {
				switch player {
				case "Nocnava_":
					sendMessage(player, "**Crashed**")
				case "Ryanosaurus":
					sendMessage(player, "**PC Exploded :myan:**")
				}
			} else {
				sendMessage(player, "**Left the game**")
			}
		}

		if !inOld && inNew {
			sendMessage(player, "**Joined the game**")
		}
	}

	if rand.Intn(7200) == 0 {
		sendMessage("Herobrine", "**Joined the game**")
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

func sendMessage(player string, message string) {
	webhook := DiscordWebhook{
		Username:  player,
		AvatarURL: fmt.Sprintf("https://minotar.net/avatar/" + player + ".png"),
		Content:   message,
		Flags:     4096, // Silent Messages
	}

	jsonData, _ := json.Marshal(webhook)

	_, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
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

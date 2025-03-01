package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"minecraft-discord-notifier/packet"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
)

var ip = "127.0.0.1"
var port int16 = 25565

const red int64 = 15598853
const green int64 = 388613

var oldStatus Status

func main() {
	// create a scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}

	// add a job to the scheduler
	_, err = s.NewJob(
		gocron.DurationJob(
			2*time.Minute,
		),
		gocron.NewTask(
			func(a string, b int) {
				app()
			},
			"hello",
			1,
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

	fmt.Println("App Starting alpine")

	ip = os.Getenv("MC_DISCORD_IP")
	parsedPort, err := strconv.Atoi(os.Getenv("MC_DISCORD_PORT"))
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}

	port = int16(parsedPort)

	status := getServerStatus()

	oldPlayers := oldStatus.Players.Sample
	newPlayers := status.Players.Sample

	allPlayers := append(oldPlayers, newPlayers...)

	for _, player := range allPlayers {
		inNew := slices.Contains(newPlayers, player)
		inOld := slices.Contains(oldPlayers, player)

		if inOld && !inNew {
			if player.Name == "Nocnava_" && rand.Intn(10) == 0 {
				sendMessage(player, "**Crashed**")
			} else {
				sendMessage(player, "**Left the game**")
			}
		}

		if !inOld && inNew {
			sendMessage(player, "**Joined the game**")
		}
	}

	oldStatus = status
}

func sendMessage(player Player, message string) {
	webhook := DiscordWebhook{
		Username:  player.Name,
		AvatarURL: fmt.Sprintf("https://minotar.net/avatar/" + strings.ReplaceAll(player.ID, "-", "") + ".png"),
		Content:   message,
		Flags:     4096, // Silent Messages
	}

	jsonData, _ := json.Marshal(webhook)

	_, err := http.Post(os.Getenv("MC_DISCORD_WEBHOOK"), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
}

func getServerStatus() Status {
	c, err := net.Dial("tcp", ip+":"+strconv.Itoa((int)(port)))
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
	defer c.Close()

	handshake := packet.NewOutboundPacket(0)
	handshake.WriteVarInt(767)
	handshake.WriteString(ip)
	handshake.WriteShort(port)
	handshake.WriteVarInt(1)
	handshake.Write(c)

	statusReq := packet.NewOutboundPacket(0)
	statusReq.Write(c)

	statusRes, err := packet.NewInboundPacket(c, time.Duration(time.Second*30))
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}

	buf, err := statusRes.ReadString()
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}

	var status Status
	json.Unmarshal([]byte(buf), &status)

	return status
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"minecraft-discord-notifier/packet"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

var ip = "127.0.0.1"
var port int16 = 25565

const red int64 = 15598853
const green int64 = 388613

func main() {
	ip = os.Getenv("MC_DISCORD_IP")
	parsedPort, err := strconv.Atoi(os.Getenv("MC_DISCORD_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	port = int16(parsedPort)

	oldStatus := getOldStatus()
	status := getServerStatus()

	oldPlayers := oldStatus.Players.Sample
	newPlayers := status.Players.Sample

	allPlayers := append(oldPlayers, newPlayers...)

	for _, player := range allPlayers {
		inNew := slices.Contains(newPlayers, player)
		inOld := slices.Contains(oldPlayers, player)

		if inOld && !inNew {
			sendLeftMessage(player)
		}

		if !inOld && inNew {
			sendJoinMessage(player)
		}
	}

}

func getOldStatus() Status {
	buf, err := os.ReadFile("state.json")
	if err != nil {
		log.Fatal()
	}

	var status Status
	json.Unmarshal([]byte(buf), &status)

	return status
}

func sendJoinMessage(player Player) {
	webhook := DiscordWebhook{
		Username:  player.Name,
		AvatarURL: fmt.Sprintf("https://minotar.net/avatar/" + strings.ReplaceAll(player.ID, "-", "") + ".png"),
		Content:   "joined the game",
	}

	jsonData, _ := json.Marshal(webhook)

	_, err := http.Post(os.Getenv("MC_DISCORD_WEBHOOK"), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
}

func sendLeftMessage(player Player) {
	webhook := DiscordWebhook{
		Username:  player.Name,
		AvatarURL: fmt.Sprintf("https://minotar.net/avatar/" + strings.ReplaceAll(player.ID, "-", "") + ".png"),
		Content:   "left the game",
	}

	jsonData, _ := json.Marshal(webhook)

	_, err := http.Post(os.Getenv("DISCORD_WEBHOOK"), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
}

func getServerStatus() Status {
	c, err := net.Dial("tcp", ip+":"+strconv.Itoa((int)(port)))
	if err != nil {
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
		log.Fatal(err)
	}

	buf, err := statusRes.ReadString()
	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile("state.json", []byte(buf), os.FileMode(0777))

	var status Status
	json.Unmarshal([]byte(buf), &status)

	return status
}

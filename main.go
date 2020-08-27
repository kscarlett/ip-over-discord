package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/songgao/water"
)

func main() {
	recipientID := flag.String("recipient", "", "ID of the user to DM")
	secret := flag.String("secret", "", "Bot token from Discord")

	flag.Parse()

	tunnel, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		log.Fatalln("failed to create device: ", err)
	}
	defer tunnel.Close()

	log.Println("created device ", tunnel.Name())

	// Connect to Discord
	session, err := discordgo.New(fmt.Sprintf("bot %s", *secret))
	if err != nil {
		log.Fatalln("cannot connect to discord: ", err)
	}

	// Get the DM channel
	channel, err := session.UserChannelCreate(*recipientID)
	if err != nil {
		log.Fatalln("could not create channel with user: ", err)
	}

	// Register callback for received messages and handle through interface
	// Loop handling of callback
	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == *recipientID {
			packet, err := base64.StdEncoding.DecodeString(m.Message.Content)
			if err != nil {
				log.Println("failed to encode packet: ", err)
			}
			log.Printf("packet recv: % x\n", packet)
			_, err = tunnel.Write(packet)
			if err != nil {
				log.Printf("failed to write packet to device %s: %s\n", tunnel.Name(), err)
			}
		}
	})

	// Infinite loop for handling packets and send them to Discord
	readAndSend(tunnel, session, channel.ID)
}

func readAndSend(ifce *water.Interface, session *discordgo.Session, channelID string) {
	packet := make([]byte, 2000)
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			log.Println("failed to read packet from tunnel: ", err)
		}

		_, err = session.ChannelMessageSend(channelID, base64.StdEncoding.EncodeToString(packet[:n]))
		if err != nil {
			log.Println("failed to send packet to discord: ", err)
		} else {
			log.Printf("packet sent: % x\n", packet)
		}
	}
}

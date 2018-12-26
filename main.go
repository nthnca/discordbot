package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	HouseCup map[string]int64
)

func main() {
	var Token string
	flag.StringVar(&Token, "t", "", "Token")
	flag.Parse()

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error discordgo.New: ", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error dg.Open: ", err)
		return
	}

	// Catch signals.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Prevent loops.
	if m.Author.ID == s.State.User.ID {
		return
	}

	houseCupHandler(s, m)
}

func houseCupHelper(s string) *UserScore {
	re := regexp.MustCompile("<@([0-9]+)>[^@]*(award|penal).* ([0-9]+)")
	m := re.FindStringSubmatch(s)
	if m == nil {
		log.Printf("No match %v", s)
		return nil
	}

	log.Printf("Nice %#v", m)
	log.Printf("Nice %v", s)
	v, err := strconv.Atoi(m[3])
	if err != nil {
		log.Printf("Invalid int %v", err)
		return nil
	}

	if m[2] == "penal" {
		v = -1 * v
	}

	var u UserScore
	u.UserId = m[1]
	u.Score = int64(v)
	return &u
}

func houseCupHandler(s *discordgo.Session, msg *discordgo.MessageCreate) {
	p := houseCupHelper(msg.Content)
	if p == nil {
		return
	}

	if HouseCup == nil {
		HouseCup = make(map[string]int64)
	}
	HouseCup[p.UserId] += p.Score

	s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("<@%s> has been awarded %d points!", p.UserId, p.Score))

	output := ""
	for k, v := range HouseCup {
		output += fmt.Sprintf("\n<@%s>: %d", k, v)
	}
	s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("The !House Cup! score is currently:\n%s", output))
}

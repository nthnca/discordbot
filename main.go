package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/golang/protobuf/proto"
)

const (
	state_filename = "discord_bot_state"
)

var (
	HouseCup map[string]int64
)

func init() {
	HouseCup = make(map[string]int64)
	houseCupLoad()
}

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

func houseCupPersist() {
	var state DiscordBotState
	for k, v := range HouseCup {
		var u UserScore
		u.UserId = k
		u.Score = int64(v)
		state.HouseCupScore = append(state.HouseCupScore, &u)
	}

	data, err := proto.Marshal(&state)
	if err != nil {
		log.Fatalf("marshalling proto: %v", err)
	}

	if er := ioutil.WriteFile(state_filename, data, 0644); er != nil {
		log.Fatalf("writing file: %v", err)
	}

}

func houseCupLoad() {
	data, err := ioutil.ReadFile(state_filename)
	if err != nil {
		log.Printf("%v", err)
	}

	var state DiscordBotState
	if er := proto.Unmarshal(data, &state); er != nil {
		log.Fatalf("Unmarshalling proto: %v", er)
	}

	for _, v := range state.HouseCupScore {
		HouseCup[v.UserId] += v.Score
	}
}

func houseCupHandler(s *discordgo.Session, msg *discordgo.MessageCreate) {
	p := houseCupHelper(msg.Content)
	if p == nil {
		return
	}

	HouseCup[p.UserId] += p.Score
	houseCupPersist()

	s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("<@%s> has been awarded %d points!", p.UserId, p.Score))

	output := ""
	for k, v := range HouseCup {
		output += fmt.Sprintf("\n<@%s>: %d", k, v)
	}
	s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("The !House Cup! score is currently:\n%s", output))
}

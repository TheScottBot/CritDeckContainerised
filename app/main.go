package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Deck struct {
	Cards []Card `json:"cards"`
}

type Card struct {
	Strength string   `json:"strength"`
	Effects  []Effect `json:"effects"`
}

type Effect struct {
	DamageType string `json:"damageType"`
	Value      string `json:"value"`
}

var client *http.Client
var Token = ""
var BotPrefix = "?"
var deckInPlay Deck
var deck Deck
var currentCard = 0

func setupDefaults() {
	Token = os.Getenv("TOKEN")

	fmt.Println("token set")
}

func loadPlayerDeck() {
	jsonFile, err := os.Open("playerDeck.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &deck)
}

func addToDeckInPlay(cardType string) {
	for _, y := range deck.Cards {
		s := strings.Trim(y.Strength, "\r\n")
		if s == cardType {
			deckInPlay.Cards = append(deckInPlay.Cards, y)
		}
	}
	shuffleDeckInPlay()
}

func shuffleDeckInPlay() {
	rand.Shuffle(len(deckInPlay.Cards), func(i, j int) { deckInPlay.Cards[i], deckInPlay.Cards[j] = deckInPlay.Cards[j], deckInPlay.Cards[i] })
}

func draw() string {
	if currentCard <= len(deckInPlay.Cards)-1 {
		var card strings.Builder
		for _, z := range deckInPlay.Cards[currentCard].Effects {
			card.WriteString(fmt.Sprintf("Damage Type: %s\n%s\n\n", z.DamageType, z.Value))
			fmt.Println("    Damage type: ", z.DamageType, "\n        ", z.Value)
		}
		currentCard++
		return card.String()
	} else {
		shuffleDeckInPlay()
		currentCard = 0
	}
	return ""
}

func createDeckInPlay(level int) {
	if level >= 0 {
		addToDeckInPlay("Setback")
	}
	if level >= 5 {
		addToDeckInPlay("Dangerous")
	}
	if level >= 9 {
		addToDeckInPlay("Life Threatening")
	}
	if level >= 13 {
		addToDeckInPlay("Deadly")
	}

	shuffleDeckInPlay()
}

var BotId string

func Start() {
	goBot, err := discordgo.New("Bot " + Token)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	u, err := goBot.User("@me")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	BotId = u.ID

	goBot.AddHandler(messageHandler)

	err = goBot.Open()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	goBot.Close()
}

func messageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	fmt.Println(message.Content)
	if message.Author.ID == BotId {
		fmt.Println("Breaking out")
		return
	}
	fmt.Println(message.Content)
	if message.Content == BotPrefix+"ping" {
		_, err := session.ChannelMessageSend(message.ChannelID, "pong")
		if err != nil {
			fmt.Println("error ", err)
		}
	}
	HandleNewDeck(message, session)
	HandleDraw(message, session)
	HandleHelp(message, session)
}

func HandleNewDeck(message *discordgo.MessageCreate, session *discordgo.Session) {
	content := strings.ToLower(message.Content)
	if strings.HasPrefix(content, BotPrefix+"newdeck") {
		deckInPlay.Cards = nil
		levelAsString := message.Content[len(BotPrefix)+len("newdeck") : len(message.Content)]
		levelAsInt, _ := strconv.Atoi(levelAsString)
		createDeckInPlay(levelAsInt)

		response := fmt.Sprintf("Crit deck created. It has: %v cards", len(deckInPlay.Cards))
		_, err := session.ChannelMessageSend(message.ChannelID, response)
		if err != nil {
			fmt.Println("error ", err)
		}
	}
}

func HandleDraw(message *discordgo.MessageCreate, session *discordgo.Session) {
	content := strings.ToLower(message.Content)
	if strings.HasPrefix(content, BotPrefix+"draw") {
		card := draw()
		response := fmt.Sprintf(card)
		_, err := session.ChannelMessageSend(message.ChannelID, response)
		if err != nil {
			fmt.Println("error ", err)
		}
	}
}

func HandleHelp(message *discordgo.MessageCreate, session *discordgo.Session) {
	content := strings.ToLower(message.Content)
	if strings.HasPrefix(content, BotPrefix+"help") {
		response := fmt.Sprintf("The bot uses the prefix ? and has 2 commands\n" +
			"?newdeck followed by the level of the party members as an integer (with no space between[I'm lazy and coded it badly]): for example ?newdeck7\n" +
			"?draw. Once a deck is created it will draw a random card and output it as a message")
		_, err := session.ChannelMessageSend(message.ChannelID, response)
		if err != nil {
			fmt.Println("error ", err)
		}
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	setupDefaults()

	loadPlayerDeck()

	Start()
}

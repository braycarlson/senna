package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/braycarlson/asol"
	"github.com/braycarlson/senna/model"
)

const (
	ddragon = "https://ddragon.leagueoflegends.com"
)

type (
	// API
	Runes = model.Runes

	// Data Dragon
	Realms = model.Realms

	// LCU
	ChampionSelection = model.ChampionSelection
	Gameflow          = model.Gameflow
	Login             = model.Login
	MatchFound        = model.MatchFound
	Page              = model.Page
	Phase             = model.Phase

	// Preferences
	ARAM         = model.ARAM
	Champion     = model.Champion
	ChampionData = model.ChampionData
	Classic      = model.Classic
	OneForAll    = model.OneForAll
	Preference   = model.Preference
	Preferences  = model.Preferences
	URF          = model.URF

	Client struct {
		*asol.Asol
		*Configuration
		*Session
	}
)

var (
	client = &Client{
		asol.NewAsol(),
		NewConfiguration(),
		NewSession(),
	}
)

func onOpen() {
	log.Println("The client is opened")
}

func onReady() {
	log.Println("The client is ready")

	var pages []Page

	request, _ := client.Get("/lol-perks/v1/pages")
	response, _ := client.RiotRequest(request)
	json.Unmarshal(response, &pages)

	for _, page := range pages {
		if client.pageName == page.Name {
			pageId := strconv.FormatFloat(page.Id, 'f', -1, 64)
			client.pageId = pageId
		}
	}

	preset := 5
	length := len(pages)

	if length > preset {
		for index, page := range pages {
			pageId := strconv.FormatFloat(page.Id, 'f', -1, 64)

			if index < (length - preset) {
				request, _ := client.Delete(
					fmt.Sprintf("/lol-perks/v1/pages/%s", pageId),
				)

				client.RiotRequest(request)
			}
		}
	}
}

func onLogin() {
	log.Println("The client is logged in")

	request, _ := client.Get("/lol-login/v1/session")
	response, _ := client.RiotRequest(request)

	var login Login
	json.Unmarshal(response, &login)

	summonerId := strconv.FormatFloat(login.SummonerId, 'f', -1, 64)
	accountId := strconv.FormatFloat(login.AccountId, 'f', -1, 64)
	username := login.Username

	client.summonerId = summonerId
	client.accountId = accountId
	client.username = username
}

func onLogout() {
	log.Println("The client is logged out")
}

func onClientClose() {
	log.Println("The client is closed")
}

func onWebsocketClose() {
	log.Println("The client's websocket closed")
}

func onReconnect() {
	log.Println("The client is reconnected")
}

func onWebsocketError(error error) {
	log.Println(error)
}

func onMatchFound(message []byte) {
	if client.autoAccept {
		var match MatchFound
		json.Unmarshal(message, &match)

		if match.Data.PlayerResponse == "None" && match.Data.Timer == 1.0 {
			time.Sleep(3000 * time.Millisecond)

			log.Println("Accepting match...")

			request, _ := client.Post("/lol-matchmaking/v1/ready-check/accept", nil)
			client.RiotRequest(request)
		}
	}
}

func onPhase(message []byte) {
	var phase Phase
	json.Unmarshal(message, &phase)

	if phase.Data != "ChampSelect" {
		client.mode = ""
	}

	if phase.Data == "None" {
		client.resetSession()
	}
}

func onSession(message []byte) {
	var championSelection ChampionSelection
	json.Unmarshal(message, &championSelection)

	phase := strings.ToLower(championSelection.Data.Timer.Phase)

	if phase == "planning" {
		return
	}

	if reflect.ValueOf(client.mode).IsZero() {
		request, _ := client.Get("/lol-gameflow/v1/session")
		data, _ := client.RiotRequest(request)

		var gameflow Gameflow
		json.Unmarshal(data, &gameflow)

		client.mode = strings.ToLower(gameflow.Map.GameMode)
	}

	if phase == "ban_pick" {
		for _, player := range championSelection.Data.MyTeam {
			summonerId := strconv.FormatFloat(player.SummonerId, 'f', -1, 64)

			if client.summonerId != summonerId {
				return
			}

			championId := strconv.FormatFloat(player.ChampionId, 'f', -1, 64)

			if client.championId == championId || championId == "0" {
				return
			}

			client.championId = championId

			// Delete the custom page
			if !reflect.ValueOf(client.pageId).IsZero() {
				request, _ := client.Delete(
					fmt.Sprintf("/lol-perks/v1/pages/%s", client.pageId),
				)

				client.RiotRequest(request)
			}

			if client.autoSpell {
				summonerspells()
			}

			if client.autoRune {
				runes()
			}

			page()
		}
	}

	if phase == "game_starting" {
		skillorder()
		itemset()
	}
}

func page() {
	var pages []Page

	request, _ := client.Get("/lol-perks/v1/pages")
	response, _ := client.RiotRequest(request)
	json.Unmarshal(response, &pages)

	for _, page := range pages {
		if client.pageName == page.Name {
			pageId := strconv.FormatFloat(page.Id, 'f', -1, 64)
			client.pageId = pageId
		}
	}
}

func summonerspells() {
	preferences := readPreferences()

	for id, champion := range preferences {
		if client.championId == id {
			client.championName = champion.Name

			var spell []string

			switch client.mode {
			case "aram":
				if client.gameType == "aram" {
					spell = append(
						spell,
						strings.ToLower(champion.ARAM.X),
						strings.ToLower(champion.ARAM.Y),
					)
				} else {
					spell = append(
						spell,
						strings.ToLower(champion.Classic.X),
						strings.ToLower(champion.Classic.Y),
					)
				}
			case "oneforall":
				spell = append(
					spell,
					strings.ToLower(champion.OneForAll.X),
					strings.ToLower(champion.OneForAll.Y),
				)
			case "urf":
				spell = append(
					spell,
					strings.ToLower(champion.URF.X),
					strings.ToLower(champion.URF.Y),
				)
			default:
				spell = append(
					spell,
					strings.ToLower(champion.Classic.X),
					strings.ToLower(champion.Classic.Y),
				)
			}

			var x, y float64

			if client.reverse {
				x = model.Spells[spell[1]]
				y = model.Spells[spell[0]]
			} else {
				x = model.Spells[spell[0]]
				y = model.Spells[spell[1]]
			}

			var payload = map[string]interface{}{
				"spell1Id": x,
				"spell2Id": y,
			}

			request, _ := client.Patch(
				"/lol-champ-select/v1/session/my-selection",
				payload,
			)

			client.RiotRequest(request)
		}
	}
}

func runes() {
	var url string

	switch client.mode {
	case "aram":
		if client.gameType == "aram" {
			url = fmt.Sprintf("%s/aram/ha/runes/%s", client.api, client.championId)
		} else {
			url = fmt.Sprintf("%s/aram/sr/runes/%s", client.api, client.championId)
		}
	case "oneforall":
		url = fmt.Sprintf("%s/ranked/runes/%s", client.api, client.championId)
	case "urf":
		url = fmt.Sprintf("%s/ranked/runes/%s", client.api, client.championId)
	default:
		url = fmt.Sprintf("%s/ranked/runes/%s", client.api, client.championId)
	}

	request, _ := client.Get(url)
	response, _ := client.WebRequest(request)

	var runes Runes
	json.Unmarshal(response, &runes)

	payload := map[string]interface{}{
		"autoModifiedSelections": [1]int{0},
		"current":                true,
		"id":                     0,
		"isActive":               true,
		"isDeletable":            true,
		"isEditable":             true,
		"isValid":                true,
		"lastModified":           0,
		"name":                   client.pageName,
		"order":                  0,
		"primaryStyleId":         runes.Primary,
		"selectedPerkIds":        runes.Runes,
		"subStyleId":             runes.Secondary,
	}

	request, _ = client.Post("/lol-perks/v1/pages", payload)
	_, err := client.RiotRequest(request)

	if err != nil {
		log.Println(err)
	}

	log.Println(
		"Setting rune page and summoner spells for",
		client.championName,
	)
}

func skillorder() {
	var url string

	switch client.mode {
	case "aram":
		if client.gameType == "aram" {
			url = fmt.Sprintf("%s/aram/ha/skills/%s", client.api, client.championId)
		} else {
			url = fmt.Sprintf("%s/aram/sr/skills/%s", client.api, client.championId)
		}
	case "oneforall":
		url = fmt.Sprintf("%s/ranked/skills/%s", client.api, client.championId)
	case "urf":
		url = fmt.Sprintf("%s/ranked/skills/%s", client.api, client.championId)
	default:
		url = fmt.Sprintf("%s/ranked/skills/%s", client.api, client.championId)
	}

	request, _ := client.Get(url)
	response, _ := client.WebRequest(request)

	var skills []string
	json.Unmarshal(response, &skills)

	log.Println(
		fmt.Sprintf(
			"Skill order for %v: %v",
			client.championName,
			strings.Trim(fmt.Sprint(skills), "[]"),
		),
	)
}

func itemset() {
	var url string

	switch client.mode {
	case "aram":
		if client.gameType == "aram" {
			url = fmt.Sprintf("%s/aram/ha/items/%s", client.api, client.championId)
		} else {
			url = fmt.Sprintf("%s/aram/sr/items/%s", client.api, client.championId)
		}
	case "oneforall":
		url = fmt.Sprintf("%s/event/sr/items/%s", client.api, client.championId)
	case "urf":
		url = fmt.Sprintf("%s/urf/sr/items/%s", client.api, client.championId)
	default:
		url = fmt.Sprintf("%s/ranked/sr/items/%s", client.api, client.championId)
	}

	request, _ := client.Get(url)
	response, _ := client.WebRequest(request)

	var itemset interface{}
	json.Unmarshal(response, &itemset)

	var payload = map[string]interface{}{
		"accountId": client.accountId,
	}

	for k, v := range itemset.(map[string]interface{}) {
		payload[k] = v
	}

	url = fmt.Sprintf("/lol-item-sets/v1/item-sets/%s/sets", client.accountId)
	request, _ = client.Put(url, payload)
	client.RiotRequest(request)
}

func logger() func() {
	err := os.MkdirAll("log", os.ModePerm)

	if err != nil {
		fmt.Println(err)
	}

	file, err := os.OpenFile(
		"log/senna.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	stdout := os.Stdout
	multiwriter := io.MultiWriter(stdout, file)
	read, write, _ := os.Pipe()

	os.Stdout = write
	os.Stderr = write

	log.SetOutput(multiwriter)
	exit := make(chan bool)

	go func() {
		_, _ = io.Copy(multiwriter, read)
		exit <- true
	}()

	return func() {
		_ = write.Close()

		<-exit

		_ = file.Close()
	}
}

func main() {
	multiwriter := logger()
	defer multiwriter()

	err := checkForUpdates()

	if err != nil {
		log.Println(err)
	}

	client.OnOpen(onOpen)
	client.OnReady(onReady)
	client.OnLogin(onLogin)
	client.OnLogout(onLogout)
	client.OnClientClose(onClientClose)
	client.OnWebsocketClose(onWebsocketClose)
	client.OnReconnect(onReconnect)
	client.OnWebsocketError(onWebsocketError)

	client.OnMessage(
		"/lol-matchmaking/v1/ready-check",
		"Update",
		onMatchFound,
	)

	client.OnMessage(
		"/lol-gameflow/v1/gameflow-phase",
		"Update",
		onPhase,
	)

	client.OnMessage(
		"/lol-champ-select/v1/session",
		"Update",
		onSession,
	)

	client.Start()
}

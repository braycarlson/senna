package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/braycarlson/asol"
	"github.com/braycarlson/senna/model"
	"github.com/go-ini/ini"
)

type (
	Asol = asol.Asol

	Client struct {
		*Asol
		*Configuration
		*Session
	}

	Configuration struct {
		api        string
		autoAccept bool
		autoRune   bool
		autoSpell  bool
		gameType   string
		pageName   string
		reverse    bool
		version    string
	}

	Session struct {
		accountId    string
		championId   string
		championName string
		mode         string
		pageId       string
		summonerId   string
		username     string
	}
)

func NewSession() *Session {
	return &Session{}
}

func NewConfiguration() *Configuration {
	_, err := os.OpenFile(
		"config.ini",
		os.O_RDWR|os.O_CREATE|os.O_EXCL,
		0666,
	)

	file, _ := ini.Load("config.ini")

	if err == nil {
		file.NewSection("senna")
		file.Section("senna").NewKey("api", "https://localhost.com:5000")
		file.Section("senna").NewKey("autoaccept", "true")
		file.Section("senna").NewKey("autorune", "true")
		file.Section("senna").NewKey("autospell", "true")
		file.Section("senna").NewKey("gametype", "aram")
		file.Section("senna").NewKey("page", "senna")
		file.Section("senna").NewKey("reverse", "false")
		file.Section("senna").NewKey("version", "0.0.1")

		err = file.SaveTo("config.ini")
	}

	section := file.Section("senna")
	api := section.Key("api").String()
	autoAccept, _ := section.Key("autoaccept").Bool()
	autoRune, _ := section.Key("autorune").Bool()
	autoSpell, _ := section.Key("autospell").Bool()
	gameType := section.Key("gametype").String()
	pageName := section.Key("page").String()
	reverse, _ := section.Key("reverse").Bool()
	version := section.Key("version").String()

	return &Configuration{
		api:        api,
		autoAccept: autoAccept,
		autoRune:   autoRune,
		autoSpell:  autoSpell,
		gameType:   gameType,
		pageName:   pageName,
		reverse:    reverse,
		version:    version,
	}
}

var (
	client = &Client{
		asol.NewAsol(),
		NewConfiguration(),
		NewSession(),
	}
)

const (
	ddragon = "https://ddragon.leagueoflegends.com"
)

func onOpen(asol *Asol) {
	fmt.Println("The client is opened")
}

func onReady(asol *Asol) {
	fmt.Println("The client is ready")

	var pages []model.Page

	request, _ := asol.Get("/lol-perks/v1/pages")
	response, _ := asol.RiotRequest(request)
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
				request, _ := asol.Delete(
					fmt.Sprintf("/lol-perks/v1/pages/%s", pageId),
				)

				asol.RiotRequest(request)
			}
		}
	}
}

func onLogin(asol *Asol) {
	fmt.Println("The client is logged in")

	request, _ := asol.Get("/lol-login/v1/session")
	response, _ := asol.RiotRequest(request)

	var login model.Login
	json.Unmarshal(response, &login)

	summonerId := strconv.FormatFloat(login.SummonerId, 'f', -1, 64)
	accountId := strconv.FormatFloat(login.AccountId, 'f', -1, 64)
	username := login.Username

	client.summonerId = summonerId
	client.accountId = accountId
	client.username = username
}

func onLogout(asol *Asol) {
	fmt.Println("The client is logged out")
}

func onClientClose(asol *Asol) {
	fmt.Println("The client is closed")
}

func onWebsocketClose(asol *Asol) {
	fmt.Println("The client's websocket closed")
}

func onReconnect(asol *Asol) {
	fmt.Println("The client is reconnected")
}

func onWebsocketError(error error) {
	fmt.Println(error)
}

func onMatchFound(asol *Asol, message []byte) {
	var match model.MatchFound
	json.Unmarshal(message, &match)

	if match.Data.PlayerResponse == "None" && match.Data.Timer == 1.0 {
		time.Sleep(3000 * time.Millisecond)

		fmt.Println("Accepting match...")

		request, _ := asol.Post("/lol-matchmaking/v1/ready-check/accept", nil)
		asol.RiotRequest(request)
	}
}

func onPhase(asol *Asol, message []byte) {
	var phase model.Phase
	json.Unmarshal(message, &phase)

	if phase.Data != "ChampSelect" {
		client.mode = ""

		if phase.Data != "GameStart" && phase.Data != "InProgress" {
			client.championId = ""
		}
	}
}

func onSession(asol *Asol, message []byte) {
	var championSelection model.ChampionSelection
	json.Unmarshal(message, &championSelection)

	phase := strings.ToLower(championSelection.Data.Timer.Phase)

	if phase == "planning" {
		return
	}

	if reflect.ValueOf(client.mode).IsZero() {
		request, _ := asol.Get("/lol-gameflow/v1/session")
		data, _ := asol.RiotRequest(request)

		var gameflow model.Gameflow
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

			var url string

			// Runes
			switch client.mode {
			case "aram":
				if client.gameType == "aram" {
					url = fmt.Sprintf("%s/aram/ha/runes/%s", client.api, championId)
				} else {
					url = fmt.Sprintf("%s/aram/sr/runes/%s", client.api, championId)
				}
			case "oneforall":
				url = fmt.Sprintf("%s/ranked/runes/%s", client.api, championId)
			case "urf":
				url = fmt.Sprintf("%s/ranked/runes/%s", client.api, championId)
			default:
				url = fmt.Sprintf("%s/ranked/runes/%s", client.api, championId)
			}

			request, _ := client.Get(url)
			response, _ := client.WebRequest(request)

			var runes model.Runes
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

			request, _ = asol.Post("/lol-perks/v1/pages", payload)
			_, err := asol.RiotRequest(request)

			if err != nil {
				fmt.Println(err)
			}

			// Pages
			var pages []model.Page

			request, _ = asol.Get("/lol-perks/v1/pages")
			response, _ = asol.RiotRequest(request)
			json.Unmarshal(response, &pages)

			for _, page := range pages {
				if client.pageName == page.Name {
					pageId := strconv.FormatFloat(page.Id, 'f', -1, 64)
					client.pageId = pageId
				}
			}
		}
	}

	if phase == "game_starting" {
		// Skill Order

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

		fmt.Println(
			strings.Trim(fmt.Sprint(skills), "[]"),
		)

		// Itemset

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

		request, _ = client.Get(url)
		response, _ = client.WebRequest(request)

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
}

func main() {
	client.OnOpen(onOpen)
	client.OnReady(onReady)
	client.OnLogin(onLogin)
	client.OnLogout(onLogout)
	client.OnClientClose(onClientClose)
	client.OnWebsocketClose(onWebsocketClose)
	client.OnReconnect(onReconnect)
	client.OnWebsocketError(onWebsocketError)

	if client.autoAccept {
		client.OnMessage(
			"/lol-matchmaking/v1/ready-check",
			"Update",
			onMatchFound,
		)
	}

	if client.autoRune {
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
	}

	client.Start()
}

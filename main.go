package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/braycarlson/asol"
	"github.com/braycarlson/senna/model"
	"github.com/go-ini/ini"
)

type (
	Asol    = asol.Asol
	Message = asol.Message

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
	file, err := ini.Load("config.ini")

	if err != nil {
		fmt.Printf(".ini file could not be read: %v", err)
		return nil
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
	preset  = 5
)

func onOpen(asol *Asol) {
	fmt.Println("The client is opened")
}

func onReady(asol *Asol) {
	fmt.Println("The client is ready")

	var pages []model.Page

	request, _ := asol.NewGetRequest("/lol-perks/v1/pages")
	response, _ := asol.RawRiotRequest(request)
	json.Unmarshal(response, &pages)

	for _, page := range pages {
		if client.pageName == page.Name {
			pageId := strconv.FormatFloat(page.Id, 'f', -1, 64)
			client.pageId = pageId
		}
	}

	var length int = len(pages)

	if length > preset {
		for index, page := range pages {
			pageId := strconv.FormatFloat(page.Id, 'f', -1, 64)

			if index < (length - preset) {
				request, _ := asol.NewDeleteRequest(
					fmt.Sprintf("/lol-perks/v1/pages/%s", pageId),
				)

				asol.RiotRequest(request)
			}
		}
	}
}

func onLogin(asol *Asol) {
	fmt.Println("The client is logged in")

	request, _ := asol.NewGetRequest("/lol-login/v1/session")
	response, _ := asol.RawRiotRequest(request)

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

func onRequest(uri string, status int) {
	fmt.Println(
		fmt.Sprintf("%d: %s", status, uri),
	)
}

func onRequestError(uri string, status int) {
	fmt.Println(
		fmt.Sprintf("%d: %s", status, uri),
	)
}

func onWebsocketError(error error) {
	fmt.Println(error)
}

func onMatchFound(asol *Asol, message *Message) {
	bytes, _ := json.Marshal(message.Data)

	var match model.MatchFound
	json.Unmarshal(bytes, &match)

	if match.Data.PlayerResponse == "None" && match.Data.Timer == 1.0 {
		time.Sleep(3000 * time.Millisecond)

		fmt.Println("Accepting match...")

		request, _ := asol.NewPostRequest("/lol-matchmaking/v1/ready-check/accept", nil)
		asol.RiotRequest(request)
	}
}

func onPhase(asol *Asol, message *Message) {
	bytes, _ := json.Marshal(message.Data)

	var phase model.Phase
	json.Unmarshal(bytes, &phase)

	if phase.Data != "ChampSelect" {
		client.mode = ""

		if phase.Data != "GameStart" && phase.Data != "InProgress" {
			client.championId = ""
		}
	}
}

func onSession(asol *Asol, message *Message) {
	bytes, _ := json.Marshal(message.Data)
	var championSelection model.ChampionSelection
	json.Unmarshal(bytes, &championSelection)

	phase := strings.ToLower(championSelection.Data.Timer.Phase)

	if phase == "planning" {
		return
	}

	if reflect.ValueOf(client.mode).IsZero() {
		request, _ := asol.NewGetRequest("/lol-gameflow/v1/session")
		data, _ := asol.RawRiotRequest(request)

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

			if client.championId == championId || player.ChampionId == 0 {
				return
			}

			client.championId = championId

			// Delete the custom page
			if !reflect.ValueOf(client.pageId).IsZero() {
				request, _ := client.NewDeleteRequest(
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

			request, _ := client.NewGetRequest(url)
			response, _ := client.RawWebRequest(request)

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

			request, _ = asol.NewPostRequest("/lol-perks/v1/pages", payload)
			_, err := asol.RawRiotRequest(request)

			if err != nil {
				fmt.Println(err)
			}

			// Pages
			var pages []model.Page

			request, _ = asol.NewGetRequest("/lol-perks/v1/pages")
			data, _ := asol.RawRiotRequest(request)
			json.Unmarshal(data, &pages)

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

		request, _ := client.NewGetRequest(url)
		response, _ := client.WebRequest(request)

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

		request, _ = client.NewGetRequest(url)
		response, _ = client.WebRequest(request)

		var payload = map[string]interface{}{
			"accountId": client.accountId,
		}

		for k, v := range response.(map[string]interface{}) {
			payload[k] = v
		}

		url = fmt.Sprintf("/lol-item-sets/v1/item-sets/%s/sets", client.accountId)
		request, _ = client.NewPutRequest(url, payload)
		client.RawRiotRequest(request)
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
	client.OnRequest(onRequest)
	client.OnRequestError(onRequestError)
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

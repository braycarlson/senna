package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/braycarlson/asol"
	"github.com/go-ini/ini"
)

type (
	Asol = asol.Asol

	Client struct {
		*Asol
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

const (
	ddragon = "https://ddragon.leagueoflegends.com"
)

func onOpen(asol *Asol) {
	log.Println("The client is opened")

	date := time.Now().Format("01-02-2006")

	if date == client.date {
		return
	}

	log.Println("Checking for updates...")

	file, _ := ini.Load("config.ini")
	file.Section("senna").NewKey("date", date)
	err := file.SaveTo("config.ini")

	log.Println(err)

	updatePreferences()
}

func onReady(asol *Asol) {
	log.Println("The client is ready")

	var pages []Page

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
	log.Println("The client is logged in")

	request, _ := asol.Get("/lol-login/v1/session")
	response, _ := asol.RiotRequest(request)

	var login Login
	json.Unmarshal(response, &login)

	summonerId := strconv.FormatFloat(login.SummonerId, 'f', -1, 64)
	accountId := strconv.FormatFloat(login.AccountId, 'f', -1, 64)
	username := login.Username

	client.summonerId = summonerId
	client.accountId = accountId
	client.username = username
}

func onLogout(asol *Asol) {
	log.Println("The client is logged out")
}

func onClientClose(asol *Asol) {
	log.Println("The client is closed")
}

func onWebsocketClose(asol *Asol) {
	log.Println("The client's websocket closed")
}

func onReconnect(asol *Asol) {
	log.Println("The client is reconnected")
}

func onWebsocketError(error error) {
	log.Println(error)
}

func onMatchFound(asol *Asol, message []byte) {
	var match MatchFound
	json.Unmarshal(message, &match)

	if match.Data.PlayerResponse == "None" && match.Data.Timer == 1.0 {
		time.Sleep(3000 * time.Millisecond)

		log.Println("Accepting match...")

		request, _ := asol.Post("/lol-matchmaking/v1/ready-check/accept", nil)
		asol.RiotRequest(request)
	}
}

func onPhase(asol *Asol, message []byte) {
	var phase Phase
	json.Unmarshal(message, &phase)

	if phase.Data != "ChampSelect" {
		client.mode = ""

		if phase.Data != "GameStart" &&
			phase.Data != "InProgress" {
			client.championId = ""
		}
	}
}

func onSession(asol *Asol, message []byte) {
	var championSelection ChampionSelection
	json.Unmarshal(message, &championSelection)

	phase := strings.ToLower(championSelection.Data.Timer.Phase)

	if phase == "planning" {
		return
	}

	if reflect.ValueOf(client.mode).IsZero() {
		request, _ := asol.Get("/lol-gameflow/v1/session")
		data, _ := asol.RiotRequest(request)

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

			// Summoner Spells
			preferences := getPreferences()

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
						x = Spells[spell[1]]
						y = Spells[spell[0]]
					} else {
						x = Spells[spell[0]]
						y = Spells[spell[1]]
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

			request, _ = asol.Post("/lol-perks/v1/pages", payload)
			_, err := asol.RiotRequest(request)

			if err != nil {
				log.Println(err)
			}

			log.Println(
				"Setting rune page and spells for",
				client.championName,
			)

			// Pages
			var pages []Page

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

		log.Println(
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

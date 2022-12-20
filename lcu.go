package main

import (
	"encoding/json"
	"fyne.io/fyne/v2/dialog"
	"strconv"
	"strings"

	"github.com/braycarlson/senna/collection"
	"github.com/braycarlson/senna/model"
)

func (senna *Senna) getChampionName(id string) string {
	file := senna.cache.GetBig(
		[]byte{},
		[]byte("champion.json"),
	)

	var champion model.Champions
	json.Unmarshal(file, &champion)

	if champion, ok := champion[id]; ok {
		return champion.Name
	}

	return ""
}

func (senna *Senna) onSearch() {
	senna.logger("senna is searching")
}

func (senna *Senna) onOpen() {
	senna.logger("senna is opened")
}

func (senna *Senna) onReady() {
	senna.logger("senna is ready")

	if senna.settings.AutoRune {
		var owned model.OwnedPages

		senna.client.SetWebsocket()
		request, _ := senna.client.Get("/lol-perks/v1/inventory")
		response, _ := senna.client.Request(request)
		json.Unmarshal(response, &owned)

		var pages []model.Page

		request, _ = senna.client.Get("/lol-perks/v1/pages")
		response, _ = senna.client.Request(request)
		json.Unmarshal(response, &pages)

		for _, page := range pages {
			if senna.settings.PageName == page.Name {
				pageId := strconv.FormatFloat(page.Id, 'f', -1, 64)

				var builder strings.Builder
				builder.WriteString("/lol-perks/v1/pages/")
				builder.WriteString(pageId)

				url := builder.String()

				request, _ := senna.client.Delete(url)
				senna.client.Request(request)
			}
		}

		request, _ = senna.client.Get("/lol-perks/v1/pages")
		response, _ = senna.client.Request(request)
		json.Unmarshal(response, &pages)

		if len(pages)-5 == owned.Count {
			dialog.ShowInformation(
				"Warning",
				"Please delete a runepage",
				senna.ui.Window,
			)
		}
	}
}

func (senna *Senna) onLogin() {
	senna.logger("senna is logged in")

	senna.client.SetWebsocket()
	request, _ := senna.client.Get("/lol-login/v1/session")
	response, _ := senna.client.Request(request)

	var login model.Login
	json.Unmarshal(response, &login)

	summonerId := strconv.FormatFloat(
		login.SummonerId,
		'f',
		-1,
		64,
	)

	accountId := strconv.FormatFloat(
		login.AccountId,
		'f',
		-1,
		64,
	)

	username := login.Username

	senna.cache.Set(
		[]byte("summonerId"),
		[]byte(summonerId),
	)

	senna.cache.Set(
		[]byte("accountId"),
		[]byte(accountId),
	)

	senna.cache.Set(
		[]byte("username"),
		[]byte(username),
	)

	senna.ui.Home.Username.Set(username)
}

func (senna *Senna) onProcessClose(message []byte) {
	var process model.ProcessControl
	json.Unmarshal(message, &process)

	if process.Data["status"] == "Stopping" {
		senna.logger("senna was disconnected")

		if senna.ui.Home.State.Text == "Stop" {
			senna.ui.Home.State.Importance = 1
			senna.ui.Home.State.Text = "Start"
			senna.ui.Home.State.Refresh()
			senna.ui.Clear()

			go senna.Stop()
		}
	}
}

func (senna *Senna) onProcessError(error error) {
	senna.logger("senna was unable to connect")

	if senna.ui.Home.State.Text == "Stop" {
		senna.ui.Home.State.Importance = 1
		senna.ui.Home.State.Text = "Start"
		senna.ui.Home.State.Refresh()
		senna.ui.Clear()

		go senna.Stop()
	}
}

func (senna *Senna) onSearchError(error error) {
	senna.logger("senna was cancelled")

	if senna.ui.Home.State.Text == "Stop" {
		senna.ui.Home.State.Importance = 1
		senna.ui.Home.State.Text = "Start"
		senna.ui.Home.State.Refresh()
		senna.ui.Clear()

		go senna.Stop()
	}
}

func (senna *Senna) onWebsocketClose() {
	senna.logger("senna's websocket was closed")
}

func (senna *Senna) onWebsocketError(error error) {
	senna.logger("senna's websocket was disrupted")
}

func (senna *Senna) onMatchFound(message []byte) {
	if senna.settings.AutoAccept {
		var match model.MatchFound
		json.Unmarshal(message, &match)

		if match.Data.Timer == 1.0 {
			senna.logger("senna found a match")
		}

		if match.Data.PlayerResponse == "None" && match.Data.Timer == 3.0 {
			senna.logger("senna is accepting the match")

			request, _ := senna.client.Post(
				"/lol-matchmaking/v1/ready-check/accept",
				nil,
			)

			senna.client.SetWebsocket()
			senna.client.Request(request)
		}
	}
}

func (senna *Senna) onPhase(message []byte) {
	var phase model.Phase
	json.Unmarshal(message, &phase)

	if phase.Data != "ChampSelect" {
		senna.cache.Del(
			[]byte("gameflow"),
		)
	}

	if phase.Data == "None" {
		senna.cache.Del(
			[]byte("championId"),
		)

		senna.cache.Del(
			[]byte("championName"),
		)

		senna.cache.Del(
			[]byte("gameflow"),
		)
	}
}

func (senna *Senna) onSession(message []byte) {
	var championSelection model.ChampionSelection
	json.Unmarshal(message, &championSelection)

	phase := strings.ToLower(championSelection.Data.Timer.Phase)

	if phase == "planning" {
		return
	}

	gameflow := senna.cache.Get(
		[]byte{},
		[]byte("gameflow"),
	)

	if len(gameflow) == 0 {
		senna.client.SetWebsocket()
		request, _ := senna.client.Get("/lol-gameflow/v1/session")
		data, _ := senna.client.Request(request)

		var gameflow model.Gameflow
		json.Unmarshal(data, &gameflow)

		senna.cache.Set(
			[]byte("gameflow"),
			[]byte(
				strings.ToLower(gameflow.Map.GameMode),
			),
		)
	}

	if phase == "ban_pick" || phase == "finalization" {
		for _, player := range championSelection.Data.MyTeam {
			summonerId := strconv.FormatFloat(player.SummonerId, 'f', -1, 64)

			sid := string(
				senna.cache.Get(
					[]byte{},
					[]byte("summonerId"),
				),
			)

			if sid == summonerId {
				championId := strconv.FormatFloat(player.ChampionId, 'f', -1, 64)

				cid := string(
					senna.cache.Get(
						[]byte{},
						[]byte("championId"),
					),
				)

				if cid == championId || championId == "0" {
					return
				}

				senna.cache.Set(
					[]byte("championId"),
					[]byte(championId),
				)

				championName := senna.getChampionName(championId)

				senna.cache.Set(
					[]byte("championName"),
					[]byte(championName),
				)

				var payload []byte

				if senna.settings.AutoSpell {
					spell := &collection.SummonerSpell{
						Cache:       senna.cache,
						Client:      senna.client,
						Preferences: senna.preferences,
						Settings:    senna.settings,
					}

					payload = spell.Get(
						&collection.Local{
							collection.Parameter{
								Asset:  "spell",
								Mode:   senna.settings.Mode,
								Region: senna.settings.Region,
							},
						},
					)

					spell.Set(payload)

					var status string = spell.Status()
					senna.ui.Event.History.Append(status)
				}

				if senna.settings.AutoRune {
					pageId := senna.cache.Get(
						[]byte{},
						[]byte("pageId"),
					)

					runepage := &collection.Runepage{
						Cache:    senna.cache,
						Client:   senna.client,
						Settings: senna.settings,
					}

					payload = runepage.Get(
						&collection.Local{
							collection.Parameter{
								Asset:  "runepage",
								Mode:   senna.settings.Mode,
								Region: senna.settings.Region,
							},
						},
					)

					runepage.Delete(pageId)
					runepage.Set(payload)
					runepage.Update()

					var status string = runepage.Status()
					senna.ui.Event.History.Append(status)
				}

				var status string

				itemset := &collection.Itemset{
					Cache:    senna.cache,
					Client:   senna.client,
					Settings: senna.settings,
				}

				payload = itemset.Get(
					&collection.Local{
						collection.Parameter{
							Asset:  "itemset",
							Mode:   senna.settings.Mode,
							Region: senna.settings.Region,
						},
					},
				)

				itemset.Set(payload)

				status = itemset.Status()
				senna.ui.Event.History.Append(status)

				skillorder := &collection.SkillOrder{
					Cache:    senna.cache,
					Client:   senna.client,
					Settings: senna.settings,
				}

				payload = skillorder.Get(
					&collection.Local{
						collection.Parameter{
							Asset:  "skillorder",
							Mode:   senna.settings.Mode,
							Region: senna.settings.Region,
						},
					},
				)

				status = skillorder.Status() + skillorder.Process(payload)
				senna.ui.Event.History.Append(status)
			}
		}
	}
}

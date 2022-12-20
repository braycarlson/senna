package collection

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/braycarlson/asol/request"
	"github.com/braycarlson/senna/model"
	"github.com/braycarlson/senna/preferences"
	"github.com/braycarlson/senna/settings"
)

type (
	SummonerSpell struct {
		Cache       *fastcache.Cache
		Client      *request.HTTPClient
		Preferences *preferences.Preferences
		Settings    *settings.Settings
	}
)

func (spell *SummonerSpell) Set(payload []byte) error {
	spell.Client.SetWebsocket()

	request, _ := spell.Client.Patch(
		"/lol-champ-select/v1/session/my-selection",
		payload,
	)

	_, err := spell.Client.Request(request)
	return err
}

func (spell *SummonerSpell) Status() string {
	championName := string(
		spell.Cache.Get(
			[]byte{},
			[]byte("championName"),
		),
	)

	var builder strings.Builder
	builder.WriteString("Setting summoner spells for ")
	builder.WriteString(championName)

	return builder.String()
}

func (spell *SummonerSpell) Get(visitor Visitor) []byte {
	return visitor.getSummonerSpell(spell)
}

func (local *Local) getSummonerSpell(spell *SummonerSpell) []byte {
	championId := string(
		spell.Cache.Get(
			[]byte{},
			[]byte("championId"),
		),
	)

	gameflow := string(
		spell.Cache.Get(
			[]byte{},
			[]byte("gameflow"),
		),
	)

	preference := spell.Preferences.Preference(championId)

	var spells []string

	switch gameflow {
	case "aram":
		spells = append(
			spells,
			strings.ToLower(preference.ARAM.X),
			strings.ToLower(preference.ARAM.Y),
		)
	case "classic":
		spells = append(
			spells,
			strings.ToLower(preference.Classic.X),
			strings.ToLower(preference.Classic.Y),
		)
	case "oneforall":
		spells = append(
			spells,
			strings.ToLower(preference.OneForAll.X),
			strings.ToLower(preference.OneForAll.Y),
		)
	case "urf":
		spells = append(
			spells,
			strings.ToLower(preference.URF.X),
			strings.ToLower(preference.URF.Y),
		)
	default:
		spells = append(
			spells,
			strings.ToLower(preference.Classic.X),
			strings.ToLower(preference.Classic.Y),
		)
	}

	var x, y float64

	if spell.Settings.Reverse {
		x = model.Spells[spells[1]]
		y = model.Spells[spells[0]]
	} else {
		x = model.Spells[spells[0]]
		y = model.Spells[spells[1]]
	}

	var data = map[string]interface{}{
		"spell1Id": x,
		"spell2Id": y,
	}

	payload, _ := json.Marshal(data)
	return payload
}

func (web *Web) getSummonerSpell(spell *SummonerSpell) []byte {
	fmt.Println("Get spell from web")
	return []byte{}
}

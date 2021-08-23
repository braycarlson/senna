package model

import (
	"encoding/json"
)

var (
	Spells = map[string]float64{
		"cleanse":     1,
		"exhaust":     3,
		"flash":       4,
		"ghost":       6,
		"heal":        7,
		"ignite":      14,
		"barrier":     21,
		"clarity":     13,
		"snowball":    32,
		"smite":       11,
		"teleport":    12,
		"to the king": 30,
		"poro toss":   31,
	}
)

type (
	ARAM struct {
		X string `json:"x"`
		Y string `json:"y"`
	}

	Champion struct {
		Type    string
		Format  string
		Version string
		Data    map[string]ChampionData
	}

	ChampionData struct {
		Version string
		Id      string
		Key     string
		Name    string
		Title   string
		Blurb   string

		Info struct {
			Attack     float64
			Defense    float64
			Magic      float64
			Difficulty float64
		}

		Image struct {
			Full   string
			Sprite string
			Group  string
			X      float64
			Y      float64
			W      float64
			H      float64
		}

		Tags    []string
		Partype string

		Stats struct {
			HP                   float64
			HPPerLevel           float64
			MP                   float64
			MPPerLevel           float64
			Movespeed            float64
			Armor                float64
			ArmorPerLevel        float64
			Spellblock           float64
			SpellblockPerLevel   float64
			AttackRange          float64
			HPRegen              float64
			HPRegenPerLevel      float64
			MPRegen              float64
			MPRegenPerLevel      float64
			Crit                 float64
			CritPerLevel         float64
			AttackDamage         float64
			AttackDamagePerLevel float64
			AttackSpeedPerLevel  float64
			AttackSpeed          float64
		}
	}

	Classic struct {
		X string `json:"x"`
		Y string `json:"y"`
	}

	OneForAll struct {
		X string `json:"x"`
		Y string `json:"y"`
	}

	Preference struct {
		Name      string    `json:"name"`
		OPGG      string    `json:"opgg"`
		ARAM      ARAM      `json:"aram"`
		Classic   Classic   `json:"classic"`
		OneForAll OneForAll `json:"oneforall"`
		URF       URF       `json:"urf"`
	}

	Preferences struct {
		Champion map[string]Preference
	}

	URF struct {
		X string `json:"x"`
		Y string `json:"y"`
	}
)

func (preferences Preferences) MarshalJSON() ([]byte, error) {
	champion, err := json.Marshal(preferences.Champion)
	return []byte(string(champion)), err
}

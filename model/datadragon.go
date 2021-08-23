package model

type (
	/*
		Method: GET
		URI: 	https://ddragon.leagueoflegends.com/realms/na.json
	*/
	Realms struct {
		CDN string
		CSS string
		DD  string
		L   string
		LG  string

		N struct {
			Champion    string
			Item        string
			Language    string
			Map         string
			Mastery     string
			ProfileIcon string
			Rune        string
			Sticker     string
			Summoner    string
		}

		ProfileIconMax float64
		Store          string
		V              string
	}
)

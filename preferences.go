package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func realms() (Realms, error) {
	url := fmt.Sprintf("%s/realms/na.json", ddragon)
	request, _ := client.Get(url)
	data, _ := client.WebRequest(request)

	var realms Realms
	json.Unmarshal(data, &realms)

	return realms, nil
}

func champions() (Champion, error) {
	realms, _ := realms()

	url := fmt.Sprintf("%s/cdn/%s/data/en_US/champion.json", ddragon, realms.N.Champion)
	request, _ := client.Get(url)
	data, _ := client.WebRequest(request)

	var champions Champion
	json.Unmarshal(data, &champions)

	return champions, nil
}

func createPreferences() {
	champions, err := champions()

	if err != nil {
		return
	}

	champion := make(map[string]Preference)

	for _, data := range champions.Data {
		preference := Preference{
			Name: data.Name,
			OPGG: strings.ToLower(data.Id),

			ARAM: ARAM{
				X: "Flash",
				Y: "Snowball",
			},
			Classic: Classic{
				X: "Flash",
				Y: "Ghost",
			},
			OneForAll: OneForAll{
				X: "Flash",
				Y: "Teleport",
			},
			URF: URF{
				X: "Flash",
				Y: "Ghost",
			},
		}

		champion[data.Key] = preference
	}

	preferences := &Preferences{
		Champion: champion,
	}

	_, err = os.OpenFile(
		"preferences.json",
		os.O_CREATE|os.O_EXCL,
		0755,
	)

	json, _ := json.MarshalIndent(preferences, "", "\t")
	_ = ioutil.WriteFile("preferences.json", json, 0644)
}

func getPreferences() map[string]Preference {
	file, err := os.OpenFile(
		"preferences.json",
		os.O_RDONLY,
		0,
	)

	if err != nil {
		log.Println(err)
	}

	fi, err := file.Stat()

	if err != nil {
		log.Println(err)
	}

	if fi.Size() == 0 {
		createPreferences()
	}

	defer file.Close()

	data, _ := ioutil.ReadAll(file)
	var preferences map[string]Preference
	json.Unmarshal(data, &preferences)

	return preferences
}

func difference(x, y []string) []string {
	temporary := make(
		map[string]struct{},
		len(y),
	)

	for _, v := range y {
		temporary[v] = struct{}{}
	}

	var difference []string

	for _, v := range x {
		if _, found := temporary[v]; !found {
			difference = append(difference, v)
		}
	}

	return difference
}

func isUpToDate() ([]string, error) {
	champions, _ := champions()
	preferences := getPreferences()

	var cid, pid []string

	for _, champion := range champions.Data {
		cid = append(cid, champion.Key)
	}

	for key, _ := range preferences {
		pid = append(pid, key)
	}

	return difference(cid, pid), nil
}

func updatePreferences() {
	id, _ := isUpToDate()

	if len(id) == 0 {
		return
	}

	champions, _ := champions()
	missing := make(map[string]Preference)

	for _, champion := range champions.Data {
		for _, cid := range id {
			if cid == champion.Key {
				preference := Preference{
					Name: champion.Name,
					OPGG: strings.ToLower(champion.Id),

					ARAM: ARAM{
						X: "Flash",
						Y: "Snowball",
					},
					Classic: Classic{
						X: "Flash",
						Y: "Ghost",
					},
					OneForAll: OneForAll{
						X: "Flash",
						Y: "Teleport",
					},
					URF: URF{
						X: "Flash",
						Y: "Ghost",
					},
				}

				missing[champion.Key] = preference
			}
		}
	}

	preferences := &Preferences{
		Champion: missing,
	}

	p := getPreferences()

	for k, v := range preferences.Champion {
		p[k] = v
	}

	if err := os.Truncate("preferences.json", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	json, _ := json.MarshalIndent(p, "", "\t")
	_ = ioutil.WriteFile("preferences.json", json, 0644)
}

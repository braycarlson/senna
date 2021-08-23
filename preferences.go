package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-ini/ini"
)

func NewDefaultPreference(champion ChampionData) Preference {
	return Preference{
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
}

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
		preference := NewDefaultPreference(data)
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

func readPreferences() map[string]Preference {
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

func isDifference() ([]string, error) {
	champions, _ := champions()
	preferences := readPreferences()

	var cid, pid []string

	for _, champion := range champions.Data {
		cid = append(cid, champion.Key)
	}

	for key, _ := range preferences {
		pid = append(pid, key)
	}

	id := difference(cid, pid)
	return id, nil
}

func updatePreferences() {
	id, _ := isDifference()

	if len(id) == 0 {
		return
	}

	champions, _ := champions()

	champion := make(map[string]Preference)

	for _, data := range champions.Data {
		for _, cid := range id {
			if cid == data.Key {
				preference := NewDefaultPreference(data)
				champion[data.Key] = preference
			}
		}
	}

	preferences := &Preferences{
		Champion: champion,
	}

	file := readPreferences()

	for k, v := range preferences.Champion {
		file[k] = v
	}

	if err := os.Truncate("preferences.json", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	json, _ := json.MarshalIndent(file, "", "\t")
	_ = ioutil.WriteFile("preferences.json", json, 0644)
}

func checkForUpdates() error {
	date := time.Now().Format("01-02-2006")

	if date == client.date {
		return nil
	}

	file, err := ini.Load("config.ini")
	file.Section("senna").NewKey("date", date)
	err = file.SaveTo("config.ini")

	if err != nil {
		return err
	}

	updatePreferences()

	return nil
}

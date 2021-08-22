package main

import (
	"os"
	"time"

	"github.com/go-ini/ini"
)

type (
	Configuration struct {
		api        string
		autoAccept bool
		autoRune   bool
		autoSpell  bool
		gameType   string
		pageName   string
		reverse    bool
		version    string
		date       string
	}
)

func NewConfiguration() *Configuration {
	_, err := os.OpenFile(
		"config.ini",
		os.O_RDWR|os.O_CREATE|os.O_EXCL,
		0666,
	)

	file, _ := ini.Load("config.ini")

	if err == nil {
		date := time.Now().Format("01-02-2006")

		file.NewSection("senna")
		file.Section("senna").NewKey("api", "https://localhost.com:5000")
		file.Section("senna").NewKey("autoaccept", "true")
		file.Section("senna").NewKey("autorune", "true")
		file.Section("senna").NewKey("autospell", "true")
		file.Section("senna").NewKey("gametype", "aram")
		file.Section("senna").NewKey("page", "senna")
		file.Section("senna").NewKey("reverse", "false")
		file.Section("senna").NewKey("version", "0.0.1")
		file.Section("senna").NewKey("date", date)

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
	date := section.Key("date").String()

	return &Configuration{
		api:        api,
		autoAccept: autoAccept,
		autoRune:   autoRune,
		autoSpell:  autoSpell,
		gameType:   gameType,
		pageName:   pageName,
		reverse:    reverse,
		version:    version,
		date:       date,
	}
}

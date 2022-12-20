package settings

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-ini/ini"
)

type (
	Settings struct {
		Path       string
		API        string
		Region     string
		Mode       string
		AutoAccept bool
		AutoRune   bool
		AutoSpell  bool
		AutoStart  bool
		PageName   string
		Reverse    bool
		Version    string
		Date       string
	}
)

func NewSettings() *Settings {
	var configuration, _ = os.UserConfigDir()
	var home = filepath.Join(configuration, "senna")
	var path = filepath.Join(home, "settings.ini")

	_, err := os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE|os.O_EXCL,
		0666,
	)

	file, _ := ini.Load(path)

	if err == nil {
		date := time.Now().Format("01-02-2006")

		file.NewSection("senna")
		file.Section("senna").NewKey("api", "https://localhost.com:5000")
		file.Section("senna").NewKey("region", "kr")
		file.Section("senna").NewKey("mode", "aram")
		file.Section("senna").NewKey("autoaccept", "true")
		file.Section("senna").NewKey("autorune", "false")
		file.Section("senna").NewKey("autospell", "false")
		file.Section("senna").NewKey("autostart", "false")
		file.Section("senna").NewKey("page", "senna")
		file.Section("senna").NewKey("reverse", "false")
		file.Section("senna").NewKey("version", "0.0.1")
		file.Section("senna").NewKey("date", date)

		err = file.SaveTo(path)
	}

	section := file.Section("senna")
	api := section.Key("api").String()
	region := section.Key("region").String()
	mode := section.Key("mode").String()
	autoAccept, _ := section.Key("autoaccept").Bool()
	autoRune, _ := section.Key("autorune").Bool()
	autoSpell, _ := section.Key("autospell").Bool()
	autoStart, _ := section.Key("autostart").Bool()
	pageName := section.Key("page").String()
	reverse, _ := section.Key("reverse").Bool()
	version := section.Key("version").String()
	date := section.Key("date").String()

	return &Settings{
		Path:       path,
		API:        api,
		Region:     region,
		Mode:       mode,
		AutoAccept: autoAccept,
		AutoRune:   autoRune,
		AutoSpell:  autoSpell,
		AutoStart:  autoStart,
		PageName:   pageName,
		Reverse:    reverse,
		Version:    version,
		Date:       date,
	}
}

func (settings *Settings) Replace(k string, v string) {
	file, _ := ini.Load(settings.Path)

	file.Section("senna").Key(k).SetValue(v)
	_ = file.SaveTo(settings.Path)
}

func (settings *Settings) Save(widget map[string]string) {
	file, _ := ini.Load(settings.Path)

	file.Section("senna").Key("api").SetValue(
		widget["api"],
	)

	file.Section("senna").Key("region").SetValue(
		widget["region"],
	)

	file.Section("senna").Key("mode").SetValue(
		widget["mode"],
	)

	file.Section("senna").Key("autoaccept").SetValue(
		widget["autoAccept"],
	)

	file.Section("senna").Key("autorune").SetValue(
		widget["autoRune"],
	)

	file.Section("senna").Key("autospell").SetValue(
		widget["autoSpell"],
	)

	file.Section("senna").Key("autostart").SetValue(
		widget["autoStart"],
	)

	file.Section("senna").Key("page").SetValue(
		widget["pageName"],
	)

	file.Section("senna").Key("reverse").SetValue(
		widget["reverse"],
	)

	_ = file.SaveTo(settings.Path)
}

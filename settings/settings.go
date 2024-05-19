package settings

import (
	"os"
	"time"

	"github.com/braycarlson/senna/filesystem"
	"github.com/go-ini/ini"
)

type (
	Settings struct {
		filesystem *filesystem.Filesystem
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
	return &Settings{}
}

func (settings *Settings) Create() error {
	_, err := os.OpenFile(
		settings.filesystem.Settings,
		os.O_RDWR|os.O_CREATE|os.O_EXCL,
		0666,
	)

	file, err := ini.Load(settings.filesystem.Settings)

	date := time.Now().Format("01-02-2006")

	file.NewSection("senna")
	file.Section("senna").NewKey("api", "https://braycarlson.duckdns.org:4443")
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

	err = file.SaveTo(settings.filesystem.Settings)

	return err
}

func (settings *Settings) Load() {
	file, _ := ini.Load(settings.filesystem.Settings)

	section := file.Section("senna")
	settings.API = section.Key("api").String()
	settings.Region = section.Key("region").String()
	settings.Mode = section.Key("mode").String()
	settings.AutoAccept, _ = section.Key("autoaccept").Bool()
	settings.AutoRune, _ = section.Key("autorune").Bool()
	settings.AutoSpell, _ = section.Key("autospell").Bool()
	settings.AutoStart, _ = section.Key("autostart").Bool()
	settings.PageName = section.Key("page").String()
	settings.Reverse, _ = section.Key("reverse").Bool()
	settings.Version = section.Key("version").String()
	settings.Date = section.Key("date").String()
}

func (settings *Settings) Replace(k string, v string) {
	file, _ := ini.Load(settings.filesystem.Settings)

	file.Section("senna").Key(k).SetValue(v)
	_ = file.SaveTo(settings.filesystem.Settings)
}

func (settings *Settings) Save(widget map[string]string) {
	file, _ := ini.Load(settings.filesystem.Settings)

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

	_ = file.SaveTo(settings.filesystem.Settings)
}

func (settings *Settings) SetFilesystem(filesystem *filesystem.Filesystem) {
	settings.filesystem = filesystem
}

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2/dialog"
	"log"
	"strings"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/braycarlson/asol"
	"github.com/braycarlson/asol/request"
	"github.com/braycarlson/senna/asset"
	"github.com/braycarlson/senna/collection"
	"github.com/braycarlson/senna/filesystem"
	"github.com/braycarlson/senna/preferences"
	"github.com/braycarlson/senna/settings"
	"github.com/braycarlson/senna/ui"
	"github.com/minio/selfupdate"
	"github.com/ncruces/zenity"
)

type (
	Senna struct {
		*asol.Asol

		asset       *asset.Asset
		cache       *fastcache.Cache
		client      *request.HTTPClient
		filesystem  *filesystem.Filesystem
		preferences *preferences.Preferences
		settings    *settings.Settings
		ui          *ui.UI
	}
)

func NewSenna() *Senna {
	var asol *asol.Asol = asol.NewAsol()
	var client *request.HTTPClient = asol.Client()
	var cache *fastcache.Cache = fastcache.New(4194304)
	var filesystem *filesystem.Filesystem = filesystem.NewFilesystem()

	var settings *settings.Settings = settings.NewSettings()
	settings.SetFilesystem(filesystem)

	if !filesystem.Exist(filesystem.Settings) {
		settings.Create()
	}

	settings.Load()

	var asset *asset.Asset = asset.NewAsset()
	asset.SetCache(cache)
	asset.SetClient(client)
	asset.SetFilesystem(filesystem)
	asset.SetSettings(settings)

	if !filesystem.Exist(filesystem.Archive) {
		asset.Download()
	}

	var preferences *preferences.Preferences = preferences.NewPreferences()
	preferences.SetFilesystem(filesystem)

	if !filesystem.Exist(filesystem.Preferences) {
		preferences.Create()
	}

	var ui *ui.UI = ui.NewUI()
	ui.SetPreferences(preferences)
	ui.SetSettings(settings)

	ui.Create()

	return &Senna{
		asol,
		asset,
		cache,
		client,
		filesystem,
		preferences,
		settings,
		ui,
	}
}

func (senna *Senna) Cache() *fastcache.Cache {
	return senna.cache
}

func (senna *Senna) Preferences() *preferences.Preferences {
	return senna.preferences
}

func (senna *Senna) Settings() *settings.Settings {
	return senna.settings
}

func (senna *Senna) update() error {
	var builder strings.Builder
	builder.WriteString(senna.settings.API)
	builder.WriteString("/senna/download")
	url := builder.String()

	senna.client.SetWeb()
	request, err := senna.client.Get(url)
	response, err := senna.client.Request(request)

	buffer := bytes.NewReader(response)
	length := buffer.Len()

	reader, err := zip.NewReader(
		buffer,
		int64(length),
	)

	for _, file := range reader.File {
		executable, _ := file.Open()

		err = selfupdate.Apply(
			executable,
			selfupdate.Options{},
		)

		if err != nil {
			if rerr := selfupdate.RollbackError(err); rerr != nil {
				fmt.Println("Failed to rollback from bad update: %v", rerr)
			}
		}

		executable.Close()
		return err
	}

	return nil
}

func (senna *Senna) logger(message string) {
	senna.ui.Event.History.Append(message)
	log.Println(message)
}

func (senna *Senna) isUpdate() bool {
	var builder strings.Builder
	builder.WriteString(senna.settings.API)
	builder.WriteString("/senna/version")
	url := builder.String()

	senna.client.SetWeb()
	request, _ := senna.client.Get(url)
	response, err := senna.client.Request(request)

	if err != nil {
		log.Println(err)
		return false
	}

	version := make(map[string]string)
	json.Unmarshal(response, &version)

	if senna.settings.Version == version["version"] {
		return false
	}

	senna.cache.Set(
		[]byte("version"),
		[]byte(version["version"]),
	)

	return true
}

func (senna *Senna) isRiotUpdate() bool {
	previous := senna.settings.Date
	current := time.Now().Format("01-02-2006")

	if current == previous {
		return false
	}

	return true
}

func main() {
	multiwriter := logger()
	defer multiwriter()

	senna := NewSenna()

	senna.logger("senna is checking for an update")
	update := senna.isUpdate()

	if update {
		senna.logger("senna found an update")

		confirm := dialog.NewConfirm(
			"Update",
			"Do you want to download the latest version of senna?",
			func(ok bool) {
				if ok {
					senna.logger("senna is updating")

					version := string(
						senna.cache.Get(
							[]byte{},
							[]byte("version"),
						),
					)

					senna.settings.Replace("version", version)
					err := senna.update()

					if err == nil {
						senna.ui.Restart()
					}
				}

				senna.logger("senna is ignoring the update")
			},
			senna.ui.Window,
		)

		confirm.Show()
	}

	if senna.isRiotUpdate() {
		err := senna.asset.Download()

		if err != nil {
			zenity.Error(
				"senna was unable to download the asset file",
				zenity.Title("Error"),
			)

			log.Println("senna was unable to download the asset file")

			if !senna.asset.IsDownloaded() {
				return
			}
		}

		current := time.Now().Format("01-02-2006")
		senna.settings.Replace("date", current)

		senna.preferences.Update()
	}

	senna.asset.Load()

	senna.ui.Window.SetOnClosed(func() {
		if senna.ui.Home.State.Text == "Stop" {
			go senna.Stop()
		}
	})

	senna.OnSearch(senna.onSearch)
	senna.OnOpen(senna.onOpen)
	senna.OnReady(senna.onReady)
	senna.OnLogin(senna.onLogin)
	senna.OnProcessError(senna.onProcessError)
	senna.OnSearchError(senna.onSearchError)
	senna.OnWebsocketClose(senna.onWebsocketClose)
	senna.OnWebsocketError(senna.onWebsocketError)

	senna.OnMessage(
		"/lol-matchmaking/v1/ready-check",
		"Update",
		senna.onMatchFound,
	)

	senna.OnMessage(
		"/lol-gameflow/v1/gameflow-phase",
		"Update",
		senna.onPhase,
	)

	senna.OnMessage(
		"/lol-champ-select/v1/session",
		"Update",
		senna.onSession,
	)

	senna.OnMessage(
		"/process-control/v1/process",
		"Update",
		senna.onProcessClose,
	)

	if senna.settings.AutoStart {
		if senna.ui.Home.State.Text == "Start" {
			senna.ui.Home.State.Importance = 0
			senna.ui.Home.State.Text = "Stop"

			go senna.Start()
		}

		senna.ui.Home.State.Refresh()
	}

	senna.ui.Settings.Restart.OnTapped = func() {
		senna.ui.Restart()
	}

	senna.ui.Home.State.OnTapped = func() {
		if senna.ui.Home.State.Text == "Start" {
			senna.ui.Home.State.Importance = 0
			senna.ui.Home.State.Text = "Stop"

			go senna.Start()
		} else {
			senna.ui.Home.State.Importance = 1
			senna.ui.Home.State.Text = "Start"
			senna.ui.Clear()

			go senna.Stop()
		}

		senna.ui.Home.State.Refresh()
	}

	senna.ui.Home.Classic.OnTapped = func() {
		gameflow := senna.cache.Get(
			[]byte{},
			[]byte("gameflow"),
		)

		if len(gameflow) == 0 {
			dialog.ShowInformation(
				"Warning",
				"You must be in champion selection",
				senna.ui.Window,
			)

			return
		}

		_, ok := senna.cache.HasGet(
			[]byte{},
			[]byte("championId"),
		)

		if ok {
			pageId := senna.cache.Get(
				[]byte{},
				[]byte("pageId"),
			)

			var status string
			var payload []byte

			var mode string = "classic"
			var region string = "kr"

			runepage := &collection.Runepage{
				Cache:    senna.cache,
				Client:   senna.client,
				Settings: senna.settings,
			}

			payload = runepage.Get(
				&collection.Local{
					collection.Parameter{
						Asset:  "runepage",
						Mode:   mode,
						Region: region,
					},
				},
			)

			runepage.Delete(pageId)
			runepage.Set(payload)
			runepage.Update()

			status = runepage.Status()
			senna.ui.Event.History.Append(status)

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
						Mode:   mode,
						Region: region,
					},
				},
			)

			spell.Set(payload)

			status = spell.Status()
			senna.ui.Event.History.Append(status)

			itemset := &collection.Itemset{
				Cache:    senna.cache,
				Client:   senna.client,
				Settings: senna.settings,
			}

			payload = itemset.Get(
				&collection.Local{
					collection.Parameter{
						Asset:  "itemset",
						Mode:   mode,
						Region: region,
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
						Mode:   mode,
						Region: region,
					},
				},
			)

			status = skillorder.Status() + skillorder.Process(payload)
			senna.ui.Event.History.Append(status)
		}
	}

	senna.ui.Home.ARAM.OnTapped = func() {
		gameflow := senna.cache.Get(
			[]byte{},
			[]byte("gameflow"),
		)

		if len(gameflow) == 0 {
			dialog.ShowInformation(
				"Warning",
				"You must be in champion selection",
				senna.ui.Window,
			)

			return
		}

		_, ok := senna.cache.HasGet(
			[]byte{},
			[]byte("championId"),
		)

		if ok {
			pageId := senna.cache.Get(
				[]byte{},
				[]byte("pageId"),
			)

			var status string
			var payload []byte

			var mode string = "aram"
			var region string = "kr"

			runepage := &collection.Runepage{
				Cache:    senna.cache,
				Client:   senna.client,
				Settings: senna.settings,
			}

			payload = runepage.Get(
				&collection.Local{
					collection.Parameter{
						Asset:  "runepage",
						Mode:   mode,
						Region: region,
					},
				},
			)

			runepage.Delete(pageId)
			runepage.Set(payload)
			runepage.Update()

			status = runepage.Status()
			senna.ui.Event.History.Append(status)

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
						Mode:   mode,
						Region: region,
					},
				},
			)

			spell.Set(payload)

			status = spell.Status()
			senna.ui.Event.History.Append(status)

			itemset := &collection.Itemset{
				Cache:    senna.cache,
				Client:   senna.client,
				Settings: senna.settings,
			}

			payload = itemset.Get(
				&collection.Local{
					collection.Parameter{
						Asset:  "itemset",
						Mode:   mode,
						Region: region,
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
						Mode:   mode,
						Region: region,
					},
				},
			)

			status = skillorder.Status() + skillorder.Process(payload)
			senna.ui.Event.History.Append(status)
		}
	}

	senna.ui.Settings.Save.OnTapped = func() {
		widget := senna.ui.Settings.Get()
		senna.settings.Save(widget)

		senna.settings = nil
		senna.settings = settings.NewSettings()
	}

	senna.ui.Preferences.Champion.OnChanged = senna.ui.Preferences.SetPreference
	senna.ui.Preferences.Save.OnTapped = senna.ui.Preferences.SavePreference
	senna.ui.Preferences.Reset.OnTapped = senna.ui.Preferences.ResetPreference

	senna.ui.Window.ShowAndRun()
}

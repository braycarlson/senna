package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"strconv"

	"github.com/braycarlson/senna/settings"
)

type (
	SettingsWidget struct {
		panel      *fyne.Container
		api        *widget.Entry
		region     *widget.Select
		mode       *widget.Select
		autoAccept *widget.Select
		autoRune   *widget.Select
		autoSpell  *widget.Select
		autoStart  *widget.Select
		pageName   *widget.Entry
		reverse    *widget.Select
		Save       *widget.Button
		Restart    *widget.Button
		settings   *settings.Settings
	}
)

func NewSettingsWidget() *SettingsWidget {
	return &SettingsWidget{}
}

func (settings *SettingsWidget) Create() {
	api := widget.NewEntry()

	api.SetText(settings.settings.API)

	region := widget.NewSelect(
		[]string{"kr"},
		nil,
	)

	region.SetSelected(settings.settings.Region)

	mode := widget.NewSelect(
		[]string{"aram", "classic"},
		nil,
	)

	mode.SetSelected(settings.settings.Mode)

	autoAccept := widget.NewSelect(
		[]string{"true", "false"},
		nil,
	)

	autoAccept.SetSelected(
		strconv.FormatBool(settings.settings.AutoAccept),
	)

	autoRune := widget.NewSelect(
		[]string{"true", "false"},
		nil,
	)

	autoRune.SetSelected(
		strconv.FormatBool(settings.settings.AutoRune),
	)

	autoSpell := widget.NewSelect(
		[]string{"true", "false"},
		nil,
	)

	autoSpell.SetSelected(
		strconv.FormatBool(settings.settings.AutoSpell),
	)

	autoStart := widget.NewSelect(
		[]string{"true", "false"},
		nil,
	)

	autoStart.SetSelected(
		strconv.FormatBool(settings.settings.AutoStart),
	)

	pageName := widget.NewEntry()

	pageName.SetText(settings.settings.PageName)

	reverse := widget.NewSelect(
		[]string{"true", "false"},
		nil,
	)

	reverse.SetSelected(
		strconv.FormatBool(settings.settings.Reverse),
	)

	save := widget.NewButton("Save", nil)
	save.Importance = 1

	restart := widget.NewButton("Restart", nil)
	restart.Importance = 0

	size := fyne.NewSize(750, 450)

	panel := container.New(
		layout.NewCenterLayout(),
		container.New(
			layout.NewGridWrapLayout(size),
			container.NewGridWithRows(
				11,
				container.New(
					layout.NewGridLayout(2),
					canvas.NewText("API", color.White),
					api,
				),
				container.New(
					layout.NewGridLayout(2),
					canvas.NewText("Region", color.White),
					region,
				),
				container.New(
					layout.NewGridLayout(2),
					canvas.NewText("Mode", color.White),
					mode,
				),
				container.New(
					layout.NewGridLayout(2),
					canvas.NewText("Autoaccept", color.White),
					autoAccept,
				),
				container.New(
					layout.NewGridLayout(2),
					canvas.NewText("Autorune", color.White),
					autoRune,
				),
				container.New(
					layout.NewGridLayout(2),
					canvas.NewText("Autospell", color.White),
					autoSpell,
				),
				container.New(
					layout.NewGridLayout(2),
					canvas.NewText("Autostart", color.White),
					autoStart,
				),
				container.New(
					layout.NewGridLayout(2),
					canvas.NewText("Page", color.White),
					pageName,
				),
				container.New(
					layout.NewGridLayout(2),
					canvas.NewText("Reverse", color.White),
					reverse,
				),
				layout.NewSpacer(),
				container.New(
					layout.NewGridLayout(2),
					save,
					restart,
				),
			),
		),
	)

	settings.panel = panel
	settings.api = api
	settings.region = region
	settings.mode = mode
	settings.autoAccept = autoAccept
	settings.autoRune = autoRune
	settings.autoSpell = autoSpell
	settings.autoStart = autoStart
	settings.pageName = pageName
	settings.reverse = reverse
	settings.Save = save
	settings.Restart = restart
}

func (settings *SettingsWidget) Get() map[string]string {
	return map[string]string{
		"api":        settings.api.Text,
		"region":     settings.region.Selected,
		"mode":       settings.mode.Selected,
		"autoAccept": settings.autoAccept.Selected,
		"autoRune":   settings.autoRune.Selected,
		"autoSpell":  settings.autoSpell.Selected,
		"autoStart":  settings.autoStart.Selected,
		"pageName":   settings.pageName.Text,
		"reverse":    settings.reverse.Selected,
	}
}

func (widget *SettingsWidget) SetSettings(settings *settings.Settings) {
	widget.settings = settings
}

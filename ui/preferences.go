package ui

import (
	"encoding/json"
	"log"
	"os"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"

	"github.com/braycarlson/senna/filesystem"
	"github.com/braycarlson/senna/model"
	"github.com/braycarlson/senna/preferences"
)

type (
	PreferencesWidget struct {
		panel      *fyne.Container
		filesystem *filesystem.Filesystem
		Champion   *widget.Select
		Identifier map[string]string
		Preference map[string]model.Preference
		Spell      []string
		ARAM_X     *widget.Select
		ARAM_Y     *widget.Select
		Classic_X  *widget.Select
		Classic_Y  *widget.Select
		OFA_X      *widget.Select
		OFA_Y      *widget.Select
		URF_X      *widget.Select
		URF_Y      *widget.Select
		Save       *widget.Button
		Reset      *widget.Button
	}
)

func NewPreferencesWidget() *PreferencesWidget {
	var preferences = preferences.NewPreferences()
	preference := preferences.Preferences()

	var name []string
	identifier := map[string]string{}

	for k, v := range preference {
		identifier[v.Name] = k
		name = append(name, v.Name)
	}

	sort.Strings(name)

	spell := []string{
		"Cleanse",
		"Exhaust",
		"Flash",
		"Ghost",
		"Heal",
		"Ignite",
		"Barrier",
		"Clarity",
		"Snowball",
		"Smite",
		"Teleport",
	}

	size := fyne.NewSize(750, 450)

	champion := widget.NewSelect(
		name,
		func(_ string) {},
	)

	champion.PlaceHolder = "Please select a champion"

	aram_x := widget.NewSelect(
		spell,
		func(_ string) {},
	)

	aram_x.PlaceHolder = " "

	aram_y := widget.NewSelect(
		spell,
		func(_ string) {},
	)

	aram_y.PlaceHolder = " "

	classic_x := widget.NewSelect(
		spell,
		func(_ string) {},
	)

	classic_x.PlaceHolder = " "

	classic_y := widget.NewSelect(
		spell,
		func(_ string) {},
	)

	classic_y.PlaceHolder = " "

	ofa_x := widget.NewSelect(
		spell,
		func(_ string) {},
	)

	ofa_x.PlaceHolder = " "

	ofa_y := widget.NewSelect(
		spell,
		func(_ string) {},
	)

	ofa_y.PlaceHolder = " "

	urf_x := widget.NewSelect(
		spell,
		func(_ string) {},
	)

	urf_x.PlaceHolder = " "

	urf_y := widget.NewSelect(
		spell,
		func(_ string) {},
	)

	urf_y.PlaceHolder = " "

	save := widget.NewButton("Save", nil)
	save.Importance = 1

	reset := widget.NewButton("Reset", nil)
	reset.Importance = 0

	panel := container.New(
		layout.NewCenterLayout(),
		container.New(
			layout.NewGridWrapLayout(size),
			container.NewGridWithRows(
				11,
				container.New(
					layout.NewMaxLayout(),
					champion,
				),
				layout.NewSpacer(),
				container.New(
					layout.NewGridLayout(3),
					canvas.NewText("ARAM", color.White),
					aram_x,
					aram_y,
				),
				container.New(
					layout.NewGridLayout(3),
					canvas.NewText("Classic", color.White),
					classic_x,
					classic_y,
				),
				container.New(
					layout.NewGridLayout(3),
					canvas.NewText("One for All", color.White),
					ofa_x,
					ofa_y,
				),
				container.New(
					layout.NewGridLayout(3),
					canvas.NewText("URF", color.White),
					urf_x,
					urf_y,
				),
				layout.NewSpacer(),
				layout.NewSpacer(),
				layout.NewSpacer(),
				layout.NewSpacer(),
				container.New(
					layout.NewGridLayout(2),
					save,
					reset,
				),
			),
		),
	)

	return &PreferencesWidget{
		panel:      panel,
		filesystem: filesystem.NewFilesystem(),
		Champion:   champion,
		Spell:      spell,
		Identifier: identifier,
		Preference: preference,
		ARAM_X:     aram_x,
		ARAM_Y:     aram_y,
		Classic_X:  classic_x,
		Classic_Y:  classic_y,
		OFA_X:      ofa_x,
		OFA_Y:      ofa_y,
		URF_X:      urf_x,
		URF_Y:      urf_y,
		Save:       save,
		Reset:      reset,
	}
}

func (preferences *PreferencesWidget) SavePreference() {
	current := preferences.Champion.Selected

	if current == "" {
		return
	}

	id := preferences.Identifier[current]
	preference := preferences.Preference[id]

	preference.ARAM.X = preferences.ARAM_X.Selected
	preference.ARAM.Y = preferences.ARAM_Y.Selected
	preference.Classic.X = preferences.Classic_X.Selected
	preference.Classic.Y = preferences.Classic_Y.Selected
	preference.OneForAll.X = preferences.OFA_X.Selected
	preference.OneForAll.Y = preferences.OFA_Y.Selected
	preference.URF.X = preferences.URF_X.Selected
	preference.URF.Y = preferences.URF_Y.Selected

	preferences.Preference[id] = preference

	if err := os.Truncate(preferences.filesystem.Preferences, 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	json, _ := json.MarshalIndent(preferences.Preference, "", "\t")
	_ = os.WriteFile(preferences.filesystem.Preferences, json, 0644)
}

func (preferences *PreferencesWidget) ResetPreference() {
	current := preferences.Champion.Selected

	if current == "" {
		return
	}

	id := preferences.Identifier[current]
	preference := preferences.Preference[id]

	preferences.ARAM_X.SetSelected(preference.ARAM.X)
	preferences.ARAM_Y.SetSelected(preference.ARAM.Y)
	preferences.Classic_X.SetSelected(preference.Classic.X)
	preferences.Classic_Y.SetSelected(preference.Classic.Y)
	preferences.URF_X.SetSelected(preference.URF.X)
	preferences.URF_Y.SetSelected(preference.URF.Y)
	preferences.OFA_X.SetSelected(preference.OneForAll.X)
	preferences.OFA_Y.SetSelected(preference.OneForAll.Y)
}

func (preferences *PreferencesWidget) SetPreference(name string) {
	id := preferences.Identifier[name]
	preference := preferences.Preference[id]

	preferences.ARAM_X.SetSelected(preference.ARAM.X)
	preferences.ARAM_Y.SetSelected(preference.ARAM.Y)
	preferences.Classic_X.SetSelected(preference.Classic.X)
	preferences.Classic_Y.SetSelected(preference.Classic.Y)
	preferences.OFA_X.SetSelected(preference.OneForAll.X)
	preferences.OFA_Y.SetSelected(preference.OneForAll.Y)
	preferences.URF_X.SetSelected(preference.URF.X)
	preferences.URF_Y.SetSelected(preference.URF.Y)
}

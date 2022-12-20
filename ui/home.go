package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type (
	HomeWidget struct {
		panel    *fyne.Container
		Username binding.String
		State    *widget.Button
		Classic  *widget.Button
		ARAM     *widget.Button
	}
)

func NewHomeWidget() *HomeWidget {
	username := binding.NewString()

	size := fyne.NewSize(750, 400)

	aram := widget.NewButton("ARAM", nil)
	aram.Importance = 0

	classic := widget.NewButton("Classic", nil)
	classic.Importance = 0

	state := widget.NewButton("Start", nil)
	state.Importance = 1

	panel := container.New(
		layout.NewCenterLayout(),
		container.New(
			layout.NewGridWrapLayout(size),
			container.New(
				layout.NewVBoxLayout(),
				container.New(
					layout.NewGridLayoutWithRows(5),
					layout.NewSpacer(),
					container.New(
						layout.NewCenterLayout(),
						widget.NewLabelWithData(username),
					),
					layout.NewSpacer(),
					container.New(
						layout.NewGridLayout(2),
						classic,
						aram,
					),
					state,
				),
			),
		),
	)

	return &HomeWidget{
		panel:    panel,
		Username: username,
		State:    state,
		Classic:  classic,
		ARAM:     aram,
	}
}

func (home *HomeWidget) clear() {
	home.Username.Set("")
}

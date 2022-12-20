package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type (
	EventLogWidget struct {
		panel   *fyne.Container
		History binding.ExternalStringList
		list    *widget.List
	}
)

func NewEventLogWidget() *EventLogWidget {
	history := binding.BindStringList(
		&[]string{"senna is awaiting a connection"},
	)

	list := widget.NewListWithData(
		history,
		func() fyne.CanvasObject {
			return widget.NewLabel("history")
		},
		func(item binding.DataItem, object fyne.CanvasObject) {
			object.(*widget.Label).Bind(item.(binding.String))
		},
	)

	callback := binding.NewDataListener(
		func() {
			list.ScrollToBottom()
		},
	)

	history.AddListener(callback)

	button := widget.NewButton("Clear", func() {
		// Temporary: https://github.com/fyne-io/fyne/issues/3100
		data, _ := history.Get()
		data = nil
		history.Set(data)
	})

	size := fyne.NewSize(750, 450)

	panel := container.New(
		layout.NewCenterLayout(),
		container.New(
			layout.NewGridWrapLayout(size),
			container.New(
				layout.NewBorderLayout(nil, button, nil, nil),
				button,
				list,
			),
		),
	)

	return &EventLogWidget{
		panel:   panel,
		History: history,
		list:    list,
	}
}

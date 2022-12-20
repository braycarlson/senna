package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"log"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

type (
	UI struct {
		App         fyne.App
		Window      fyne.Window
		Home        *HomeWidget
		Event       *EventLogWidget
		Preferences *PreferencesWidget
		Settings    *SettingsWidget
	}
)

func NewUI() *UI {
	app := app.New()
	window := app.NewWindow("senna")

	window.Resize(fyne.NewSize(1000, 500))
	window.CenterOnScreen()
	window.SetMaster()

	var home *HomeWidget = NewHomeWidget()
	var event *EventLogWidget = NewEventLogWidget()
	var preferences *PreferencesWidget = NewPreferencesWidget()
	var settings *SettingsWidget = NewSettingsWidget()

	tabs := container.NewAppTabs(
		container.NewTabItem(
			"Home",
			home.panel,
		),
		container.NewTabItem(
			"Event Log",
			event.panel,
		),
		container.NewTabItem(
			"Preferences",
			preferences.panel,
		),
		container.NewTabItem(
			"Settings",
			settings.panel,
		),
	)

	tabs.SetTabLocation(container.TabLocationLeading)
	window.SetContent(tabs)

	var ui UI
	ui.App = app
	ui.Window = window
	ui.Home = home
	ui.Event = event
	ui.Preferences = preferences
	ui.Settings = settings

	return &ui
}

func (ui *UI) Clear() {
	ui.Home.clear()
}

func (ui *UI) Restart() {
	executable, err := os.Executable()

	if err != nil {
		log.Println(err)
	}

	args := os.Args
	env := os.Environ()

	ui.App.Quit()

	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command(executable, args[1:]...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = env

		err := cmd.Run()

		if err == nil {
			os.Exit(0)
		}
	default:
		syscall.Exec(executable, args, env)
	}
}

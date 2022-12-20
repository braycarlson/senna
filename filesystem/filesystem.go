package filesystem

import (
	"errors"
	"os"
	"path/filepath"
)

type (
	Filesystem struct {
		Archive     string
		Asset       string
		Home        string
		Log         string
		Preferences string
		Settings    string
	}
)

func NewFilesystem() *Filesystem {
	var configuration, _ = os.UserConfigDir()

	// Directories
	var home = filepath.Join(configuration, "senna")
	os.Mkdir(home, os.ModePerm)

	var asset = filepath.Join(home, "asset")
	os.Mkdir(asset, os.ModePerm)

	// Files
	var archive = filepath.Join(asset, "asset.zip")
	var log = filepath.Join(home, "log.txt")
	var preferences = filepath.Join(home, "preferences.json")
	var settings = filepath.Join(home, "settings.ini")

	return &Filesystem{
		Archive:     archive,
		Asset:       asset,
		Home:        home,
		Log:         log,
		Preferences: preferences,
		Settings:    settings,
	}
}

func (filesystem *Filesystem) Exist(filename string) bool {
	_, err := os.OpenFile(
		filename,
		os.O_RDONLY,
		0444,
	)

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

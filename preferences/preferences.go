package preferences

import (
	"archive/zip"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/braycarlson/senna/filesystem"
	"github.com/braycarlson/senna/model"
)

type (
	Preferences struct {
		filesystem *filesystem.Filesystem
	}
)

func NewPreferences() *Preferences {
	return &Preferences{}
}

func NewDefaultPreference(name string, opgg string) model.Preference {
	return model.Preference{
		Name: name,
		OPGG: strings.ToLower(opgg),

		ARAM: model.ARAM{
			X: "Flash",
			Y: "Snowball",
		},
		Classic: model.Classic{
			X: "Flash",
			Y: "Ghost",
		},
		OneForAll: model.OneForAll{
			X: "Flash",
			Y: "Teleport",
		},
		URF: model.URF{
			X: "Flash",
			Y: "Ghost",
		},
	}
}

func (preferences *Preferences) Champion() model.Champions {
	reader, _ := zip.OpenReader(preferences.filesystem.Archive)
	defer reader.Close()

	for _, file := range reader.File {
		filename := strings.TrimSuffix(
			filepath.Base(file.Name),
			filepath.Ext(file.Name),
		)

		if filename == "champion" {
			champion, _ := file.Open()
			defer champion.Close()

			buffer, _ := io.ReadAll(champion)

			var champions model.Champions
			json.Unmarshal(buffer, &champions)

			return champions
		}
	}

	return nil
}

func (preferences *Preferences) championIdentifier() []string {
	champion := preferences.Champion()

	identifier := make(
		[]string,
		len(champion),
	)

	index := 0

	for key, _ := range champion {
		identifier[index] = key
		index++
	}

	return identifier
}

func (preferences *Preferences) Create() {
	os.OpenFile(
		preferences.filesystem.Preferences,
		os.O_RDWR|os.O_CREATE|os.O_EXCL,
		0666,
	)

	champions := preferences.Champion()
	champion := make(map[string]model.Preference)

	for id, data := range champions {
		champion[id] = NewDefaultPreference(
			data.Name,
			data.Key,
		)
	}

	preference := &model.Preferences{
		Champion: champion,
	}

	os.OpenFile(
		preferences.filesystem.Preferences,
		os.O_CREATE|os.O_EXCL,
		0755,
	)

	json, _ := json.MarshalIndent(preference, "", "\t")
	_ = os.WriteFile(preferences.filesystem.Preferences, json, 0644)
}

func (preferences *Preferences) Path() string {
	return preferences.filesystem.Preferences
}

func (preferences *Preferences) Preferences() map[string]model.Preference {
	file, err := os.Open(preferences.filesystem.Preferences)
	defer file.Close()

	if err != nil {
		log.Println(err)
	}

	information, err := file.Stat()

	if err != nil {
		log.Println(err)
	}

	if information.Size() == 0 {
		preferences.Create()
	}

	data, _ := os.ReadFile(preferences.filesystem.Preferences)

	var preference map[string]model.Preference
	json.Unmarshal(data, &preference)

	return preference
}

func (preferences *Preferences) Preference(id string) model.Preference {
	file := preferences.Preferences()
	return file[id]
}

func getDifference(x, y []string) []string {
	temporary := make(
		map[string]struct{},
		len(y),
	)

	for _, v := range y {
		temporary[v] = struct{}{}
	}

	var difference []string

	for _, v := range x {
		if _, found := temporary[v]; !found {
			difference = append(difference, v)
		}
	}

	return difference
}

func (preferences *Preferences) isDifference() ([]string, error) {
	identifier := preferences.championIdentifier()
	preference := preferences.Preferences()

	var cid, pid []string

	for _, id := range identifier {
		cid = append(cid, id)
	}

	for id, _ := range preference {
		pid = append(pid, id)
	}

	id := getDifference(cid, pid)
	return id, nil
}

func (preferences *Preferences) SetFilesystem(filesystem *filesystem.Filesystem) {
	preferences.filesystem = filesystem
}

func (preferences *Preferences) Update() {
	missing, _ := preferences.isDifference()

	if len(missing) == 0 {
		return
	}

	champions := preferences.Champion()
	champion := make(map[string]model.Preference)

	for id, data := range champions {
		for mid := range missing {
			tid := strconv.Itoa(mid)

			if tid == id {
				champion[id] = NewDefaultPreference(
					data.Name,
					data.Key,
				)
			}
		}
	}

	preference := &model.Preferences{
		Champion: champion,
	}

	file := preferences.Preferences()

	for k, v := range preference.Champion {
		file[k] = v
	}

	if err := os.Truncate(preferences.filesystem.Preferences, 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	json, _ := json.MarshalIndent(file, "", "\t")
	_ = os.WriteFile(preferences.filesystem.Preferences, json, 0644)
}

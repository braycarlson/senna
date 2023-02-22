package asset

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/braycarlson/asol/request"
	"github.com/braycarlson/senna/filesystem"
	"github.com/braycarlson/senna/settings"
)

type (
	Strategy interface {
		load(asset *Asset) error
	}

	File struct{}

	Archive struct{}

	Asset struct {
		cache      *fastcache.Cache
		client     *request.HTTPClient
		filesystem *filesystem.Filesystem
		settings   *settings.Settings
		strategy   Strategy
	}
)

func NewAsset() *Asset {
	var strategy *Archive = &Archive{}

	return &Asset{
		strategy: strategy,
	}
}

func (asset *Asset) IsDownloaded() bool {
	file, err := os.Stat(asset.filesystem.Archive)

	if errors.Is(err, os.ErrNotExist) || file.Size() < 5120 {
		return false
	}

	return true
}

func (asset *Asset) Download() error {
	var builder strings.Builder
	builder.WriteString(asset.settings.API)
	builder.WriteString("/senna/assets")
	url := builder.String()

	asset.client.SetWeb()
	request, _ := asset.client.Get(url)
	data, err := asset.client.Request(request)

	if err != nil {
		return err
	}

	archive, _ := os.OpenFile(
		asset.filesystem.Archive,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0755,
	)

	defer archive.Close()
	_, err = archive.Write(data)
	return err
}

func (asset *Asset) SetCache(cache *fastcache.Cache) {
	asset.cache = cache
}

func (asset *Asset) SetClient(client *request.HTTPClient) {
	asset.client = client
}

func (asset *Asset) SetFilesystem(filesystem *filesystem.Filesystem) {
	asset.filesystem = filesystem
}

func (asset *Asset) SetSettings(settings *settings.Settings) {
	asset.settings = settings
}

func (asset *Asset) setStrategy(strategy Strategy) {
	asset.strategy = strategy
}

func (asset *Asset) Load() error {
	err := asset.strategy.load(asset)
	return err
}

func (file *File) isExtracted(asset *Asset) bool {
	directory, _ := os.ReadDir(asset.filesystem.Asset)

	if len(directory) == 1 {
		return false
	}

	return true
}

func (file *File) extract(asset *Asset) {
	reader, _ := zip.OpenReader(asset.filesystem.Archive)
	defer reader.Close()

	for _, file := range reader.File {
		buffer, _ := file.Open()

		content, _ := os.OpenFile(
			filepath.Join(asset.filesystem.Asset, file.Name),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0777,
		)

		io.Copy(content, buffer)

		content.Close()
		buffer.Close()
	}
}

func (file *File) load(asset *Asset) error {
	if !file.isExtracted(asset) {
		file.extract(asset)
	}

	directory, _ := os.ReadDir(asset.filesystem.Asset)

	for _, path := range directory {
		filename := path.Name()

		extension := filepath.Ext(filename)

		if extension == ".json" {
			full := filepath.Join(
				asset.filesystem.Asset,
				filename,
			)

			buffer, _ := os.ReadFile(full)

			asset.cache.SetBig(
				[]byte(filename),
				buffer,
			)
		}
	}

	return nil
}

func (archive *Archive) load(asset *Asset) error {
	reader, err := zip.OpenReader(asset.filesystem.Archive)
	defer reader.Close()

	for _, file := range reader.File {
		content, _ := file.Open()
		defer content.Close()

		buffer, _ := io.ReadAll(content)

		asset.cache.SetBig(
			[]byte(file.Name),
			buffer,
		)
	}

	return err
}

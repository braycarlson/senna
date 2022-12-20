package collection

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/braycarlson/asol/request"
	"github.com/braycarlson/senna/model"
	"github.com/braycarlson/senna/settings"
)

type (
	Runepage struct {
		Cache    *fastcache.Cache
		Client   *request.HTTPClient
		Settings *settings.Settings
	}
)

func (runepage *Runepage) setPageName(data []byte) []byte {
	var page model.Runepage
	json.Unmarshal(data, &page)

	page.Name = runepage.Settings.PageName
	modified, _ := json.Marshal(page)

	return modified
}

func (runepage *Runepage) Delete(pageId []byte) {
	if len(pageId) > 0 {
		page := string(pageId)

		var builder strings.Builder
		builder.WriteString("/lol-perks/v1/pages/")
		builder.WriteString(page)

		url := builder.String()

		runepage.Client.SetWebsocket()
		request, _ := runepage.Client.Delete(url)
		runepage.Client.Request(request)
	}
}

func (runepage *Runepage) Set(payload []byte) error {
	runepage.Client.SetWebsocket()
	request, _ := runepage.Client.Post("/lol-perks/v1/pages", payload)
	_, err := runepage.Client.Request(request)

	return err
}

func (runepage *Runepage) Status() string {
	championName := string(
		runepage.Cache.Get(
			[]byte{},
			[]byte("championName"),
		),
	)

	var builder strings.Builder
	builder.WriteString("Setting runepage for ")
	builder.WriteString(championName)

	return builder.String()
}

func (runepage *Runepage) Update() {
	var pages []model.Page

	runepage.Client.SetWebsocket()
	request, _ := runepage.Client.Get("/lol-perks/v1/pages")
	response, _ := runepage.Client.Request(request)
	json.Unmarshal(response, &pages)

	for _, page := range pages {
		if runepage.Settings.PageName == page.Name {
			pageId := strconv.FormatFloat(page.Id, 'f', -1, 64)

			runepage.Cache.Set(
				[]byte("pageId"),
				[]byte(pageId),
			)
		}
	}
}

func (runepage *Runepage) Get(visitor Visitor) []byte {
	return visitor.getRunepage(runepage)
}

func (local *Local) getRunepage(runepage *Runepage) []byte {
	championId := string(
		runepage.Cache.Get(
			[]byte{},
			[]byte("championId"),
		),
	)

	var parameter string = local.build()

	file := runepage.Cache.GetBig(
		[]byte{},
		[]byte(parameter),
	)

	var runepages model.Runepages
	json.Unmarshal(file, &runepages)

	page := runepages[championId]
	byte, _ := json.Marshal(page)

	modified := runepage.setPageName(byte)
	return modified
}

func (web *Web) getRunepage(runepage *Runepage) []byte {
	championId := string(
		runepage.Cache.Get(
			[]byte{},
			[]byte("championId"),
		),
	)

	var parameter string = web.build()

	var builder strings.Builder
	builder.WriteString(runepage.Settings.API)
	builder.WriteString(parameter)
	builder.WriteString(championId)
	url := builder.String()

	runepage.Client.SetWeb()
	request, _ := runepage.Client.Get(url)
	response, _ := runepage.Client.Request(request)

	modified := runepage.setPageName(response)
	return modified
}

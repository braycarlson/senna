package collection

import (
	"encoding/json"
	"strings"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/braycarlson/asol/request"
	"github.com/braycarlson/senna/model"
	"github.com/braycarlson/senna/settings"
)

type (
	Itemset struct {
		Cache    *fastcache.Cache
		Client   *request.HTTPClient
		Settings *settings.Settings
	}
)

func (itemset *Itemset) Set(payload []byte) error {
	var set model.Itemset
	json.Unmarshal(payload, &set)

	accountId := string(
		itemset.Cache.Get(
			[]byte{},
			[]byte("accountId"),
		),
	)

	set.AccountId = accountId
	payload, _ = json.Marshal(set)

	var builder strings.Builder
	builder.WriteString("/lol-item-sets/v1/item-sets/")
	builder.WriteString(accountId)
	builder.WriteString("/sets")

	url := builder.String()

	itemset.Client.SetWebsocket()
	request, err := itemset.Client.Put(url, payload)
	_, err = itemset.Client.Request(request)

	return err
}

func (itemset *Itemset) Status() string {
	championName := string(
		itemset.Cache.Get(
			[]byte{},
			[]byte("championName"),
		),
	)

	var builder strings.Builder
	builder.WriteString("Setting itemset for ")
	builder.WriteString(championName)

	return builder.String()
}

func (itemset *Itemset) Get(visitor Visitor) []byte {
	return visitor.getItemset(itemset)
}

func (local *Local) getItemset(itemset *Itemset) []byte {
	championId := string(
		itemset.Cache.Get(
			[]byte{},
			[]byte("championId"),
		),
	)

	var parameter string = local.build()

	file := itemset.Cache.GetBig(
		[]byte{},
		[]byte(parameter),
	)

	var itemsets model.Itemsets
	json.Unmarshal(file, &itemsets)

	set := itemsets[championId]

	byte, _ := json.Marshal(set)
	return byte
}

func (web *Web) getItemset(itemset *Itemset) []byte {
	championId := string(
		itemset.Cache.Get(
			[]byte{},
			[]byte("championId"),
		),
	)

	var parameter string = web.build()

	var builder strings.Builder
	builder.WriteString(itemset.Settings.API)
	builder.WriteString(parameter)
	builder.WriteString(championId)
	url := builder.String()

	itemset.Client.SetWeb()
	request, _ := itemset.Client.Get(url)
	response, _ := itemset.Client.Request(request)

	return response
}

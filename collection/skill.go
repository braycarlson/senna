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
	SkillOrder struct {
		Cache    *fastcache.Cache
		Client   *request.HTTPClient
		Settings *settings.Settings
	}
)

func (skillorder *SkillOrder) Process(payload []byte) string {
	var skill []string
	json.Unmarshal(payload, &skill)

	var builder strings.Builder

	builder.WriteString(
		strings.Trim(
			strings.Join(skill, " "),
			"[]",
		),
	)

	return builder.String()
}

func (skillorder *SkillOrder) Status() string {
	championName := string(
		skillorder.Cache.Get(
			[]byte{},
			[]byte("championName"),
		),
	)

	var builder strings.Builder
	builder.WriteString("Skill Order for ")
	builder.WriteString(championName)
	builder.WriteString(": ")

	return builder.String()
}

func (skillorder *SkillOrder) Get(visitor Visitor) []byte {
	return visitor.getSkillOrder(skillorder)
}

func (local *Local) getSkillOrder(skillorder *SkillOrder) []byte {
	championId := string(
		skillorder.Cache.Get(
			[]byte{},
			[]byte("championId"),
		),
	)

	var parameter string = local.build()

	file := skillorder.Cache.GetBig(
		[]byte{},
		[]byte(parameter),
	)

	var skillorders model.SkillOrders
	json.Unmarshal(file, &skillorders)

	order := skillorders[championId]

	byte, _ := json.Marshal(order)
	return byte
}

func (web *Web) getSkillOrder(skillorder *SkillOrder) []byte {
	championId := string(
		skillorder.Cache.Get(
			[]byte{},
			[]byte("championId"),
		),
	)

	var parameter string = web.build()

	var builder strings.Builder
	builder.WriteString(skillorder.Settings.API)
	builder.WriteString(parameter)
	builder.WriteString(championId)
	url := builder.String()

	skillorder.Client.SetWeb()
	request, _ := skillorder.Client.Get(url)
	response, _ := skillorder.Client.Request(request)

	return response
}

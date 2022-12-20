package collection

import (
	"strings"
)

type (
	Local struct {
		Parameter Parameter
	}

	Web struct {
		Parameter Parameter
	}

	Visitable interface {
		Get(Visitor) []byte
	}

	Visitor interface {
		getItemset(*Itemset) []byte
		getSkillOrder(*SkillOrder) []byte
		getSummonerSpell(*SummonerSpell) []byte
		getRunepage(*Runepage) []byte
	}

	Parameter struct {
		Asset  string
		Mode   string
		Region string
	}
)

func (local *Local) build() string {
	var builder strings.Builder

	builder.WriteString(local.Parameter.Mode)
	builder.WriteString("_")
	builder.WriteString(local.Parameter.Asset)
	builder.WriteString("_")
	builder.WriteString(local.Parameter.Region)
	builder.WriteString(".json")

	return builder.String()
}

func (web *Web) build() string {
	var builder strings.Builder

	builder.WriteString("/")
	builder.WriteString(web.Parameter.Mode)
	builder.WriteString("/")
	builder.WriteString(web.Parameter.Asset)
	builder.WriteString("/")
	builder.WriteString(web.Parameter.Region)
	builder.WriteString("/")

	return builder.String()
}

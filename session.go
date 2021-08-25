package main

type (
	Session struct {
		accountId    string
		championId   string
		championName string
		mode         string
		pageId       string
		summonerId   string
		username     string
	}
)

func NewSession() *Session {
	return &Session{}
}

func (session *Session) resetSession() {
	session.championId = ""
	session.championName = ""
	session.mode = ""
}

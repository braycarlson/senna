package main

import (
	"fmt"

	"github.com/braycarlson/asol/client"
	"github.com/braycarlson/asol/session"
)

var asol *client.Client

func onOpen(session *session.Session) {
	fmt.Println("Opened")
}

func onReady(session *session.Session) {
	fmt.Println("Ready")
}

func onLogin(session *session.Session) {
	fmt.Println("Logged in")
}

func onLogout(session *session.Session) {
	fmt.Println("Logged out")
}

func onClientClose(session *session.Session) {
	fmt.Println("Client closed")
}

func onWebsocketClose(session *session.Session) {
	fmt.Println("Websocket closed")
}

func onReconnect(session *session.Session) {
	fmt.Println("Reconnected")
}

func onError(error error) {
	fmt.Println(error)
}

func main() {
	asol := client.NewClient()

	asol.OnOpen(onOpen)
	asol.OnReady(onReady)
	asol.OnLogin(onLogin)
	asol.OnLogout(onLogout)
	asol.OnClientClose(onClientClose)
	asol.OnWebsocketClose(onWebsocketClose)
	asol.OnReconnect(onReconnect)
	asol.OnError(onError)

	asol.Start()
}

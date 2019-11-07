package controller

import "net/http"

var ws = newWebsocketController()

func Initialize() {
	registerRoutes()
}

func registerRoutes() {
	http.HandleFunc("/ws", ws.handleMessage)
}

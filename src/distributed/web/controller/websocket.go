package controller

import (
	"bytes"
	"encoding/gob"
	"log"
	"net/http"
	"sync"

	"github.com/collinewait/pp-monitoring-system/src/distributed/dto"
	"github.com/collinewait/pp-monitoring-system/src/distributed/qutils"
	"github.com/collinewait/pp-monitoring-system/src/distributed/web/model"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

const url = "amqp://guest:guest@localhost:5672"

type websocketController struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	sockets  []*websocket.Conn
	mutex    sync.Mutex
	upgrader websocket.Upgrader
}

func newWebsocketController() *websocketController {
	wsc := new(websocketController)
	wsc.conn, wsc.ch = qutils.GetChannel(url)
	wsc.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	go wsc.listenForSources()
	go wsc.listenForMessages()

	return wsc
}

func (wsc *websocketController) handleMessage(w http.ResponseWriter, r *http.Request) {
	socket, _ := wsc.upgrader.Upgrade(w, r, nil)
	wsc.addSocket(socket)
	go wsc.ListenForDiscoveryRequests(socket)
}

func (wsc *websocketController) addSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	wsc.sockets = append(wsc.sockets, socket)
	wsc.mutex.Unlock()
}

func (wsc *websocketController) removeSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	socket.Close()

	for i := range wsc.sockets {
		if wsc.sockets[i] == socket {
			wsc.sockets = append(wsc.sockets[:i], wsc.sockets[i+1:]...)
		}
	}
	wsc.mutex.Unlock()
}

func (wsc *websocketController) listenForSources() {
	q := qutils.GetQueue("", wsc.ch, true)
	wsc.ch.QueueBind(
		q.Name,
		"",
		qutils.WebappSourceExchange,
		false,
		nil,
	)

	msgs, _ := wsc.ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	for msg := range msgs {
		sensor, err := model.GetSensorByName(string(msg.Body))

		if err != nil {
			log.Println(err.Error())
		}
		wsc.sendMessage(message{
			Type: "source",
			Data: sensor,
		})
	}
}

func (wsc *websocketController) listenForMessages() {
	q := qutils.GetQueue("", wsc.ch, true)
	wsc.ch.QueueBind(
		q.Name,
		"",
		qutils.WebappReadingsExchange,
		false,
		nil,
	)

	msgs, _ := wsc.ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	for msg := range msgs {
		buf := bytes.NewBuffer(msg.Body)
		dec := gob.NewDecoder(buf)
		sm := dto.SensorMessage{}
		dec.Decode(&sm)

		wsc.sendMessage(message{
			Type: "reading",
			Data: sm,
		})
	}
}

func (wsc *websocketController) sendMessage(msg message) {
	sockectsToRemove := []*websocket.Conn{}

	for _, socket := range wsc.sockets {
		err := socket.WriteJSON(msg)

		if err != nil {
			sockectsToRemove = append(sockectsToRemove, socket)
		}
	}

	for _, socket := range sockectsToRemove {
		wsc.removeSocket(socket)
	}
}

func (wsc *websocketController) ListenForDiscoveryRequests(socket *websocket.Conn) {
	for {
		msg := message{}
		err := socket.ReadJSON(&msg)

		if err != nil {
			wsc.removeSocket(socket)
			break
		}

		if msg.Type == "discover" {
			wsc.ch.Publish(
				"",
				qutils.WebappDiscoveryQueue,
				false,
				false,
				amqp.Publishing{},
			)
		}
	}
}

type message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

package main

import (
	"errors"
	"sync/atomic"
	"time"

	websocket "github.com/gorilla/websocket"
)

var (
	// Time allowed to read the next pong message from the client.
	pongWait time.Duration

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod time.Duration

	// Time allowed to write the file to the client.
	writeWait time.Duration
)

func init() {
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait = 10 * time.Second
}

type WS struct {
	ws       *websocket.Conn
	msg_chan chan []byte
	dead     int32
}

func (ws *WS) listener() {
	ws.ws.SetReadLimit(512)
	ws.ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.ws.SetPongHandler(func(string) error { ws.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ws.ReadMessage()
		if err != nil {
			atomic.StoreInt32(&ws.dead, 1)
			break
		}
	}
}

func (ws *WS) serve() {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		atomic.StoreInt32(&ws.dead, 1)
		pingTicker.Stop()
		ws.ws.Close()
	}()
	for {
		select {
		case msg, ok := <-ws.msg_chan:
			if !ok {
				return
			}
			if msg != nil {
				ws.ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			}
		case <-pingTicker.C:
			ws.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}

}

func (ws *WS) ListenAndServe() {
	go ws.serve()
	ws.listener()
}

func (ws *WS) Send(msg []byte) error {
	select {
	case ws.msg_chan <- msg:
		return nil
	default:
	}
	return errors.New("channel is full or closed")
}

func (ws *WS) Close() error {
	close(ws.msg_chan)
	return nil
}

func (ws *WS) Alive() bool {
	return atomic.LoadInt32(&ws.dead) == 0
}

func NewWS(ws *websocket.Conn) *WS {
	return &WS{
		ws:       ws,
		msg_chan: make(chan []byte, 10),
	}
}

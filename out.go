package main

import (
	"sync"

	"golang.org/x/net/websocket"
)

// chan buffer size
const BufferSize = 8

// Live2DDriver forwards messages to connected clients, that are, Live2DViews.
type Live2DDriver struct {
	msgChans []chan []byte
	mu       sync.RWMutex
}

func NewLive2DDriver() *Live2DDriver {
	return &Live2DDriver{
		msgChans: []chan []byte{},
	}
}

// ForwardMessageTo the WebSocket connection.
//
// Use SendMessage to send messages.
func (d *Live2DDriver) ForwardMessageTo(ws *websocket.Conn) {
	ch := make(chan []byte, BufferSize)

	// add

	d.mu.Lock()
	d.msgChans = append(d.msgChans, ch)
	d.mu.Unlock()

	verboseLogf("Start ForwardMessageTo: %s by chan %v.", ws.RemoteAddr(), ch)

	// forward

	messageForwarder(ch, ws) // 阻塞

	// clean up

	close(ch)

	d.mu.Lock()
	for i, c := range d.msgChans {
		if c == ch {
			d.msgChans = append(d.msgChans[:i], d.msgChans[i+1:]...)
			break
		}
	}
	d.mu.Unlock()

	verboseLogf("Stop ForwardMessageTo: %s by chan %v.", ws.RemoteAddr(), ch)
}

// SendMessage to WebSocket clients.
func (d *Live2DDriver) SendMessage(msg []byte) {
	verboseLogf("SendMessage: %s", string(msg))

	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, ch := range d.msgChans {
		if ch != nil {
			ch <- msg
		}
	}
}

// deprecated 似乎会搞出更多 goroutines 不划算了。
// 直接传个 *Live2DDriver，自己调 SendMessage 多舒服 (see in.go)。
func (d *Live2DDriver) ForwardMessageFrom(msgCh <-chan []byte) {
	for msg := range msgCh {
		d.SendMessage(msg)
	}
}

// messageForwarder forwards messages from the message channel to the websocket
// connection.
//
// The message channel is expected to receive JSON strings (bytes):
//
//	`{"motion": "shake"}`
//	`{"expression": "f03"}`
func messageForwarder(msgCh <-chan []byte, ws *websocket.Conn) {
	for msg := range msgCh {
		verboseLogf("fwd msg: %s -> %s (chan %v).", string(msg), ws.RemoteAddr(), msgCh)
		_, err := ws.Write(msg)
		if err != nil {
			verboseLogf("fwd msg to %s (chan %v) error: %s.", ws.RemoteAddr(), msgCh, err)
			break
		}
	}
	ws.Close()
}

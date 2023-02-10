package main

import (
	"sync"

	"golang.org/x/net/websocket"
)

// chan buffer size
const BufferSize = 8

type MessageForwarder interface {
	ForwardMessageTo(ws *websocket.Conn)
	SendMessage(msg []byte)
	ForwardMessageFrom(msgCh <-chan []byte)
}

// messageForwarder forwards messages to connected clients, that are, Live2DViews.
type messageForwarder struct {
	msgChans []chan []byte
	mu       sync.RWMutex
}

func NewMessageForwarder() *messageForwarder {
	return &messageForwarder{
		msgChans: []chan []byte{},
	}
}

// ForwardMessageTo the WebSocket connection.
//
// Use SendMessage to send messages.
//
// Block until the websocket connection is closed.
func (f *messageForwarder) ForwardMessageTo(ws *websocket.Conn) {
	ch := make(chan []byte, BufferSize)

	// add

	f.mu.Lock()
	f.msgChans = append(f.msgChans, ch)
	f.mu.Unlock()

	verboseLogf("Start ForwardMessageTo: %s by chan %v.", ws.RemoteAddr(), ch)

	// forward

	forwardMessage(ch, ws) // 阻塞

	// clean up

	close(ch)

	f.mu.Lock()
	for i, c := range f.msgChans {
		if c == ch {
			f.msgChans = append(f.msgChans[:i], f.msgChans[i+1:]...)
			break
		}
	}
	f.mu.Unlock()

	verboseLogf("Stop ForwardMessageTo: %s by chan %v.", ws.RemoteAddr(), ch)
}

// SendMessage to WebSocket clients.
//
// Block until message is sent to all clients.
func (f *messageForwarder) SendMessage(msg []byte) {
	verboseLogf("SendMessage: %s", string(msg))

	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, ch := range f.msgChans {
		if ch != nil {
			ch <- msg
		}
	}
}

// deprecated 似乎会搞出更多 goroutines 不划算了。
// 直接传个 *Live2DDriver，自己调 SendMessage 多舒服 (see in.go)。
//
// Block until the message channel is closed.
func (f *messageForwarder) ForwardMessageFrom(msgCh <-chan []byte) {
	for msg := range msgCh {
		f.SendMessage(msg)
	}
}

// forwardMessage forwards messages from the message channel to the websocket
// connection.
//
// The message channel is expected to receive JSON strings (bytes):
//
//	`{"motion": "shake"}`
//	`{"expression": "f03"}`
func forwardMessage(msgCh <-chan []byte, ws *websocket.Conn) {
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

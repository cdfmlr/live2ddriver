package wsforwarder

import (
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func TestMessageForwarder(t *testing.T) {
	f := NewMessageForwarder()

	// msg src
	msgCh := make(chan []byte, 1)

	// in
	go func() {
		t.Logf("messageForwarder: forward message from %v", msgCh)
		f.ForwardMessageFrom(msgCh)
	}()

	// out
	go func() {
		http.Handle("/ws_fwd_test", websocket.Handler(func(c *websocket.Conn) {
			t.Logf("websocket connection from %s", c.RemoteAddr().String())
			f.ForwardMessageTo(c)
		}))

		t.Logf("ListenAndServe WebSocket: %s", ":9101")
		if err := http.ListenAndServe(":9101", nil); err != nil {
			panic("ListenAndServe: " + err.Error())
		}
	}()

	// test

	time.Sleep(1 * time.Second)

	msgs := []string{"hello", "world"}

	t.Logf("websocket.Dial: %s", "ws://localhost:9101/ws_fwd_test")
	client, err := websocket.Dial("ws://localhost:9101/ws_fwd_test", "", "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}

	// goroutine 1: send
	go func() {
		time.Sleep(1 * time.Second)

		for _, msg := range msgs {
			t.Logf("send: %s", msg)
			msgCh <- []byte(msg)
		}
	}()

	// goroutine 2: recv
	for _, msg := range msgs {
		var recvMsg string
		if err := websocket.Message.Receive(client, &recvMsg); err != nil {
			t.Errorf("websocket.Message.Receive: %s", err.Error())
		}
		if recvMsg != msg {
			t.Errorf("recvMsg != msg: %s != %s", recvMsg, msg)
		}
		t.Logf("recv: %s", recvMsg)
	}

	client.Close()
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

// #region CLI

var (
	wsAddr   = flag.String("wsAddr", ":9001", "(out) forward messages to WebSocket server address")
	httpAddr = flag.String("httpAddr", ":9002", "(in) forward messages from HTTP server address. Empty to disable.")
	stdin    = flag.Bool("stdin", false, "(in) forward messages from stdin")
	verbose  = flag.Bool("verbose", false, "verbose mode")
)

func cli() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Printf("Forward messages from stdin | http to WebSocket clients.\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if !*stdin && *httpAddr == "" {
		fmt.Fprintf(os.Stderr, "Error: no input source: -stdin or -httpAddr is required.\n")
		flag.Usage()
		os.Exit(1)
	}
}

// #endregion CLI

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	cli()

	forwarder := NewMessageForwarder()

	http.Handle("/live2d", websocket.Handler(func(c *websocket.Conn) {
		forwarder.ForwardMessageTo(c)
	}))

	if *stdin {
		go func() {
			forwarder.ForwardMessageFromStdin()
		}()
	}
	if *httpAddr != "" {
		go func() {
			forwarder.ForwardMessageFromHTTP(*httpAddr)
		}()
	}

	verboseLogf("(out) Listening WebSocket on %s/live2d...\n", *wsAddr)
	if err := http.ListenAndServe(*wsAddr, nil); err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

// #region log

func verboseLogf(format string, a ...interface{}) {
	if *verbose {
		log.Printf(format, a...)
	}
}

// #endregion log

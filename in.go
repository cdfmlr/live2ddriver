package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// ForwardMessageFromHTTP read Live2DRequest from HTTP request and send it to MessageForwarder.
func ForwardMessageFromHTTP(msgForwarder *messageForwarder, addr string) {
	verboseLogf("(in) Forwarding messages from HTTP (%s/live2d) to WebSocket clients...\n", addr)

	router := gin.Default()
	router.GET("/live2d", func(c *gin.Context) {
		var req Live2DRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		j, err := json.Marshal(req)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		msgForwarder.SendMessage(j)
	})
	router.Run(addr)
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// ForwardMessageFromStdin read Live2DRequest from stdin and send it to MessageForwarder.
func ForwardMessageFromStdin(msgForwarder *messageForwarder) {
	verboseLogf("(in) Forwarding messages from stdin to WebSocket clients...\n")
	time.Sleep(time.Millisecond * 200) // 太快了日志和输入提示交错不好看
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Enter a message to send: ")
	for {
		scanner.Scan()
		msgForwarder.SendMessage(scanner.Bytes())
		time.Sleep(time.Millisecond * 200) // 太快了日志和输入提示交错不好看
		fmt.Printf("Enter a message to send: ")
	}
}

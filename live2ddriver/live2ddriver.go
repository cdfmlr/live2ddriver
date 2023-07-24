package live2ddriver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

type Live2DDriver interface {
	// recv textIn and generate Live2DRequest
	Drive(textIn string) Live2DRequest
}

// DriveLive2DChan do:
//
//	chIn -> text -> driver -> Live2DRequest -> chOut
//
// no blocking
func DriveLive2DChan(driver Live2DDriver, chIn <-chan string) (chOut chan<- Live2DRequest) {
	chOut = make(chan Live2DRequest, BufferSize)
	go func() {
		for textIn := range chIn {
			res := driver.Drive(textIn)

			if reflect.ValueOf(res).IsZero() {
				continue
			}

			chOut <- res
		}
	}()
	return chOut
}

// DriveLive2DHTTP listen on addr and serve http request.
// Get text from request body, and return Live2DRequest as response.
// The Live2DRequest will be generated by driver. And send to chOut
// after json.Marshal.
//
// No blocking.
func DriveLive2DHTTP(driver Live2DDriver, addr string) (chOut chan []byte) {
	chOut = make(chan []byte, BufferSize)
	go func() {
		router := gin.New()
		// router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 	return fmt.Sprintf("%s INFO [GIN] %s %s %s => %v %s",
		// 		param.TimeStamp.Format("2006/01/02 15:04:05"),
		// 		param.ClientIP, param.Method, param.Path,
		// 		param.StatusCode, param.ErrorMessage,
		// 	)
		// }))
		router.Use(gin.Recovery())
		router.POST("/driver", func(c *gin.Context) {
			body := c.Request.Body
			defer body.Close()

			text, err := ioutil.ReadAll(body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			res := driver.Drive(string(text))
			if reflect.ValueOf(res).IsZero() {
				c.JSON(http.StatusBadRequest, gin.H{"warn": "empty Drive req"})
				return
			}

			j, err := json.Marshal(res)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			chOut <- j
		})
		router.Run(addr)
	}()
	return chOut
}

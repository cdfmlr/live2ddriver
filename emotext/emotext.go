package emotext

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

var EmotextServer = "http://localhost:9003/"

var client = &http.Client{}

func init() {
	client.Timeout = 5 * time.Second

	es := os.Getenv("EMOTEXT_SERVER")
	if es != "" {
		EmotextServer = es
	}
}

// Query returns the emotions of the given text.
//
// The returned EmotionResult contains the emotion in **21 categories**, and the
// polarity of the text.
//
// Use Emotions21To7 to convert the 21 categories to 7 categories if needed:
//
//	result, _ := emotext.Query(text)
//	result.Emotions = Emotions21To7(result.Emotions)
func Query(text string) (EmotionResult, error) {
	var result EmotionResult

	body := strings.NewReader(text)
	resp, err := client.Post(EmotextServer, "text/plain", body)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}

package emotext

import (
	"testing"
)

// before go test, run:
//  cd ../emotext
//  poetry shell
//  PYTHONPATH=$PYTHONPATH:. python emotext/httpapi.py --port 9003

func TestEmotext(t *testing.T) {
	w := make(chan struct{}, 1)

	var result EmotionResult
	var err error

	t.Run("query", func(t *testing.T) {
		defer func() {
			w <- struct{}{}
		}()

		result, err = Query("我很开心")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(result)
	})
	<-w

	t.Run("e21to7", func(t *testing.T) {
		defer func() {
			w <- struct{}{}
		}()
		e7 := Emotions21To7(result.Emotions)
		t.Log(e7)
	})
	<-w
}

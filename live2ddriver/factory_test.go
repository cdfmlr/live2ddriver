package live2ddriver

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestEmoMapperFactory_Create(t *testing.T) {
	jsonConfig := `{
      "type": "stateless",
      "config": {
        "motionFromEmotion": {
          "happiness": "tap_body",
          "sadness": "flick_head"
        },
        "expressionFromPolarity": {
          "positive": "f04"
        }
      }
    }`

	expected := NewStatelessEmoMapper(
		map[EmotionsKey]Motion{
			"happiness": Motion("tap_body"),
			"sadness":   Motion("flick_head"),
		},
		map[PolarityKey]Expression{
			"positive": Expression("f04"),
		},
	)

	var factory EmoMapperFactory

	// json -> factory

	err := json.Unmarshal([]byte(jsonConfig), &factory)
	if err != nil {
		t.Errorf("json.Unmarshal failed: %v", err)
	}

	// factory -> mapper

	emoMapper, err := factory.Create()
	if err != nil {
		t.Errorf("factory.Create failed: %v", err)
	}

	// test

	if emoMapper == nil {
		t.Errorf("emoMapper is nil")
	}

	if _, ok := emoMapper.(*statelessEmoMapper); !ok {
		t.Errorf("emoMapper is not a statelessEmoMapper")
	}

	// compare the two mappers

	got := emoMapper.(*statelessEmoMapper)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("emoMapper: got %v, want %v", got, expected)
	}
}

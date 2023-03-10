package live2ddriver

import (
	"live2ddriver/emotext"
	"sync"

	"golang.org/x/exp/constraints"
)

// This file is a driver for following model (shizuku):
// https://cdn.jsdelivr.net/gh/guansss/pixi-live2d-display/test/assets/shizuku/shizuku.model.json

// #region shizukuExpressions

type shizukuExpression = string

const (
	shizukuExpressionShy     shizukuExpression = "f01"
	shizukuExpressionUnhappy shizukuExpression = "f02"
	shizukuExpressionAnger   shizukuExpression = "f03"
	shizukuExpressionSmile   shizukuExpression = "f04"
)

// Polarity => Expression
var shizukuExpressionFromPolarity = map[string]shizukuExpression{
	// from expressions
	// "happiness": shizukuExpressionShy,
	// "goodness":  shizukuExpressionShy,
	// "anger":     shizukuExpressionAnger,
	// "sadness":   shizukuExpressionUnhappy,
	// "fear":      shizukuExpressionSmile,
	// "dislike":   shizukuExpressionUnhappy,
	// "surprise":  shizukuExpressionSmile,

	"neutrality": shizukuExpressionShy,
	"positive":   shizukuExpressionSmile,
	"negative":   shizukuExpressionUnhappy,
	"both":       shizukuExpressionShy,
}

// #endregion shizukuExpressions

// #region shizukuMotions

type shizukuMotion = string

const (
	shizukuMotionIdle      shizukuMotion = "idle"
	shizukuMotionEnergetic shizukuMotion = "tap_body"
	shizukuMotionShy       shizukuMotion = "pinch_in"
	shizukuMotionTsundere  shizukuMotion = "pinch_out"
	shizukuMotionShock     shizukuMotion = "shake"
	shizukuMotionChat      shizukuMotion = "flick_head"
)

// Emotion => Motion
var shizukuMotionsFromEmotions = map[string]shizukuMotion{
	"happiness": shizukuMotionEnergetic,
	"goodness":  shizukuMotionChat,
	"anger":     shizukuMotionTsundere,
	"sadness":   shizukuMotionChat,
	"fear":      shizukuMotionShy,
	"dislike":   shizukuMotionTsundere,
	"surprise":  shizukuMotionShock,
}

// #endregion shizukuMotions

// #region shizukuDriver

type shizukuDriver struct {
	currentExpression shizukuExpression
	currentMotion     shizukuMotion

	emotion  emotext.Emotions // emotion => motion
	polarity emotext.Polarity // polarity => expression

	mu sync.RWMutex
}

func NewShizukuDriver() Live2DDriver {
	return &shizukuDriver{}
}

var EmotionChangeFactor float32 = 0.2

func (d *shizukuDriver) Drive(textIn string) Live2DRequest {
	d.updateEmotion(textIn)

	var req Live2DRequest

	d.mu.Lock()
	defer d.mu.Unlock()

	maxEmotion := keyOfMaxValue(d.emotion)
	maxPolarity := keyOfMaxValue(d.polarity)

	if d.currentExpression != shizukuExpressionFromPolarity[maxPolarity] {
		d.currentExpression = shizukuExpressionFromPolarity[maxPolarity]
		req.Expression = d.currentExpression
	}

	if d.currentMotion != shizukuMotionsFromEmotions[maxEmotion] {
		d.currentMotion = shizukuMotionsFromEmotions[maxEmotion]
		req.Motion = d.currentMotion
	}

	return req
}

// updateEmotion call emotext to analyze the text, and update the emotion
// and polarity of the driver.
//
// Lock inside.
func (d *shizukuDriver) updateEmotion(text string) error {
	emoResult, err := emotext.Query(text)
	if err != nil {
		return err
	}

	emotions := emotext.Emotions21To7(emoResult.Emotions)
	polarity := emoResult.Polarity

	// calculate: e = e + e0 * factor
	d.mu.RLock()
	// emotions.Add(&d.emotion, EmotionChangeFactor)
	for k, v := range d.emotion {
		emotions[k] += v * EmotionChangeFactor
	}
	// polarity.Add(&d.polarity, EmotionChangeFactor)
	for k, v := range d.polarity {
		polarity[k] += v * EmotionChangeFactor
	}
	d.mu.RUnlock()

	// writeback
	d.mu.Lock()
	d.emotion = emotions
	d.polarity = polarity
	d.mu.Unlock()

	return nil
}

// #endregion shizukuDriver

func keyOfMaxValue[K comparable, V constraints.Ordered](m map[K]V) K {
	var maxKey K = *new(K)
	var maxValue V = *new(V)

	for k, v := range m {
		if v >= maxValue {
			maxKey, maxValue = k, v
		}
	}

	return maxKey
}

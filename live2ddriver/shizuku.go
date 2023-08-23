package live2ddriver

// Model specific drivers are deprecated. Everything isto be removed in the 
// later v0.2.x. Keeping it here for compatibility & testing (emotion_test.go).

import (
	"live2ddriver/emotext"
	"sync"
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
//	def updateEmotion(self, e: Emotion):
//	    self.Emotion = (e + self.Emotion * factor) / sum(e + self.Emotion * factor)
//
// Lock inside.
func (d *shizukuDriver) updateEmotion(text string) error {
	emoResult, err := emotext.Query(text)
	if err != nil {
		return err
	}

	emotions := emotext.Emotions21To7(emoResult.Emotions)
	polarity := emoResult.Polarity

	// calculate: e = (e + e0 * factor) / sum(e + e0 * factor)

	var sumEmotion, sumPolarity float32 = 0.0, 0.0

	d.mu.RLock()
	// emotions.Add(&d.emotion, EmotionChangeFactor)
	for k, v := range d.emotion {
		emotions[k] += v * EmotionChangeFactor
		sumEmotion += emotions[k]
	}
	// polarity.Add(&d.polarity, EmotionChangeFactor)
	for k, v := range d.polarity {
		polarity[k] += v * EmotionChangeFactor
		sumPolarity += polarity[k]
	}
	d.mu.RUnlock()

	if sumEmotion <= 1e-6 {
		sumEmotion = 1.0
	}
	if sumPolarity <= 1e-6 {
		sumPolarity = 1.0
	}

	for k, v := range emotions {
		emotions[k] = v / sumEmotion
	}
	for k, v := range polarity {
		polarity[k] = v / sumPolarity
	}

	// writeback
	d.mu.Lock()
	d.emotion = emotions
	d.polarity = polarity
	d.mu.Unlock()

	return nil
}

// #endregion shizukuDriver

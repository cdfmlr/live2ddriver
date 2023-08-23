package live2ddriver

import (
	"math"

	"github.com/murchinroom/emotextcligo"
	"golang.org/x/exp/constraints"
)

type (
	Emotion emotextcligo.EmotionResult

	Motion     string
	Expression string

	EmotionsKey = string // typeof(key(Emotion.Emotions))
	PolarityKey = string // typeof(key(Emotion.Polarity))

)

// EmotionExpressionMapper is the interface for Emotion <-> Live2d Motion & Expression mapping.
type EmotionExpressionMapper interface {
	Map(e Emotion) (Motion, Expression)
}

// statelessEmoMapper is a stateless EmotionExpression implementation.
//
// Its full name is:
//
//	StatelessMaximumAPosterioriEmotionToLive2DMotionAndExpressionMapper
//
// It maps Emotion to Motion & Expression by maps.
// The maps are constructed by user.
// Mathematically it's a Maximum A Posteriori (MAP) Estimation.
type statelessEmoMapper struct {
	// Emotion => Motion
	MotionFromEmotion map[EmotionsKey]Motion
	// Polarity => Expression
	ExpressionFromPolarity map[PolarityKey]Expression
}

// NewStatelessEmoMapper returns a statelessMapper.
func NewStatelessEmoMapper(
	motionFromEmotion map[EmotionsKey]Motion,
	expressionFromPolarity map[PolarityKey]Expression,
) EmotionExpressionMapper {
	return &statelessEmoMapper{
		MotionFromEmotion:      motionFromEmotion,
		ExpressionFromPolarity: expressionFromPolarity,
	}
}

func (m *statelessEmoMapper) Map(e Emotion) (Motion, Expression) {
	motion := m.MotionFromEmotion[keyOfMaxValue(e.Emotions)]
	expression := m.ExpressionFromPolarity[keyOfMaxValue(e.Polarity)]

	return motion, expression
}

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

// statefulEmoMapper is a stateful EmotionExpression implementation.
//
// Its state is a short-term memory to keep & update the emotion:
//
//	u = sigmoid(a + b * x^{t-1} + c * h^{t-1})
//	r = sigmoid(d + e * x^{t-1} + f * h^{t-1})
//	h^t = u * x^{t-1} + (1 - u) * sigmoid(i + j * x^t + k * r * h^{t-1})
//
// in which, x^t is the new-gain emotion (i.e. stimulus) at time t, h^t is the updated emotion.
// a, b, c, d, e, f, i, j, k are the parameters of the model, all of them are
// in [-1, 1]. u is the update gate. r is the reset (forget) gate. The sigmoid
// function is defined as: sigmoid(x) = (1 + x / (1+|x|)) / 2.
//
// Let's explain it in a more intuitive way:
//
// Based on a hypothesis that the memory of emotion is short, we assume that
// it's a Markov process. So we can use the emotion at time t-1 (h^{t-1}) plus
// a new-gain emotion (x^t) to calculate the emotion at time t (h^{t}):
//
//	h^t = w_0 * x^t + w_1 * Markov(h^{t-1}, ..., h^{t-n})
//
// Stolen idea from the Gated recurrent units (GRUs), we use 2 gates to control
// the update of emotion: update gate (u) and reset gate (r).
//
// The difference is that we use u * x^{t-1} to replace the u * h^{t-1} in GRU.
// I feel that it's more reasonable here to make it feels like a Markov
// process.
//
// I set the parameters based on my intuition:
//
//	a = -1.0, b =  0.5, c = -0.5
//	d =  1.0, e = -0.3, f =  0.3
//	i =  0.0, j =  1.0, k = -0.1
//
// These numbers make the emotion more stable. Looks like the Lenz's law:
// It's tested (see Test_statefulEmoMapper in emotion_test.go) to be able to 
// forget the very-old memory but still keep sensitive to the newer memory & 
// the new stimulus.
//
// ⚠️  I am a math muggle. I don't know if it's correct or not even wrong.
//
// Its full name is:
//
//	StatefulMaximumAPosterioriEmotionToLive2DMotionAndExpressionMapper
type statefulEmoMapper struct {
	statelessEmoMapper // emotion -> motion & expression mapper

	emotionX Emotion // short-term memory: received emotion
	emotionH Emotion // short-term memory: updated emotion
}

func NewStatefulEmoMapper(
	motionFromEmotion map[EmotionsKey]Motion,
	expressionFromPolarity map[PolarityKey]Expression,
) EmotionExpressionMapper {
	return &statefulEmoMapper{
		statelessEmoMapper: statelessEmoMapper{
			MotionFromEmotion:      motionFromEmotion,
			ExpressionFromPolarity: expressionFromPolarity,
		},
		emotionX: Emotion{},
		emotionH: Emotion{},
	}
}

func (m *statefulEmoMapper) Map(e Emotion) (Motion, Expression) {
	var eCopy Emotion
	eCopy.Emotions = make(map[EmotionsKey]float32)
	for k, v := range e.Emotions {
		eCopy.Emotions[k] = v
	}
	eCopy.Polarity = make(map[PolarityKey]float32)
	for k, v := range e.Polarity {
		eCopy.Polarity[k] = v
	}

	// log.Printf("[DEBUG] statefulEmoMapper.Map: before update: %v", e)
	eCopy = m.updateEmotion(eCopy)
	// log.Printf("[DEBUG] statefulEmoMapper.Map: after update: %v", e)

	return m.statelessEmoMapper.Map(eCopy)
}

func (m *statefulEmoMapper) updateEmotion(e Emotion) (updated Emotion) {
	updated = Emotion{
		Emotions: make(map[EmotionsKey]float32),
		Polarity: make(map[PolarityKey]float32),
	}

	// Emotion.Emotions

	sumEmotions := float32(0)
	for _, v := range e.Emotions {
		sumEmotions += v
	}
	if sumEmotions > 1e-6 {
		for k, v := range e.Emotions {
			e.Emotions[k] = v / sumEmotions
		}
	}

	for k := range m.emotionH.Emotions {
		if _, ok := e.Emotions[k]; !ok {
			e.Emotions[k] = 0
		}
	}

	for k, xNew := range e.Emotions {
		xPrev := m.emotionX.Emotions[k]
		hPrev := m.emotionH.Emotions[k]

		hNew := m.hypothesisElem(float64(xNew), float64(xPrev), float64(hPrev))

		updated.Emotions[k] = float32(hNew)
	}

	// Emotion.Polarity

	sumPolarity := float32(0)
	for _, v := range e.Polarity {
		sumPolarity += v
	}
	if sumPolarity > 1e-6 {
		for k, v := range e.Polarity {
			e.Polarity[k] = v / sumPolarity
		}
	}

	for k := range m.emotionH.Polarity {
		if _, ok := e.Polarity[k]; !ok {
			e.Polarity[k] = 0
		}
	}

	for k, xNew := range e.Polarity {
		xPrev := m.emotionX.Polarity[k]
		hPrev := m.emotionH.Polarity[k]

		hNew := m.hypothesisElem(float64(xNew), float64(xPrev), float64(hPrev))

		updated.Polarity[k] = float32(hNew)
	}

	// sliding window

	m.emotionX = e
	m.emotionH = updated

	return updated
}

func (m *statefulEmoMapper) hypothesisElem(xNew, xPrev, hPrev float64) (hNew float64) {
	//	u = sigmoid(a + b * x^{t-1} + c * h^{t-1})
	//	r = sigmoid(d + e * x^{t-1} + f * h^{t-1})
	//	h^t = u * x^{t-1} + (1 - u) * sigmoid(i + j * x^t + k * r * h^{t-1})

	// a = -1, b =  0.5, c = -0.5
	// d =  1, e = -0.3, f =  0.3
	// i =  0, j =    1, k = -0.1

	u := sigmoid(-1 + 0.5*xPrev - 0.5*hPrev)
	r := sigmoid(1 - 0.3*xPrev + 0.3*hPrev)

	// sigmod(x) = 0.5 (1 + x / (1+abs(x))) 在 x \in [-1, 1] 时映射出来的 y 范围比较小，而这里想要更线性一点
	// 尝试写了两个函数，实验下来还不如下面这个真线性的 ¯\_(ツ)_/¯ 所以就用这个了（这里其实主要就是是要截负即可）
	linearS := func(x float64) float64 { // _/¯
		if x <= 0 {
			return 0
		} else if x >= 1 {
			return 1
		}
		return x
	}

	hNew = u*hPrev + (1-u)*linearS(0+1*xNew-0.1*r*hPrev)

	// log.Printf("[DEBUG] statefulEmoMapper.hypothesisElem: xNew = %v, xPrev = %v, hPrev = %v, hNew = %v, u = %v, r = %v\n", xNew, xPrev, hPrev, hNew, u, r)

	// never out of range
	if hNew > 1 {
		hNew = 1
	} else if hNew < 0 {
		hNew = 0
	}

	return hNew
}

func sigmoid(x float64) float64 {
	return (1 + x/(1+math.Abs(x))) / 2
}

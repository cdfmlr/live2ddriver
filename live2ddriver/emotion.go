package live2ddriver

import (
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

// statelessMapper is a stateless EmotionExpression implementation.
//
// Its full name is:
//
//	StatelessMaximumAPosterioriEmotionToLive2DMotionAndExpressionMapper
//
// It maps Emotion to Motion & Expression by maps.
// The maps are constructed by user.
// Mathematically it's a Maximum A Posteriori (MAP) Estimation.
type statelessMapper struct {
	// Emotion => Motion
	MotionFromEmotion map[EmotionsKey]Motion
	// Polarity => Expression
	ExpressionFromPolarity map[PolarityKey]Expression
}

// NewStatelessMapper returns a statelessMapper.
func NewStatelessMapper(
	motionFromEmotion map[EmotionsKey]Motion,
	expressionFromPolarity map[PolarityKey]Expression,
) EmotionExpressionMapper {
	return &statelessMapper{
		MotionFromEmotion:      motionFromEmotion,
		ExpressionFromPolarity: expressionFromPolarity,
	}
}

func (m *statelessMapper) Map(e Emotion) (Motion, Expression) {
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

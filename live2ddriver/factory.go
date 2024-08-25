package live2ddriver

import (
	"errors"
	"fmt"
)

//// Factory ////

// EmoMapperFactory is a helper to create EmotionExpressionMapper
// based on encodable configuration.
// (from YAML config file to EmotionExpressionMapper)
type EmoMapperFactory struct {
	Type   MapperType
	Config EmoMapperConfig
}

// Create an EmotionExpressionMapper based on the factory configuration.
func (f *EmoMapperFactory) Create() (EmotionExpressionMapper, error) {
	if err := f.Config.validate(); err != nil {
		return nil, err
	}

	switch f.Type {
	case StatelessEmoMapperType:
		return NewStatelessEmoMapper(f.Config.MotionFromEmotion, f.Config.ExpressionFromPolarity), nil
	case StatefulEmoMapperType:
		return NewStatefulEmoMapper(f.Config.MotionFromEmotion, f.Config.ExpressionFromPolarity), nil
	default:
		return nil, ErrInvalidEmoMapperType
	}
}

//// arg Type ////

// MapperType is an encodable representation of EmotionExpressionMapper type.
type MapperType = string

const (
	StatelessEmoMapperType MapperType = "stateless"
	StatefulEmoMapperType  MapperType = "stateful"
)

//// arg Config ////

// EmoMapperConfig is the configuration for EmotionExpressionMapper.
type EmoMapperConfig struct {
	// Emotion => Motion, e.g. "happiness" => "happy"
	MotionFromEmotion map[EmotionsKey]Motion // 其实就是 map[string]string
	// Polarity => Expression, e.g. "positive" => "smile"
	ExpressionFromPolarity map[PolarityKey]Expression // 其实就是 map[string]string
}

func (c *EmoMapperConfig) validate() error {
	if len(c.MotionFromEmotion) == 0 {
		return fmt.Errorf("%w: empty MotionFromEmotion", ErrInvalidEmoMapperConfig)
	}
	if len(c.ExpressionFromPolarity) == 0 {
		return fmt.Errorf("%w: empty ExpressionFromPolarity", ErrInvalidEmoMapperConfig)
	}
	return nil
}

//// errors ////

var (
	ErrInvalidEmoMapperType   = errors.New("invalid EmotionExpressionMapper type")
	ErrInvalidEmoMapperConfig = errors.New("invalid EmotionExpressionMapper config")
)

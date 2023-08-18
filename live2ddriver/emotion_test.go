package live2ddriver

import "testing"

func Test_statelessEmoMapper(t *testing.T) {
	// reuse the legacy shizuku emotion mapper (7 emotions)

	motionFromEmotion := map[EmotionsKey]Motion{}
	for k, v := range shizukuMotionsFromEmotions {
		motionFromEmotion[k] = Motion(v)
	}

	expressionFromPolarity := map[PolarityKey]Expression{}
	for k, v := range shizukuExpressionFromPolarity {
		expressionFromPolarity[k] = Expression(v)
	}

	emoMapper := NewStatelessEmoMapper(
		motionFromEmotion,
		expressionFromPolarity,
	)

	// tests
	type Args struct {
		emo Emotion
	}

	type Want struct {
		motion     Motion
		expression Expression
	}

	type testCase struct {
		name string
		args Args
		want Want
	}

	testCases := []testCase{
		{
			name: "EmptyEmotionEmptyPolarity",
			args: Args{emo: Emotion{}},
			want: Want{},
		},
		{
			name: "EmptyEmotionOnePolarity",
			args: Args{emo: Emotion{
				Polarity: map[PolarityKey]float32{"positive": 1.0},
			}},
			want: Want{
				expression: Expression(shizukuExpressionFromPolarity["positive"]),
			},
		},
		{
			name: "OneEmotionEmptyPolarity",
			args: Args{emo: Emotion{
				Emotions: map[EmotionsKey]float32{"happiness": 1.0},
			}},
			want: Want{
				motion: Motion(shizukuMotionsFromEmotions["happiness"]),
			},
		},
		{
			name: "OneEmotionOnePolarity",
			args: Args{emo: Emotion{
				Emotions: map[EmotionsKey]float32{"happiness": 1.0},
				Polarity: map[PolarityKey]float32{"positive": 1.0},
			}},
			want: Want{
				motion:     Motion(shizukuMotionsFromEmotions["happiness"]),
				expression: Expression(shizukuExpressionFromPolarity["positive"]),
			},
		},
		{
			name: "MultiEmotionMultiPolarity",
			args: Args{emo: Emotion{
				Emotions: map[EmotionsKey]float32{
					"happiness": 0.8,
					"sadness":   0.5,
				},
				Polarity: map[PolarityKey]float32{
					"positive": 0.5,
					"negative": 0.2,
				},
			}},
			want: Want{
				motion:     Motion(shizukuMotionsFromEmotions["happiness"]),
				expression: Expression(shizukuExpressionFromPolarity["positive"]),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			motion, expression := emoMapper.Map(tc.args.emo)

			if motion != tc.want.motion {
				t.Errorf("motion = %v, want %v", motion, tc.want.motion)
			}

			if expression != tc.want.expression {
				t.Errorf("expression = %v, want %v", expression, tc.want.expression)
			}

			t.Logf("emotion (%v) => motion (%v) & expression (%v)", tc.args.emo, motion, expression)
		})
	}
}

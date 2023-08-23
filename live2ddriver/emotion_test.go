package live2ddriver

import (
	"math"
	"testing"
)

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

func Test_statefulEmoMapper(t *testing.T) {
	// reuse the legacy shizuku emotion mapper (7 emotions)

	motionFromEmotion := map[EmotionsKey]Motion{}
	for k, v := range shizukuMotionsFromEmotions {
		motionFromEmotion[k] = Motion(v)
	}

	expressionFromPolarity := map[PolarityKey]Expression{}
	for k, v := range shizukuExpressionFromPolarity {
		expressionFromPolarity[k] = Expression(v)
	}

	emoMapper := NewStatefulEmoMapper(
		motionFromEmotion,
		expressionFromPolarity,
	).(*statefulEmoMapper)

	var ekv = func(k string) float64 {
		// 救命为啥以前写了 float32 ？？？
		return float64(emoMapper.emotionH.Emotions[k])
	}

	t.Run("1_happiness_1.0", func(t *testing.T) {
		// 1. 设置一个 happiness = 1.0 的 Emotion, fear 用来凑数（sum(emotions) = 1）并测试几次操作后的累积效果
		//    状态转移后期望接近 1.0
		e := Emotion{Emotions: map[EmotionsKey]float32{"happiness": 1.0, "fear": 0.0}}
		emoMapper.Map(e)
		if ekv("happiness") < 0.7 {
			t.Errorf("emotionH.Emotions[\"happiness\"] = %v, want >= 0.7", ekv("happiness"))
		}
		
		t.Logf("\ninput: %v\noutput: %v", e, emoMapper.emotionH)
	})

	t.Run("2_sadness_0.5", func(t *testing.T) {
		// 2. 设置一个 sadness = 0.5 的 Emotion，隐含 happiness = 0
		//    happiness 应该下降，sadness 更新后不应超过刺激值 0.5
		e := Emotion{Emotions: map[EmotionsKey]float32{"sadness": 0.5, "fear": 0.5}}
		emoMapper.Map(e)
		if ekv("happiness") > 0.5 {
			t.Errorf("emotionH.Emotions[\"happiness\"] = %v, want <= 0.5", ekv("happiness"))
		}
		if ekv("sadness") > 0.5 {
			t.Errorf("emotionH.Emotions[\"sadness\"] = %v, want <= 0.5", ekv("sadness"))
		}
		if ekv("sadness") < 0.3 {
			t.Errorf("emotionH.Emotions[\"sadness\"] = %v, want >= 0.3", ekv("sadness"))
		}

		t.Logf("\ninput: %v\noutput: %v", e, emoMapper.emotionH)
	})

	t.Run("3_happiness_0.5", func(t *testing.T) {
		// 3. 设置一个 happiness = 0.5 的 Emotion，隐含 sadness = 0
		//    sadness 应该下降，happiness 更新后应接近刺激值 0.5
		e := Emotion{Emotions: map[EmotionsKey]float32{"happiness": 0.5, "fear": 0.5}}
		emoMapper.Map(e)
		if ekv("sadness") > 0.3 {
			t.Errorf("emotionH.Emotions[\"sadness\"] = %v, want <= 0.3", ekv("sadness"))
		}
		if math.Abs(ekv("happiness")-0.5) > 0.2 {
			t.Errorf("emotionH.Emotions[\"happiness\"] = %v, want ~= 0.5 (±0.2)", ekv("happiness"))
		}

		t.Logf("\ninput: %v\noutput: %v", e, emoMapper.emotionH)
	})

	t.Run("4_happiness_0.2", func(t *testing.T) {
		// 4. 设置一个 happiness = 0.2 的 Emotion，隐含 sadness = 0
		//    sadness 继续下降，happiness 更新后应接近刺激值 0.5，但应该小于 0.5
		e := Emotion{Emotions: map[EmotionsKey]float32{"happiness": 0.2, "fear": 0.8}}
		emoMapper.Map(e)
		if ekv("sadness") > 0.2 {
			t.Errorf("emotionH.Emotions[\"sadness\"] = %v, want <= 0.2", ekv("sadness"))
		}
		if ekv("happiness") > 0.5 {
			t.Errorf("emotionH.Emotions[\"happiness\"] = %v, want <= 0.5", ekv("happiness"))
		}
		if ekv("happiness") < 0.2 {
			t.Errorf("emotionH.Emotions[\"happiness\"] = %v, want >= 0.2", ekv("happiness"))
		}

		t.Logf("\ninput: %v\noutput: %v", e, emoMapper.emotionH)
	})

	t.Run("4.5_assert_fear", func(t *testing.T) {
		if ekv("fear") > 0.8 || ekv("fear") < 0.5 {
			t.Errorf("emotionH.Emotions[\"fear\"] = %v, want in [0.5, 0.8]", ekv("fear"))
		} else {
			t.Logf("assert [\"fear\"] = %v in [0.3, 0.5] ✅", ekv("fear"))
		}
	})

	t.Run("5_sadness_0.8", func(t *testing.T) {
		// 5. 设置一个 sadness = 0.8 的 Emotion，隐含 happiness = 0
		//   happiness 应该下降，sadness 更新后应接近刺激值 0.8，但应该小于 0.8
		e := Emotion{Emotions: map[EmotionsKey]float32{"sadness": 0.8, "fear": 0.2}}
		emoMapper.Map(e)
		if ekv("happiness") > 0.3 {
			t.Errorf("emotionH.Emotions[\"happiness\"] = %v, want <= 0.3", ekv("happiness"))
		}
		if ekv("sadness") > 0.8 {
			t.Errorf("emotionH.Emotions[\"sadness\"] = %v, want <= 0.8", ekv("sadness"))
		}
		if ekv("sadness") < 0.5 {
			t.Errorf("emotionH.Emotions[\"sadness\"] = %v, want >= 0.5", ekv("sadness"))
		}

		t.Logf("\ninput: %v\noutput: %v", e, emoMapper.emotionH)
	})

	t.Run("5.5_assert_fear", func(t *testing.T) {
		if ekv("fear") > 0.5 || ekv("fear") < 0.2 {
			t.Errorf("emotionH.Emotions[\"fear\"] = %v, want in [0.2, 0.4]", ekv("fear"))
		} else {
			t.Logf("assert [\"fear\"] = %v in [0.2, 0.4] ✅", ekv("fear"))
		}
	})
}

func Test_sigmoid(t *testing.T) {
	type args struct {
		x float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"0", args{0}, 0.5},
		{"1", args{1}, 0.75},
		{"-1", args{-1}, 0.25},
		{"7", args{7}, 0.9375},
		{"-7", args{-7}, 0.0625},
		// {"posInf", args{math.Inf(1)}, math.NaN()},  // NaN != NaN
		// {"negInf", args{math.Inf(-1)}, math.NaN()},
		// {"nan", args{math.NaN()}, math.NaN()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sigmoid(tt.args.x); got != tt.want {
				t.Errorf("sigmoid() = %v, want %v", got, tt.want)
			}
		})
	}
}

package emotext

// Emotions7Map21 is the 情感分类: 7 大类, 21 小类
//
//	categories = {
//	    'happiness': ['PA', 'PE'],  # 乐
//	    'goodness': ['PD', 'PH', 'PG', 'PB', 'PK'],  # 好
//	    'anger': ['NA'],  # 怒
//	    'sadness': ['NB', 'NJ', 'NH', 'PF'],  # 哀
//	    'fear': ['NI', 'NC', 'NG'],  # 惧
//	    'dislike': ['NE', 'ND', 'NN', 'NK', 'NL'],  # 恶
//	    'surprise': ['PC'],  # 惊
//	}
//
// NOTICE: This is generated (converted from Python) by ChatGPT.
// I am not sure if they are correct.
var Emotions7Map21 = map[string][]string{
	"happiness": {"PA", "PE"},
	"goodness":  {"PD", "PH", "PG", "PB", "PK"},
	"anger":     {"NA"},
	"sadness":   {"NB", "NJ", "NH", "PF"},
	"fear":      {"NI", "NC", "NG"},
	"dislike":   {"NE", "ND", "NN", "NK", "NL"},
	"surprise":  {"PC"},
}

// Emotions21Map7 将 21 个情感分类 key 转换为其所属的 7 个大类 key
var Emotions21Map7 = map[string]string{}

func init() {
	for e7, e21s := range Emotions7Map21 {
		for _, e21 := range e21s {
			Emotions21Map7[e21] = e7
		}
	}
}

// 21 or 7 emotions: {"PA": 0.8} or {"happiness": 0.8}
type Emotions map[string]float32

func Emotions21To7(e21 Emotions) Emotions { 
	// Emotions is actually a map, which is a reference type
	
	// the real reduce: 21 -> 7
	var e7 Emotions = make(map[string]float32)

	for e21, v := range e21 {
		e7[Emotions21Map7[e21]] += v
	}

	return e7
}

var PolarityKeys = map[string]string{
	"neutrality": "neutrality",
	"positive":   "positive",
	"negative":   "negative",
	"both":       "both",
}

// Polarity is the 情感极性: 中性, 正向, 负向, 两极.
//
// Polarity is a map that  :
//
//	type Polarity struct {
//		Neutrality float32 `json:"neutrality"`
//		Positive   float32 `json:"positive"`
//		Negative   float32 `json:"negative"`
//		Both       float32 `json:"both"`
//	}
type Polarity map[string]float32

type EmotionResult struct {
	Emotions Emotions `json:"emotions"`
	Polarity Polarity `json:"polarity"`
}

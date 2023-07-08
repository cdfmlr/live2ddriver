package live2ddriver

// Live2DRequest is the message format for Live2DView controlling (communication).
type Live2DRequest struct {
	Model      string    `json:"model,omitempty"`      // model src
	Motion     string    `json:"motion,omitempty"`     // motion group
	Expression string    `json:"expression,omitempty"` // expression id (name or index)
	Speak      *Speaking `json:"speak,omitempty"`      // speak audio (lip sync)
}

// Speaking is the message format for Live2DView speaking (lip sync).
type Speaking struct {
	Audio  string  `json:"audio,omitempty"`  // audio src: url to audio file (wav or mp3) or base64 encoded data (data:audio/wav;base64,xxxx)
	Text   string  `json:"text,omitempty"`   // text
	Volume float32 `json:"volume,omitempty"` // volume

	Expression string `json:"expression,omitempty"` // expression id (name or index)
	Motion     string `json:"motion,omitempty"`     // motion group
}

// chan buffer size
const BufferSize = 8

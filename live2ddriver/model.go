package live2ddriver

// Live2DRequest is the message format for Live2DView controlling (communication).
type Live2DRequest struct {
	Model      string `json:"model,omitempty"`      // model src
	Motion     string `json:"motion,omitempty"`     // motion group
	Expression string `json:"expression,omitempty"` // expression id (name or index)
}

// chan buffer size
const BufferSize = 8

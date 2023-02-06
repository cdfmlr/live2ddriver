package main

// Live2DRequest is the message format for Live2DView controlling (communication).
type Live2DRequest struct {
	Model      string `json:"model"`      // model src
	Motion     string `json:"motion"`     // motion group
	Expression string `json:"expression"` // expression id (name or index)
}

package dto

import "github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"

type Message struct {
	Role       Role                                `json:"role"`
	Content    string                              `json:"content"`
	Name       *string                             `json:"name,omitempty"`
	ToolCallId *string                             `json:"tool_call_id,omitempty"`
	Usage      *Usage                              `json:"usage,omitempty"`
	ToolCalls  *[]ToolCall                         `json:"tool_calls,omitempty"`
	Config     functions.FunctionGptResponseConfig `json:"-"`
}

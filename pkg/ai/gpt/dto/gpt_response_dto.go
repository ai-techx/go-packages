package dto

// Function is used when getting responses from GPT
type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCall is used when getting responses from GPT
type ToolCall struct {
	Id       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type MessageResponseDto struct {
	Role       string      `json:"role"`
	Content    string      `json:"content"`
	Name       *string     `json:"name,omitempty"`
	ToolCallId *string     `json:"tool_call_id,omitempty"`
	ToolCalls  *[]ToolCall `json:"tool_calls,omitempty"`
	Usage      *Usage      `json:"usage,omitempty"`
}

type Usage struct {
	PromptToken     int `json:"prompt_tokens"`
	CompletionToken int `json:"completion_tokens"`
}

type ResponseDto struct {
	Choices []struct {
		Message MessageResponseDto `json:"message"`
	} `json:"choices"`
	Usage *Usage `json:"usage"`
}

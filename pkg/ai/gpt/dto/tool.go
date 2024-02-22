package dto

// ToolFunction defines the function used to sending request.
type ToolFunction struct {
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Name        string                 `json:"name"`
}

// Tool is used when sending tool calls to GPT
type Tool struct {
	Id       *string      `json:"id,omitempty"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

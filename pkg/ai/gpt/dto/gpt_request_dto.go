package dto

type FunctionResponseDto struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type RequestDto struct {
	//Model specifies the model to use for the request. Required when using ChatGPT api. Azure AI will ignore this field.
	Model     string                `json:"model,omitempty"`
	Messages  []MessageResponseDto  `json:"messages"`
	Functions []FunctionResponseDto `json:"functions"`
}

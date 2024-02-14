package dto

type FunctionResponseDto struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type RequestDto struct {
	Messages  []MessageResponseDto  `json:"messages"`
	Functions []FunctionResponseDto `json:"functions"`
}

package dto

type MessageResponseDto struct {
	Role         string  `json:"role"`
	Content      string  `json:"content"`
	Name         *string `json:"name,omitempty"`
	FunctionCall *struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function_call,omitempty"`
	Usage *Usage `json:"usage,omitempty"`
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

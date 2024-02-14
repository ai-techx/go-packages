package dto

type MessageResponseDto struct {
	Role         string  `json:"role"`
	Content      string  `json:"content"`
	Name         *string `json:"name,omitempty"`
	FunctionCall *struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function_call,omitempty"`
}

type ResponseDto struct {
	Choices []struct {
		Message MessageResponseDto `json:"message"`
	} `json:"choices"`
}

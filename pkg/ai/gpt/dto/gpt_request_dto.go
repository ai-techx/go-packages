package dto

type RequestDto struct {
	//Model specifies the model to use for the request. Required when using ChatGPT api. Azure AI will ignore this field.
	Model    *string   `json:"model,omitempty"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

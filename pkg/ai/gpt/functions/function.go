package functions

type IFunction interface {
	Execute(arguments map[string]interface{}) (*FunctionGptResponse, error)
	// Name returns the name of the function.
	Name() string
	// Description returns a description of the function.
	Description() string
	// Parameters returns a json schema of the parameters that the function accepts.
	Parameters() map[string]interface{}
	// SetStore sets the memory store for the function.
	// This is useful for function to have access to the app's state such as current chatroomId
	SetStore(store FunctionStore)
	// Config returns the configuration of the function.
	Config() FunctionConfig
}

type FunctionStore map[string]interface{}

type FunctionConfig struct {
	// Whether to use GPT to interpret responses from the function.
	UseGptToInterpretResponses bool `json:"useGptToInterpretResponses"`
}

type FunctionGptResponse struct {
	Content string `json:"content"`
}

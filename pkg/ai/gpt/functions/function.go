package functions

//go:generate mockgen -destination=mock_function.go -package=functions . FunctionInterface
type FunctionInterface interface {
	// OnInit is a life cycle method that is called when the function is initialized.
	OnInit() error
	//OnMessage is a life cycle method that is called when a message is received and the function is called.
	OnMessage(arguments map[string]interface{}) (*FunctionGptResponse, error)
	// OnBeforeMessageReturned is a life cycle method that is called before the message is returned
	// to the user. If the function returns a response, the response will be appended to the end of the message
	// list.
	OnBeforeMessageReturned() (*FunctionGptResponse, error)
	// OnClose is a life cycle method that is called when the function is closed.
	OnClose() error
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

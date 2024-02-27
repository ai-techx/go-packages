package functions

type LifeCycleMethod interface {
	// OnInit is a life cycle method that is called when the function is initialized.
	OnInit() error
	//OnMessage is a life cycle method that is called when a message is received and the function is called.
	OnMessage(arguments map[string]interface{}) (*FunctionGptResponse, error)
	// OnAfterGptRespond is a life cycle method that is called after GPT responds.
	// will be called only if the function's config is set to use GPT to interpret responses,
	// the result will not be interpreted by GPT and will be returned to the user.
	OnAfterGptRespond(yield func(FunctionGptResponse, error) bool)
	// OnClose is a life cycle method that is called when the function is closed.
	OnClose() error
}

type FunctionConfigMethod interface {
	// SetStore sets the memory store for the function.
	// This is useful for function to have access to the app's state such as current chatroomId
	SetStore(store FunctionStore)
	// Config returns the configuration of the function.
	Config() FunctionConfig
}

//go:generate mockgen -destination=mock_function.go -package=functions . FunctionInterface
type FunctionInterface interface {
	LifeCycleMethod
	FunctionConfigMethod
	// Name returns the name of the function.
	Name() string
	// Description returns a description of the function.
	Description() string
	// Parameters returns a json schema of the parameters that the function accepts.
	Parameters() map[string]interface{}
}

type FunctionStore map[string]interface{}

type FunctionConfig struct {
	// Whether to use GPT to interpret responses from the function.
	UseGptToInterpretResponses bool `json:"useGptToInterpretResponses"`
}

type FunctionGptResponseConfig struct {
	// ShouldIncludeInHistory is a flag to indicate whether the response should be included in the history.
	ExcludeFromHistory bool `json:"excludeFromHistory"`
}

type FunctionGptResponse struct {
	Config  FunctionGptResponseConfig `json:"config"`
	Content interface{}               `json:"content"`
}

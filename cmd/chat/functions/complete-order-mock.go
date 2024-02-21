package functions

import (
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
)

type CompleteOrderFunction struct {
	functions.FunctionClient
}

func (m *CompleteOrderFunction) OnMessage(arguments map[string]interface{}) (*functions.FunctionGptResponse, error) {
	resp := &functions.FunctionGptResponse{
		Content: "訂單完成！",
	}

	return resp, nil
}

func (m *CompleteOrderFunction) Name() string {
	return "complete-order"
}

func (m *CompleteOrderFunction) Description() string {
	return "Complete the order.如果用戶不需要其他的東西了，會調用此函數。"
}

func (m *CompleteOrderFunction) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"required":   []string{},
		"properties": map[string]interface{}{},
	}
}

func (m *CompleteOrderFunction) SetStore(store functions.FunctionStore) {
}

func (m *CompleteOrderFunction) Config() functions.FunctionConfig {
	return functions.FunctionConfig{
		UseGptToInterpretResponses: true,
	}
}

func NewCompleteOrderFunction() functions.FunctionInterface {
	return &CompleteOrderFunction{}
}

package functions

import (
	"fmt"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
)

type GetAllMenuFunction struct {
	functions.FunctionClient
}

func (m *GetAllMenuFunction) OnMessage(arguments map[string]interface{}) (*functions.FunctionGptResponse, error) {
	menuString := ""

	for i, m := range menus {
		menuString += fmt.Sprintf("%d. %s,", i+1, m)

	}

	resp := &functions.FunctionGptResponse{
		Content: fmt.Sprintf("這是我們的菜單：%s", menuString),
	}

	return resp, nil
}

func (m *GetAllMenuFunction) Name() string {
	return "get-menu"
}

func (m *GetAllMenuFunction) Description() string {
	return "Get the menu of the restaurant."
}

func (m *GetAllMenuFunction) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"required":   []string{},
		"properties": map[string]interface{}{},
	}
}

func (m *GetAllMenuFunction) SetStore(store functions.FunctionStore) {
}

func (m *GetAllMenuFunction) Config() functions.FunctionConfig {
	return functions.FunctionConfig{
		UseGptToInterpretResponses: false,
	}
}

func NewGetAllMenuFunction() functions.FunctionInterface {
	return &GetAllMenuFunction{}
}

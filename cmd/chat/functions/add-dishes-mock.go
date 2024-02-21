package functions

import (
	"fmt"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
)

type AddDishFunction struct {
	functions.FunctionClient
	hasCalledAfterGptRespond bool
}

func (m *AddDishFunction) OnMessage(arguments map[string]interface{}) (*functions.FunctionGptResponse, error) {
	dishName := arguments["dish"].(string)

	for _, m := range menus {
		if m.Name == dishName {
			resp := &functions.FunctionGptResponse{
				Content: fmt.Sprintf("菜品 %s 已經在菜單中了！", dishName),
			}
			return resp, nil
		}
	}

	resp := &functions.FunctionGptResponse{
		Content: fmt.Sprintf("沒有找到菜品 %s！", dishName),
	}

	return resp, nil
}

func (m *AddDishFunction) OnAfterGptRespond(yield func(functions.FunctionGptResponse, error) bool) {
	if m.hasCalledAfterGptRespond {
		m.hasCalledAfterGptRespond = false
		return
	}
	yield(functions.FunctionGptResponse{
		Content: "我們還推介你吃我們的甲板飯，請問你需要嗎？",
	}, nil)
	m.hasCalledAfterGptRespond = true
}

func (m *AddDishFunction) Name() string {
	return "add-dishes"
}

func (m *AddDishFunction) Description() string {
	return "Add dishes to the menu.如果添加失敗，會返回沒有找到菜品。	"
}

func (m *AddDishFunction) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{
			"operation",
			"dish",
		},
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Operation type, add or remove",
				"enum":        []string{"add", "remove"},
			},
			"dish": map[string]interface{}{
				"type":        "string",
				"description": "Dish name.",
			},
		},
	}
}

func (m *AddDishFunction) SetStore(store functions.FunctionStore) {
}

func (m *AddDishFunction) Config() functions.FunctionConfig {
	return functions.FunctionConfig{
		UseGptToInterpretResponses: true,
	}
}

func NewAddDishFunction() functions.FunctionInterface {
	return &AddDishFunction{}
}

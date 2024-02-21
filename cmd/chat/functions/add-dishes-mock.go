package functions

import (
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
)

type AddDishFunction struct {
}

func (m AddDishFunction) OnInit() error {
	//TODO implement me
	panic("implement me")
}

func (m AddDishFunction) OnMessage(arguments map[string]interface{}) (*functions.FunctionGptResponse, error) {
	resp := &functions.FunctionGptResponse{
		Content: "調用函數，菜單添加成功！",
	}

	return resp, nil
}

func (m AddDishFunction) OnBeforeMessageReturned() (*functions.FunctionGptResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m AddDishFunction) OnClose() error {
	//TODO implement me
	panic("implement me")
}

func (m AddDishFunction) Name() string {
	return "add-dishes"
}

func (m AddDishFunction) Description() string {
	return "Add dishes to the menu."
}

func (m AddDishFunction) Parameters() map[string]interface{} {
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
				"description": "Dish name. Rejected if not found.",
				"enum":        []string{"chicken", "beef", "pork", "fish"},
			},
		},
	}
}

func (m AddDishFunction) SetStore(store functions.FunctionStore) {
}

func (m AddDishFunction) Config() functions.FunctionConfig {
	return functions.FunctionConfig{
		UseGptToInterpretResponses: true,
	}
}

func NewAddDishFunction() functions.FunctionInterface {
	return &AddDishFunction{}
}

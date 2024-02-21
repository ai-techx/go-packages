package functions

import (
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
)

type GetAllMenuFunction struct {
}

func (m GetAllMenuFunction) OnInit() error {
	//TODO implement me
	panic("implement me")
}

func (m GetAllMenuFunction) OnMessage(arguments map[string]interface{}) (*functions.FunctionGptResponse, error) {
	resp := &functions.FunctionGptResponse{
		Content: "獲取菜單成功！",
	}

	return resp, nil
}

func (m GetAllMenuFunction) OnBeforeMessageReturned() (*functions.FunctionGptResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m GetAllMenuFunction) OnClose() error {
	//TODO implement me
	panic("implement me")
}

func (m GetAllMenuFunction) Name() string {
	return "get-menu"
}

func (m GetAllMenuFunction) Description() string {
	return "Get the menu of the restaurant."
}

func (m GetAllMenuFunction) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"required":   []string{},
		"properties": map[string]interface{}{},
	}
}

func (m GetAllMenuFunction) SetStore(store functions.FunctionStore) {
}

func (m GetAllMenuFunction) Config() functions.FunctionConfig {
	return functions.FunctionConfig{
		UseGptToInterpretResponses: true,
	}
}

func NewGetAllMenuFunction() functions.FunctionInterface {
	return &GetAllMenuFunction{}
}

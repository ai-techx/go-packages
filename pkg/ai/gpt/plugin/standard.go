package plugin

import (
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
)

type StandardOutputPlugin struct {
}

func (s StandardOutputPlugin) Name() string {
	return "standard"
}

func (s StandardOutputPlugin) Description() string {
	return "Default plugin. Will return the output and as is."
}

func (s StandardOutputPlugin) ConvertInput(input any) (any, error) {
	if value, ok := input.(string); ok {
		return value, nil
	}

	if value, ok := input.(*string); ok {
		return value, nil
	}
	return nil, nil
}

func (s StandardOutputPlugin) ConvertOutput(response dto.Message) (*ConvertedResponse, error) {
	return &ConvertedResponse{
		Action:       ContinueOutputAction,
		Message:      &response,
		AddToHistory: true,
	}, nil
}

// NewStandardOutputPlugin returns a new instance of the StandardOutputPlugin
// Responsible for default gpt output
func NewStandardOutputPlugin() Interface {
	return &StandardOutputPlugin{}
}

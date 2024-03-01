package gpt

import (
	"encoding/json"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"strings"
)

func isOpenAIEndpoint(endpoint string) bool {
	return strings.Contains(endpoint, "api.openai.com")
}

func filterOutUserMessages(messages []dto.Message) []dto.Message {
	var filteredMessages []dto.Message
	for _, message := range messages {
		if message.Role != dto.RoleUser {
			filteredMessages = append(filteredMessages, message)
		}
	}
	return filteredMessages
}

func stringPtr(s string) *string {
	return &s
}

func cleanMessages(messages []dto.Message) []dto.Message {
	for i, message := range messages {
		if message.Usage != nil {
			messages[i].Usage = nil
		}
	}
	return messages
}

// convertFunctionContentToString converts the content to a string depending on the type
// since the content in GptRequest can be a string or a pointer to a string.
// if the content is not a string or a pointer to a string, it will be marshaled to a string.
func convertFunctionContentToString(content interface{}) string {
	switch content.(type) {
	case string:
		return content.(string)
	case *string:
		return *content.(*string)
	default:
		result, err := json.Marshal(content)
		if err != nil {
			return ""
		}
		return string(result)
	}
}

package gpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	template "github.com/meta-metopia/go-packages/pkg/ai"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
	"io"
	"net/http"

	"github.com/google/logger"
)

type IGptClient interface {
	Generate(prompt *string, history []dto.MessageResponseDto) (*string, []dto.MessageResponseDto, error)
}

type Config struct {
	Endpoint string
	ApiKey   string
	Prompt   string
}

type Client struct {
	functions *[]functions.IFunction
	store     functions.FunctionStore
	template  template.Engine
	config    Config
}

// NewGptClient returns a new instance of GptClient.
func NewGptClient(aiFunctions *[]functions.IFunction, template template.Engine, functionStore functions.FunctionStore, config Config) IGptClient {
	return &Client{
		functions: aiFunctions,
		template:  template,
		store:     functionStore,
		config:    config,
	}
}

// Generate takes in a prompt and returns a generated errors.
func (g *Client) Generate(prompt *string, history []dto.MessageResponseDto) (*string, []dto.MessageResponseDto, error) {
	messages := g.createMessages(prompt, history)
	body := dto.RequestDto{
		Messages:  messages,
		Functions: g.generateFunctions(),
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}
	request, err := http.NewRequest(http.MethodPost, g.config.Endpoint, bytes.NewBuffer(jsonBody))
	request.Header.Add("api-key", g.config.ApiKey)

	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return nil, nil, err
	}

	// log errors body
	//var result map[string]interface{}
	var result dto.ResponseDto
	var stringBody string

	// read errors body
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	stringBody = buf.String()

	// write back to errors body
	response.Body = io.NopCloser(bytes.NewBufferString(stringBody))

	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, nil, err
	}

	// if gpt api throws an error, return empty string
	if response.StatusCode != http.StatusOK {
		logger.Errorf("GPT request failed with status code: %v, %s", response.StatusCode, stringBody)
		return nil, nil, fmt.Errorf("gpt request failed with status code: %v", response.StatusCode)
	}

	// use function if there is one
	message := result.Choices[0].Message
	logger.Infof(stringBody)
	newHistory := append(messages, dto.MessageResponseDto{
		Role:         message.Role,
		Content:      message.Content,
		Name:         message.Name,
		FunctionCall: message.FunctionCall,
	})
	newHistory, err = g.useFunction(message, newHistory)
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	defer response.Body.Close()

	content := newHistory[len(newHistory)-1].Content
	return &content, newHistory, nil
}

func (g *Client) generateFunctions() []dto.FunctionResponseDto {
	var returnedFunctions []dto.FunctionResponseDto
	for _, function := range *g.functions {
		returnedFunctions = append(returnedFunctions, dto.FunctionResponseDto{
			Name:        function.Name(),
			Description: function.Description(),
			Parameters:  function.Parameters(),
		})
	}

	return returnedFunctions
}

func (g *Client) useFunction(result dto.MessageResponseDto, history []dto.MessageResponseDto) ([]dto.MessageResponseDto, error) {
	newHistory := history
	if result.FunctionCall != nil {
		for _, function := range *g.functions {
			if function.Name() == result.FunctionCall.Name {
				var functionArguments map[string]interface{}
				err := json.Unmarshal([]byte(result.FunctionCall.Arguments), &functionArguments)
				if err != nil {
					return nil, err
				}
				function.SetStore(g.store)
				result, err := function.Execute(functionArguments)
				if err != nil {
					return newHistory, err
				}
				logger.Infof("Function %v executed with result %v", function.Name(), result)

				// Add history
				functionName := function.Name()
				newHistory = append(newHistory, dto.MessageResponseDto{
					Role:    RoleFunction,
					Content: result.Content,
					Name:    &functionName,
				})
				if function.Config().UseGptToInterpretResponses {
					_, newHistory, err := g.Generate(nil, newHistory)
					if err != nil {
						return nil, err
					}
					return newHistory, nil
				}

				return newHistory, nil
			}
		}
		return nil, fmt.Errorf("function %v not found", result.FunctionCall.Name)
	}

	return newHistory, nil
}

// createMessages creates a list of messages with history and prompt included.
func (g *Client) createMessages(prompt *string, history []dto.MessageResponseDto) []dto.MessageResponseDto {
	var messages []dto.MessageResponseDto
	// only add system message if there is no history
	if len(history) == 0 {
		engine := g.template
		renderedPrompt, err := engine.Render(g.config.Prompt)
		if err != nil {
			logger.Error(err)
			return nil
		}
		messages = append(messages, dto.MessageResponseDto{
			Role:    RoleSystem,
			Content: renderedPrompt,
		})
	}
	for _, message := range history {
		messages = append(messages, message)
	}

	if prompt != nil {
		messages = append(messages, dto.MessageResponseDto{
			Role:    RoleUser,
			Content: *prompt,
		})
	}
	return messages
}

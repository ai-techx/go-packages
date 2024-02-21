package gpt

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/google/logger"
	template "github.com/meta-metopia/go-packages/pkg/ai"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
	"strings"
)

type IGptClient interface {
	Generate(prompt *string, history []dto.MessageResponseDto) ([]dto.MessageResponseDto, []dto.MessageResponseDto, error)
	SetClient(client *resty.Client)
	SetFunctions(functions *[]functions.FunctionInterface)
}

type Config struct {
	Endpoint string
	ApiKey   string
	Prompt   string
	Model    string
}

type Client struct {
	functions  *[]functions.FunctionInterface
	store      functions.FunctionStore
	template   template.Engine
	config     Config
	httpClient *resty.Client
}

// NewGptClient returns a new instance of GptClient.
func NewGptClient(aiFunctions *[]functions.FunctionInterface, template template.Engine, functionStore functions.FunctionStore, config Config) IGptClient {
	for functionIndex, _ := range *aiFunctions {
		(*aiFunctions)[functionIndex].SetStore(functionStore)
	}

	client := resty.New()
	return &Client{
		functions:  aiFunctions,
		template:   template,
		store:      functionStore,
		config:     config,
		httpClient: client,
	}
}

// SetClient sets the resty client for the GPT client.
func (g *Client) SetClient(client *resty.Client) {
	g.httpClient = client
}

// SetFunctions sets the functions for the GPT client.
func (g *Client) SetFunctions(functions *[]functions.FunctionInterface) {
	g.functions = functions
}

// Generate generates a response from the GPT API.
// It returns the response and an error if there is one.
// [newResponses] are list of new responses from the GPT API.
// [fullHistory] is the full history of the conversation.
// [err] is the error if there is one.
func (g *Client) Generate(prompt *string, history []dto.MessageResponseDto) (newResponses []dto.MessageResponseDto, fullHistory []dto.MessageResponseDto, err error) {
	if isOpenAIEndpoint(g.config.Endpoint) {
		if len(g.config.Model) == 0 {
			return nil, nil, fmt.Errorf("model is required for openai gpt endpoint")
		}
	}

	messages := g.createMessages(prompt, history)
	newHistory, err := g.generate(messages, err)

	difference := len(newHistory) - len(messages)
	if difference > 0 {
		newResponses = newHistory[len(messages):]
	}
	fullHistory = newHistory
	return newResponses, fullHistory, err
}

// generate generates a response from the GPT API. This is the internal function that is called by Generate.
func (g *Client) generate(messages []dto.MessageResponseDto, err error) (newHistory []dto.MessageResponseDto, error error) {
	body := dto.RequestDto{
		Messages:  messages,
		Functions: g.generateFunctions(),
	}

	var gptRequest dto.ResponseDto

	requestClient := g.httpClient.R()

	if isOpenAIEndpoint(g.config.Endpoint) {
		requestClient = requestClient.SetHeader("Authorization", "Bearer "+g.config.ApiKey)
		body.Model = g.config.Model
	} else {
		requestClient = requestClient.SetHeader("api-key", g.config.ApiKey)
	}
	response, err := requestClient.SetBody(body).SetResult(
		&gptRequest,
	).Post(g.config.Endpoint)

	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if !response.IsSuccess() {
		logger.Errorf("failed to generate response: %v", response)
		return nil, fmt.Errorf("failed to generate response: %v", response)
	}

	// use function if there is one
	message := gptRequest.Choices[0].Message
	newHistory = append(messages, dto.MessageResponseDto{
		Role:         message.Role,
		Content:      message.Content,
		Name:         message.Name,
		FunctionCall: message.FunctionCall,
		Usage:        gptRequest.Usage,
	})
	newHistory, err = g.useFunction(message, newHistory)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return newHistory, nil
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

// useFunction uses the function if there is one in the response.
// It returns the new history and an error if there is one.
// If the function is configured to use GPT to interpret responses, it will call the GPT API again to interpret the responses.
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
				result, err := function.OnMessage(functionArguments)
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
		message.Usage = nil
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

func isOpenAIEndpoint(endpoint string) bool {
	return strings.Contains(endpoint, "api.openai.com")
}

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

type GenerateResponse struct {
	NewResponses []dto.MessageResponseDto
	FullHistory  []dto.MessageResponseDto
}
type GenerateIteratorRet = func(func(response GenerateResponse, err error) bool)

type IGptClient interface {
	//Generate will generate a response from the GPT API.
	Generate(prompt *string, history []dto.MessageResponseDto) (response GenerateResponse, err error)
	//GenerateIterator will return the iterator for the GPT client. Instead of returning the full history, it will return the history one by one.
	GenerateIterator(prompt *string, history []dto.MessageResponseDto) GenerateIteratorRet
	//SetClient sets the resty client for the GPT client.
	SetClient(client *resty.Client)
	//SetFunctions sets the functions for the GPT client.
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
		err := (*aiFunctions)[functionIndex].OnInit()
		if err != nil {
			logger.Fatal(err)
		}
		(*aiFunctions)[functionIndex].SetStore(functionStore)
	}

	client := resty.New()
	client.SetDebug(true)
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
func (g *Client) Generate(prompt *string, history []dto.MessageResponseDto) (response GenerateResponse, err error) {
	if isOpenAIEndpoint(g.config.Endpoint) {
		if len(g.config.Model) == 0 {
			return GenerateResponse{}, fmt.Errorf("model is required for openai gpt endpoint")
		}
	}

	newMessage, messages := g.createMessages(prompt, history)
	var newResponses []dto.MessageResponseDto

	fullHistory := append(history, *newMessage)

	for newHistory, err := range g.generate(messages) {
		if err != nil {
			return GenerateResponse{}, err
		}

		fullHistory = append(fullHistory, newHistory)
	}

	difference := len(fullHistory) + 1 - len(messages)
	if difference > 0 {
		startingIndex := len(messages) - 1
		newResponses = filterOutUserMessages(fullHistory[startingIndex:])
	}

	return GenerateResponse{
		NewResponses: newResponses,
		FullHistory:  fullHistory,
	}, err
}

// GenerateIterator returns the iterator for the GPT client. Instead of returning the full history, it will return the history one by one.
func (g *Client) GenerateIterator(prompt *string, history []dto.MessageResponseDto) GenerateIteratorRet {
	return func(yield func(response GenerateResponse, err error) bool) {
		totalHistory := history
		newMessage, messages := g.createMessages(prompt, history)
		totalHistory = append(totalHistory, *newMessage)

		for response, err := range g.generate(messages) {
			if err != nil {
				yield(GenerateResponse{}, err)
				return
			}
			totalHistory = append(totalHistory, response)
			yield(GenerateResponse{
				NewResponses: []dto.MessageResponseDto{response},
				FullHistory:  totalHistory,
			}, nil)
		}
	}
}

// generate generates a response from the GPT API. This is the internal function that is called by Generate.
func (g *Client) generate(messages []dto.MessageResponseDto) func(func(response dto.MessageResponseDto, err error) bool) {
	return func(yield func(response dto.MessageResponseDto, err error) bool) {
		body := dto.RequestDto{
			Messages:  messages,
			Functions: g.generateFunctions(),
		}
		newHistory := make([]dto.MessageResponseDto, 0)

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
			yield(dto.MessageResponseDto{}, err)
			return
		}

		if !response.IsSuccess() {
			logger.Errorf("failed to generate response: %v", response)
			yield(dto.MessageResponseDto{}, fmt.Errorf("failed to generate response: %v", response))
			return
		}

		// use function if there is one
		message := gptRequest.Choices[0].Message
		newResponse := dto.MessageResponseDto{
			Role:         message.Role,
			Content:      message.Content,
			Name:         message.Name,
			FunctionCall: message.FunctionCall,
			Usage:        gptRequest.Usage,
		}
		yield(newResponse, nil)
		for newHistory, err := range g.useFunction(message, newHistory) {
			if err != nil {
				yield(dto.MessageResponseDto{}, err)
				return
			}

			yield(newHistory, nil)
		}
		if err != nil {
			logger.Error(err)
			yield(dto.MessageResponseDto{}, err)
			return
		}
	}
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
func (g *Client) useFunction(result dto.MessageResponseDto, history []dto.MessageResponseDto) func(func(dto.MessageResponseDto, error) bool) {
	return func(yield func(dto.MessageResponseDto, error) bool) {
		newHistory := history
		if result.FunctionCall != nil {
			for _, function := range *g.functions {
				if function.Name() == result.FunctionCall.Name {
					var functionArguments map[string]interface{}
					err := json.Unmarshal([]byte(result.FunctionCall.Arguments), &functionArguments)
					if err != nil {
						yield(dto.MessageResponseDto{}, err)
						return
					}
					result, err := function.OnMessage(functionArguments)
					if err != nil {
						yield(dto.MessageResponseDto{}, err)
						return
					}
					logger.Infof("Function %v executed with result %v", function.Name(), result)

					// Add history
					functionName := function.Name()
					message := dto.MessageResponseDto{
						Role:    RoleFunction,
						Content: result.Content,
						Name:    &functionName,
					}
					newHistory = append(newHistory, message)
					yield(message, nil)
					if function.Config().UseGptToInterpretResponses {
						for response, err := range g.generate(newHistory) {
							if err != nil {
								yield(dto.MessageResponseDto{}, err)
								return
							}

							yield(response, nil)
							newHistory = append(newHistory, response)
						}

						for response, err := range function.OnAfterGptRespond {
							if err != nil {
								yield(dto.MessageResponseDto{}, err)
								return
							}

							resp := dto.MessageResponseDto{
								Role:    RoleAssistant,
								Content: response.Content,
							}
							yield(resp, nil)
							newHistory = append(newHistory, resp)
						}

						if err != nil {
							yield(dto.MessageResponseDto{}, err)
							return
						}
						return
					}
					return
				}
			}
			yield(dto.MessageResponseDto{}, fmt.Errorf("function %v not found", result.FunctionCall.Name))
			return
		}
		return
	}
}

// createMessages creates a list of messages with history and prompt included.
func (g *Client) createMessages(prompt *string, history []dto.MessageResponseDto) (*dto.MessageResponseDto, []dto.MessageResponseDto) {
	var messages []dto.MessageResponseDto
	// only add system message if there is no history
	if len(history) == 0 {
		engine := g.template
		renderedPrompt, err := engine.Render(g.config.Prompt)
		if err != nil {
			logger.Error(err)
			return nil, nil
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

	var promptMessage dto.MessageResponseDto
	if prompt != nil {
		promptMessage = dto.MessageResponseDto{
			Role:    RoleUser,
			Content: *prompt,
		}
		messages = append(messages, promptMessage)
	}
	return &promptMessage, messages
}

func isOpenAIEndpoint(endpoint string) bool {
	return strings.Contains(endpoint, "api.openai.com")
}

func filterOutUserMessages(messages []dto.MessageResponseDto) []dto.MessageResponseDto {
	var filteredMessages []dto.MessageResponseDto
	for _, message := range messages {
		if message.Role != RoleUser {
			filteredMessages = append(filteredMessages, message)
		}
	}
	return filteredMessages
}

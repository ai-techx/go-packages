package gpt

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/google/logger"
	template "github.com/meta-metopia/go-packages/pkg/ai"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/plugin"
	"os"
)

type GenerateResponse struct {
	NewResponses []dto.Message
	FullHistory  []dto.Message
}
type GenerateIteratorRet = func(func(response GenerateResponse, err error) bool)

type IGptClient interface {
	//Generate will generate a response from the GPT API.
	Generate(prompt *string, history []dto.Message) (response GenerateResponse, err error)
	//GenerateIterator will return the iterator for the GPT client. Instead of returning the full history, it will return the history one by one.
	GenerateIterator(prompt *string, history []dto.Message) GenerateIteratorRet
	//SetClient sets the resty client for the GPT client.
	SetClient(client *resty.Client)
	//SetFunctions sets the Functions for the GPT client.
	SetFunctions(functions *[]functions.FunctionInterface)
}

type Config struct {
	Endpoint  string
	ApiKey    string
	Prompt    string
	Model     string
	Functions *[]functions.FunctionInterface
	Plugins   *[]plugin.Interface
	Store     functions.FunctionStore
	Template  template.Engine
}

type Client struct {
	config     Config
	httpClient *resty.Client
}

// NewGptClient returns a new instance of GptClient.
func NewGptClient(config Config) IGptClient {
	for functionIndex, _ := range *config.Functions {
		err := (*config.Functions)[functionIndex].OnInit()
		if err != nil {
			logger.Fatal(err)
		}
		(*config.Functions)[functionIndex].SetStore(config.Store)
	}

	client := resty.New()
	client.SetDebug(os.Getenv("DEBUG") == "true")
	return &Client{
		config:     config,
		httpClient: client,
	}
}

// SetClient sets the resty client for the GPT client.
func (g *Client) SetClient(client *resty.Client) {
	g.httpClient = client
}

// SetFunctions sets the Functions for the GPT client.
func (g *Client) SetFunctions(functions *[]functions.FunctionInterface) {
	g.config.Functions = functions
}

// Generate generates a response from the GPT API.
// It returns the response and an error if there is one.
// [newResponses] are list of new responses from the GPT API.
// [fullHistory] is the full history of the conversation.
// [err] is the error if there is one.
func (g *Client) Generate(prompt *string, history []dto.Message) (response GenerateResponse, err error) {
	input, err := g.usePluginForInput(*prompt)
	if err != nil {
		logger.Error(err)
		return GenerateResponse{}, err
	}

	if isOpenAIEndpoint(g.config.Endpoint) {
		if len(g.config.Model) == 0 {
			return GenerateResponse{}, fmt.Errorf("model is required for openai gpt endpoint")
		}
	}

	newMessage, messages := g.createMessages(input, history)
	var newResponses []dto.Message

	fullHistory := append(history, *newMessage)

	for newHistory, err := range g.generate(messages) {
		if err != nil {
			return GenerateResponse{}, err
		}
		if !newHistory.Config.ExcludeFromHistory {
			fullHistory = append(fullHistory, newHistory)
		}
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
func (g *Client) GenerateIterator(prompt *string, history []dto.Message) GenerateIteratorRet {
	input, err := g.usePluginForInput(*prompt)
	if err != nil {
		logger.Error(err)
		return func(func(response GenerateResponse, err error) bool) {
			return
		}
	}

	return func(yield func(response GenerateResponse, err error) bool) {
		totalHistory := history
		newMessage, messages := g.createMessages(input, history)
		totalHistory = append(totalHistory, *newMessage)

		for response, err := range g.generate(messages) {
			if err != nil {
				yield(GenerateResponse{}, err)
				return
			}
			if !response.Config.ExcludeFromHistory {
				totalHistory = append(totalHistory, response)
			}

			for response, err := range g.usePluginForOutput(response) {
				if err != nil {
					yield(GenerateResponse{}, err)
					return
				}
				yield(GenerateResponse{
					NewResponses: []dto.Message{response},
					FullHistory:  totalHistory,
				}, nil)

				if response.Config.ExcludeFromHistory {
					continue
				}
				totalHistory = append(totalHistory, response)
			}
		}
	}
}

// generate generates a response from the GPT API. This is the internal function that is called by Generate.
func (g *Client) generate(messages []dto.Message) func(func(response dto.Message, err error) bool) {
	return func(yield func(response dto.Message, err error) bool) {
		body := dto.RequestDto{
			Messages: cleanMessages(messages),
			Tools:    g.generateFunctions(),
		}

		var gptRequest dto.ResponseDto
		requestClient := g.httpClient.R()

		if isOpenAIEndpoint(g.config.Endpoint) {
			requestClient = requestClient.SetHeader("Authorization", "Bearer "+g.config.ApiKey)
			body.Model = stringPtr(g.config.Model)
		} else {
			requestClient = requestClient.SetHeader("api-key", g.config.ApiKey)
		}
		response, err := requestClient.SetBody(body).SetResult(
			&gptRequest,
		).Post(g.config.Endpoint)

		if err != nil {
			logger.Error(err)
			yield(dto.Message{}, err)
			return
		}

		if !response.IsSuccess() {
			logger.Errorf("failed to generate response: %v", response)
			yield(dto.Message{}, fmt.Errorf("failed to generate response: %v", response))
			return
		}

		// use function if there is one
		message := gptRequest.Choices[0].Message
		newResponse := dto.Message{
			Role:      message.Role,
			Content:   message.Content,
			Usage:     gptRequest.Usage,
			ToolCalls: message.ToolCalls,
		}
		yield(newResponse, nil)
		messages = append(messages, newResponse)
		for newHistory, err := range g.useFunction(message, messages) {
			if err != nil {
				yield(dto.Message{}, err)
				return
			}

			yield(newHistory, nil)
		}
		if err != nil {
			logger.Error(err)
			yield(dto.Message{}, err)
			return
		}
	}
}

func (g *Client) generateFunctions() []dto.Tool {
	var returnedFunctions []dto.Tool
	for _, function := range *g.config.Functions {
		returnedFunctions = append(returnedFunctions, dto.Tool{
			Type: "function",
			Function: dto.ToolFunction{
				Name:        function.Name(),
				Description: function.Description(),
				Parameters:  function.Parameters(),
			},
		})
	}

	return returnedFunctions
}

// usePluginForInput uses the plugin for the input.
func (g *Client) usePluginForInput(input string) (*string, error) {
	output := &input
	var err error
	if g.config.Plugins == nil {
		return output, nil
	}

	for _, foundPlugin := range *g.config.Plugins {
		output, err = foundPlugin.ConvertInput(output)
		if err != nil {
			return nil, err
		}
	}
	return output, nil
}

// usePluginForOutput uses the plugin for the output.
func (g *Client) usePluginForOutput(response dto.Message) func(yield func(response dto.Message, err error) bool) {
	return func(yield func(response dto.Message, err error) bool) {
		if g.config.Plugins == nil {
			return
		}

		yield(response, nil)
		for _, foundPlugin := range *g.config.Plugins {
			convertedResponse, err := foundPlugin.ConvertOutput(response)
			if err != nil {
				yield(dto.Message{}, err)
				return
			}
			if convertedResponse != nil {
				yield(*convertedResponse.Message, nil)
			}
		}
	}

}

// useFunction uses the function if there is one in the response.
// It returns the new history and an error if there is one.
// If the function is configured to use GPT to interpret responses, it will call the GPT API again to interpret the responses.
func (g *Client) useFunction(result dto.MessageResponseDto, history []dto.Message) func(func(dto.Message, error) bool) {
	return func(yield func(dto.Message, error) bool) {
		newHistory := history
		if result.ToolCalls != nil && len(*result.ToolCalls) > 0 {
			for _, toolCall := range *result.ToolCalls {
				for _, function := range *g.config.Functions {
					if function.Name() == toolCall.Function.Name {
						var functionArguments map[string]interface{}
						err := json.Unmarshal([]byte(toolCall.Function.Arguments), &functionArguments)
						if err != nil {
							yield(dto.Message{}, err)
							return
						}
						result, err := function.OnMessage(functionArguments)
						if err != nil {
							yield(dto.Message{}, err)
							return
						}
						logger.Infof("Function %v executed with result %v", function.Name(), result)

						// Add history
						message := dto.Message{
							Role:       dto.RoleTool,
							Content:    convertFunctionContentToString(result.Content),
							ToolCallId: &toolCall.Id,
							Config:     result.Config,
						}
						if !result.Config.ExcludeFromHistory {
							newHistory = append(newHistory, message)
						}
						yield(message, nil)
						if function.Config().UseGptToInterpretResponses {
							for response, err := range g.generate(newHistory) {
								if err != nil {
									yield(dto.Message{}, err)
									return
								}

								yield(response, nil)
								newHistory = append(newHistory, response)
							}

							if err != nil {
								yield(dto.Message{}, err)
								return
							}
							return
						}
						for response, err := range function.OnAfterGptRespond {
							if err != nil {
								yield(dto.Message{}, err)
								return
							}

							resp := dto.Message{
								Role:    dto.RoleAssistant,
								Content: convertFunctionContentToString(response.Content),
								Config:  response.Config,
							}
							yield(resp, nil)
							if !response.Config.ExcludeFromHistory {
								newHistory = append(newHistory, resp)
							}
						}
						return
					}
				}
			}
		}
		return
	}
}

// createMessages creates a list of messages with history and prompt included.
func (g *Client) createMessages(prompt *string, history []dto.Message) (*dto.Message, []dto.Message) {
	var messages []dto.Message
	// only add system message if there is no history

	engine := g.config.Template
	renderedPrompt, err := engine.Render(g.config.Prompt)
	if err != nil {
		logger.Error(err)
		return nil, nil
	}
	messages = append(messages, dto.Message{
		Role:    dto.RoleSystem,
		Content: renderedPrompt,
	})

	for _, message := range history {
		message.Usage = nil
		messages = append(messages, message)
	}

	var promptMessage dto.Message
	if prompt != nil {
		promptMessage = dto.Message{
			Role:    dto.RoleUser,
			Content: *prompt,
		}
		messages = append(messages, promptMessage)
	}
	return &promptMessage, messages
}

package gpt

import (
	"github.com/go-resty/resty/v2"
	"github.com/google/logger"
	"github.com/jarcoal/httpmock"
	template "github.com/meta-metopia/go-packages/pkg/ai"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"testing"
)

type GptTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	client *resty.Client
}

func (suite *GptTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.client = resty.New()
	httpmock.ActivateNonDefault(suite.client.GetClient())
}

func (suite *GptTestSuite) TearDownTest() {
	httpmock.DeactivateAndReset()
	suite.ctrl.Finish()
}

func (suite *GptTestSuite) TestGptWithoutFunctionCall() {
	logger.Init("TestLogger", true, false, io.Discard)
	engine := template.NewMockEngine(suite.ctrl)
	engine.EXPECT().Render(gomock.Any()).Return("Mock Data", nil).AnyTimes()

	body := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Mock Data",
				},
			},
		},
	}
	url := "http://localhost:8080"
	responder, err := httpmock.NewJsonResponder(http.StatusOK, body)
	if err != nil {
		suite.T().Fatal(err)
	}
	httpmock.RegisterResponder("POST", url, responder)

	aiFunctions := make([]functions.FunctionInterface, 0)
	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		Config{
			Endpoint:  url,
			ApiKey:    "123",
			Functions: &aiFunctions,
			Template:  engine,
			Store:     functionStore,
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	newResponses, err := client.Generate(&prompt, []dto.Message{})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(newResponses.NewResponses), 1)
	assert.Equal(suite.T(), newResponses.NewResponses[0].Role, dto.RoleAssistant)

	assert.Equal(suite.T(), 2, len(newResponses.FullHistory))

	assert.Equal(suite.T(), newResponses.FullHistory[0].Role, dto.RoleUser)
	assert.Equal(suite.T(), newResponses.FullHistory[1].Role, dto.RoleAssistant)

	newResponses, err = client.Generate(&prompt, newResponses.FullHistory)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(newResponses.NewResponses), 1)
	assert.Equal(suite.T(), 4, len(newResponses.FullHistory))

	assert.Equal(suite.T(), newResponses.FullHistory[0].Role, dto.RoleUser)
	assert.Equal(suite.T(), newResponses.FullHistory[1].Role, dto.RoleAssistant)
	assert.Equal(suite.T(), newResponses.FullHistory[2].Role, dto.RoleUser)
	assert.Equal(suite.T(), newResponses.FullHistory[3].Role, dto.RoleAssistant)
}

func (suite *GptTestSuite) TestGptWithoutFunctionCallIterator() {
	logger.Init("TestLogger", true, false, io.Discard)
	engine := template.NewMockEngine(suite.ctrl)
	engine.EXPECT().Render(gomock.Any()).Return("Mock Data", nil).AnyTimes()

	body := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Mock Data",
				},
			},
		},
	}
	url := "http://localhost:8080"
	responder, err := httpmock.NewJsonResponder(http.StatusOK, body)
	if err != nil {
		suite.T().Fatal(err)
	}
	httpmock.RegisterResponder("POST", url, responder)

	aiFunctions := make([]functions.FunctionInterface, 0)
	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		Config{
			Endpoint:  url,
			ApiKey:    "123",
			Functions: &aiFunctions,
			Template:  engine,
			Store:     functionStore,
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"

	var finalResponse GenerateResponse
	for response, err := range client.GenerateIterator(&prompt, []dto.Message{}) {
		assert.Nil(suite.T(), err)
		finalResponse = response
	}

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(finalResponse.NewResponses), 1)
	assert.Equal(suite.T(), finalResponse.NewResponses[0].Role, dto.RoleAssistant)

	assert.Equal(suite.T(), 2, len(finalResponse.FullHistory))

	assert.Equal(suite.T(), finalResponse.FullHistory[0].Role, dto.RoleUser)
	assert.Equal(suite.T(), finalResponse.FullHistory[1].Role, dto.RoleAssistant)

	//finalResponse, err = client.Generate(&prompt, finalResponse.FullHistory)
	for response, err := range client.GenerateIterator(&prompt, finalResponse.FullHistory) {
		assert.Nil(suite.T(), err)
		finalResponse = response
	}

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(finalResponse.NewResponses), 1)
	assert.Equal(suite.T(), 4, len(finalResponse.FullHistory))

	assert.Equal(suite.T(), finalResponse.FullHistory[0].Role, dto.RoleUser)
	assert.Equal(suite.T(), finalResponse.FullHistory[1].Role, dto.RoleAssistant)
	assert.Equal(suite.T(), finalResponse.FullHistory[2].Role, dto.RoleUser)
	assert.Equal(suite.T(), finalResponse.FullHistory[3].Role, dto.RoleAssistant)
}

func (suite *GptTestSuite) TestGptWithError() {
	logger.Init("TestLogger", true, false, io.Discard)
	engine := template.NewMockEngine(suite.ctrl)
	engine.EXPECT().Render(gomock.Any()).Return("Mock Data", nil).AnyTimes()

	url := "http://localhost:8080"
	responder, err := httpmock.NewJsonResponder(http.StatusOK, "invalid json")
	if err != nil {
		suite.T().Fatal(err)
	}
	httpmock.RegisterResponder("POST", url, responder)

	aiFunctions := make([]functions.FunctionInterface, 0)
	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		Config{
			Endpoint:  url,
			ApiKey:    "123",
			Functions: &aiFunctions,
			Template:  engine,
			Store:     functionStore,
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	response, err := client.Generate(&prompt, []dto.Message{})

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), response.NewResponses)
	assert.Nil(suite.T(), response.FullHistory)
}

func (suite *GptTestSuite) TestGptWithErrorIterator() {
	logger.Init("TestLogger", true, false, io.Discard)
	engine := template.NewMockEngine(suite.ctrl)
	engine.EXPECT().Render(gomock.Any()).Return("Mock Data", nil).AnyTimes()

	url := "http://localhost:8080"
	responder, err := httpmock.NewJsonResponder(http.StatusOK, "invalid json")
	if err != nil {
		suite.T().Fatal(err)
	}
	httpmock.RegisterResponder("POST", url, responder)

	aiFunctions := make([]functions.FunctionInterface, 0)
	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		Config{
			Endpoint:  url,
			ApiKey:    "123",
			Functions: &aiFunctions,
			Template:  engine,
			Store:     functionStore,
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	for response, err := range client.GenerateIterator(&prompt, []dto.Message{}) {
		assert.NotNil(suite.T(), err)
		assert.Nil(suite.T(), response.NewResponses)
		assert.Nil(suite.T(), response.FullHistory)

	}
}

func (suite *GptTestSuite) TestGptWithErrorFromServer() {
	logger.Init("TestLogger", true, false, io.Discard)
	engine := template.NewMockEngine(suite.ctrl)
	engine.EXPECT().Render(gomock.Any()).Return("Mock Data", nil).AnyTimes()

	url := "http://localhost:8080"

	responder, err := httpmock.NewJsonResponder(http.StatusInternalServerError, "Internal Server Error")
	if err != nil {
		suite.T().Fatal(err)
	}
	httpmock.RegisterResponder("POST", url, responder)

	aiFunctions := make([]functions.FunctionInterface, 0)
	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		Config{
			Endpoint:  url,
			ApiKey:    "123",
			Functions: &aiFunctions,
			Template:  engine,
			Store:     functionStore,
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	response, err := client.Generate(&prompt, []dto.Message{})

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), response.NewResponses)
	assert.Nil(suite.T(), response.FullHistory)
}

func (suite *GptTestSuite) TestGptWithFunctionCall() {
	logger.Init("TestLogger", true, false, io.Discard)
	engine := template.NewMockEngine(suite.ctrl)
	engine.EXPECT().Render(gomock.Any()).Return("Mock Data", nil).AnyTimes()

	body := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role": "assistant",
					"tool_calls": []map[string]interface{}{
						{
							"id":   "1",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "Mock Function",
								"arguments": `{"prompt":"Prompt"}`,
							},
						},
					},
				},
			},
		},
	}
	url := "http://localhost:8080"
	responder, err := httpmock.NewJsonResponder(http.StatusOK, body)
	if err != nil {
		suite.T().Fatal(err)
	}
	httpmock.RegisterResponder("POST", url, responder)

	var aiFunctions = make([]functions.FunctionInterface, 1)
	function := functions.NewMockFunctionInterface(suite.ctrl)
	aiFunctions[0] = function
	function.EXPECT().OnMessage(gomock.Any()).Return(&functions.FunctionGptResponse{Content: "Mock Function Response"}, nil).Times(1)
	function.EXPECT().Name().Return("Mock Function").AnyTimes()
	function.EXPECT().Description().Return("Mock Function Description").Times(1)
	function.EXPECT().Parameters().Return(map[string]interface{}{}).Times(1)
	function.EXPECT().SetStore(gomock.Any()).Times(1)
	function.EXPECT().Config().Return(functions.FunctionConfig{UseGptToInterpretResponses: false}).Times(1)
	function.EXPECT().OnInit().Times(1)
	function.EXPECT().OnAfterGptRespond(gomock.Any()).AnyTimes()

	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		Config{
			Endpoint:  url,
			ApiKey:    "123",
			Functions: &aiFunctions,
			Template:  engine,
			Store:     functionStore,
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	response, err := client.Generate(&prompt, []dto.Message{})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(response.NewResponses), 2)
	assert.Equal(suite.T(), 3, len(response.FullHistory))

	assert.Equal(suite.T(), response.FullHistory[0].Role, dto.RoleUser)
	assert.Equal(suite.T(), response.FullHistory[1].Role, dto.RoleAssistant)
	assert.Equal(suite.T(), response.FullHistory[2].Role, dto.RoleTool)
}

// When Multiple function exists, every tool response should follow by a function call
func (suite *GptTestSuite) TestGptWithMultipleFunctionCalls() {
	logger.Init("TestLogger", true, false, io.Discard)
	engine := template.NewMockEngine(suite.ctrl)
	engine.EXPECT().Render(gomock.Any()).Return("Mock Data", nil).AnyTimes()

	body := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role": "assistant",
					"tool_calls": []map[string]interface{}{
						{
							"id":   "1",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "Mock Function",
								"arguments": `{"prompt":"Prompt"}`,
							},
						},
					},
				},
			},
		},
	}

	body2 := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Mock Data",
				},
			},
		},
	}

	url := "http://localhost:8080"
	toolResponse, err := httpmock.NewJsonResponse(http.StatusOK, body)

	if err != nil {
		suite.T().Fatal(err)
	}
	assistantResponse, err := httpmock.NewJsonResponse(http.StatusOK, body2)

	responder := httpmock.ResponderFromMultipleResponses([]*http.Response{
		toolResponse,
		toolResponse,
		assistantResponse,
	})
	httpmock.RegisterResponder("POST", url, responder)

	var aiFunctions = make([]functions.FunctionInterface, 1)
	function := functions.NewMockFunctionInterface(suite.ctrl)
	aiFunctions[0] = function
	function.EXPECT().OnMessage(gomock.Any()).Return(&functions.FunctionGptResponse{Content: "Mock Function Response"}, nil).AnyTimes()
	function.EXPECT().Name().Return("Mock Function").AnyTimes()
	function.EXPECT().Description().Return("Mock Function Description").AnyTimes()
	function.EXPECT().Parameters().Return(map[string]interface{}{}).AnyTimes()
	function.EXPECT().SetStore(gomock.Any()).Times(1)
	function.EXPECT().Config().Return(functions.FunctionConfig{UseGptToInterpretResponses: true}).AnyTimes()
	function.EXPECT().OnInit().Times(1)
	function.EXPECT().OnAfterGptRespond(gomock.Any()).AnyTimes()

	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		Config{
			Endpoint:  url,
			ApiKey:    "123",
			Functions: &aiFunctions,
			Template:  engine,
			Store:     functionStore,
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	response, err := client.Generate(&prompt, []dto.Message{})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(response.NewResponses), 5)
	assert.Equal(suite.T(), 6, len(response.FullHistory))

	assert.Equal(suite.T(), response.FullHistory[0].Role, dto.RoleUser)
	assert.Equal(suite.T(), response.FullHistory[1].Role, dto.RoleAssistant)
	assert.Equal(suite.T(), response.FullHistory[2].Role, dto.RoleTool)
	assert.Equal(suite.T(), response.FullHistory[3].Role, dto.RoleAssistant)
	assert.Equal(suite.T(), response.FullHistory[4].Role, dto.RoleTool)
	assert.Equal(suite.T(), response.FullHistory[5].Role, dto.RoleAssistant)
}

func (suite *GptTestSuite) TestGptWithFunctionCallAndUseGptToInterpret() {
	logger.Init("TestLogger", true, false, io.Discard)
	engine := template.NewMockEngine(suite.ctrl)
	engine.EXPECT().Render(gomock.Any()).Return("Mock Data", nil).AnyTimes()

	body, _ := httpmock.NewJsonResponse(http.StatusOK, map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role": "assistant",
					"tool_calls": []map[string]interface{}{
						{
							"id":   "1",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "Mock Function",
								"arguments": `{"prompt":"Prompt"}`,
							},
						},
					},
				},
			},
		},
	})

	body2, _ := httpmock.NewJsonResponse(http.StatusOK, map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Mock Data",
				},
			},
		},
	})
	url := "http://localhost:8080"
	responder := httpmock.ResponderFromMultipleResponses([]*http.Response{body, body2})
	httpmock.RegisterResponder("POST", url, responder)

	var aiFunctions = make([]functions.FunctionInterface, 1)
	function := functions.NewMockFunctionInterface(suite.ctrl)
	aiFunctions[0] = function
	function.EXPECT().OnMessage(gomock.Any()).Return(&functions.FunctionGptResponse{Content: "Mock Function Response"}, nil).Times(1)
	function.EXPECT().Name().Return("Mock Function").AnyTimes()
	function.EXPECT().Description().Return("Mock Function Description").AnyTimes()
	function.EXPECT().Parameters().Return(map[string]interface{}{}).AnyTimes()
	function.EXPECT().SetStore(gomock.Any()).Times(1)
	function.EXPECT().Config().Return(functions.FunctionConfig{UseGptToInterpretResponses: true}).Times(1)
	function.EXPECT().OnInit().Times(1)
	function.EXPECT().OnAfterGptRespond(gomock.Any()).AnyTimes()

	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		Config{
			Endpoint:  url,
			ApiKey:    "123",
			Functions: &aiFunctions,
			Template:  engine,
			Store:     functionStore,
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	response, err := client.Generate(&prompt, []dto.Message{})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(response.NewResponses), 3)
	assert.Equal(suite.T(), 4, len(response.FullHistory))
	assert.Equal(suite.T(), response.FullHistory[0].Role, dto.RoleUser)
	assert.Equal(suite.T(), response.FullHistory[1].Role, dto.RoleAssistant)
	assert.Equal(suite.T(), response.FullHistory[2].Role, dto.RoleTool)
	assert.Equal(suite.T(), response.FullHistory[3].Role, dto.RoleAssistant)
}

func (suite *GptTestSuite) TestShouldIncludeInHistoryTrue() {
	logger.Init("TestLogger", true, false, io.Discard)
	engine := template.NewMockEngine(suite.ctrl)
	engine.EXPECT().Render(gomock.Any()).Return("Mock Data", nil).AnyTimes()

	body, _ := httpmock.NewJsonResponse(http.StatusOK, map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role": "assistant",
					"tool_calls": []map[string]interface{}{
						{
							"id":   "1",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "Mock Function",
								"arguments": `{"prompt":"Prompt"}`,
							},
						},
					},
				},
			},
		},
	})

	url := "http://localhost:8080"
	responder := httpmock.ResponderFromResponse(body)
	httpmock.RegisterResponder("POST", url, responder)

	var aiFunctions = make([]functions.FunctionInterface, 1)
	function := functions.NewMockFunctionInterface(suite.ctrl)
	aiFunctions[0] = function
	function.EXPECT().OnMessage(gomock.Any()).Return(&functions.FunctionGptResponse{Content: "Mock Function Response", Config: functions.FunctionGptResponseConfig{ExcludeFromHistory: false}}, nil).AnyTimes()
	function.EXPECT().Name().Return("Mock Function").AnyTimes()
	function.EXPECT().Description().Return("Mock Function Description").AnyTimes()
	function.EXPECT().Parameters().Return(map[string]interface{}{}).AnyTimes()
	function.EXPECT().SetStore(gomock.Any()).Times(1)
	function.EXPECT().Config().Return(functions.FunctionConfig{UseGptToInterpretResponses: false}).AnyTimes()
	function.EXPECT().OnInit().Times(1)
	function.EXPECT().OnAfterGptRespond(gomock.Any()).AnyTimes()

	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		Config{
			Endpoint:  url,
			ApiKey:    "123",
			Functions: &aiFunctions,
			Template:  engine,
			Store:     functionStore,
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	// Repeat with ShouldIncludeInHistory set to false
	response, err := client.Generate(&prompt, []dto.Message{})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 2, len(response.NewResponses))
	assert.Equal(suite.T(), 3, len(response.FullHistory))
	assert.Equal(suite.T(), response.FullHistory[0].Role, dto.RoleUser)
	assert.Equal(suite.T(), response.FullHistory[1].Role, dto.RoleAssistant)
	assert.Equal(suite.T(), response.FullHistory[2].Role, dto.RoleTool)
}

func (suite *GptTestSuite) TestShouldIncludeInHistoryFalse() {
	logger.Init("TestLogger", true, false, io.Discard)
	engine := template.NewMockEngine(suite.ctrl)
	engine.EXPECT().Render(gomock.Any()).Return("Mock Data", nil).AnyTimes()

	body, _ := httpmock.NewJsonResponse(http.StatusOK, map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role": "assistant",
					"tool_calls": []map[string]interface{}{
						{
							"id":   "1",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "Mock Function",
								"arguments": `{"prompt":"Prompt"}`,
							},
						},
					},
				},
			},
		},
	})

	url := "http://localhost:8080"
	responder := httpmock.ResponderFromResponse(body)
	httpmock.RegisterResponder("POST", url, responder)

	var aiFunctions = make([]functions.FunctionInterface, 1)
	function := functions.NewMockFunctionInterface(suite.ctrl)
	aiFunctions[0] = function
	function.EXPECT().OnMessage(gomock.Any()).Return(&functions.FunctionGptResponse{Content: "Mock Function Response", Config: functions.FunctionGptResponseConfig{ExcludeFromHistory: true}}, nil).AnyTimes()
	function.EXPECT().Name().Return("Mock Function").AnyTimes()
	function.EXPECT().Description().Return("Mock Function Description").AnyTimes()
	function.EXPECT().Parameters().Return(map[string]interface{}{}).AnyTimes()
	function.EXPECT().SetStore(gomock.Any()).Times(1)
	function.EXPECT().Config().Return(functions.FunctionConfig{UseGptToInterpretResponses: false}).AnyTimes()
	function.EXPECT().OnInit().Times(1)
	function.EXPECT().OnAfterGptRespond(gomock.Any()).AnyTimes()

	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		Config{
			Endpoint:  url,
			ApiKey:    "123",
			Functions: &aiFunctions,
			Template:  engine,
			Store:     functionStore,
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	response, err := client.Generate(&prompt, []dto.Message{})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, len(response.NewResponses))
	assert.Equal(suite.T(), 2, len(response.FullHistory))
	assert.Equal(suite.T(), response.FullHistory[0].Role, dto.RoleUser)
	assert.Equal(suite.T(), response.FullHistory[1].Role, dto.RoleAssistant)
}

func (suite *GptTestSuite) TestSetPlugin() {
	client := &Client{}
	client.SetPlugins(&[]plugin.Interface{plugin.NewStandardOutputPlugin()})
	assert.Equal(suite.T(), len(*client.config.Plugins), 1)
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(GptTestSuite))
}

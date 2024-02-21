package gpt

import (
	"github.com/go-resty/resty/v2"
	"github.com/google/logger"
	"github.com/jarcoal/httpmock"
	template "github.com/meta-metopia/go-packages/pkg/ai"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
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
		&aiFunctions,
		engine,
		functionStore,
		Config{
			Endpoint: url,
			ApiKey:   "123",
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	newResponses, history, err := client.Generate(&prompt, []dto.MessageResponseDto{})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(newResponses), 1)
	assert.Equal(suite.T(), 3, len(history))

	newResponses, history, err = client.Generate(&prompt, history)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(newResponses), 1)
	assert.Equal(suite.T(), 5, len(history))
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
		&aiFunctions,
		engine,
		functionStore,
		Config{
			Endpoint: url,
			ApiKey:   "123",
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	content, history, err := client.Generate(&prompt, []dto.MessageResponseDto{})

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), content)
	assert.Equal(suite.T(), 0, len(history))
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
		&aiFunctions,
		engine,
		functionStore,
		Config{
			Endpoint: url,
			ApiKey:   "123",
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	content, history, err := client.Generate(&prompt, []dto.MessageResponseDto{})

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), content)
	assert.Equal(suite.T(), 0, len(history))
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
					"function_call": map[string]interface{}{
						"name":      "Mock Function",
						"arguments": `{"prompt":"Prompt"}`,
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

	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		&aiFunctions,
		engine,
		functionStore,
		Config{
			Endpoint: url,
			ApiKey:   "123",
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	newResponses, history, err := client.Generate(&prompt, []dto.MessageResponseDto{})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(newResponses), 2)
	assert.Equal(suite.T(), 4, len(history))
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
					"function_call": map[string]interface{}{
						"name":      "Mock Function",
						"arguments": `{"prompt":"Prompt"}`,
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

	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		&aiFunctions,
		engine,
		functionStore,
		Config{
			Endpoint: url,
			ApiKey:   "123",
		},
	)
	client.SetClient(suite.client)

	prompt := "Prompt"
	newResponses, history, err := client.Generate(&prompt, []dto.MessageResponseDto{})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(newResponses), 3)
	assert.Equal(suite.T(), 5, len(history))
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(GptTestSuite))
}
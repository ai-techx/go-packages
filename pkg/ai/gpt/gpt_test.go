package gpt

import (
	"encoding/json"
	"github.com/jarcoal/httpmock"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
	"github.com/meta-metopia/go-packages/pkg/ai/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"net/http"
	"testing"
)

type GptTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
}

func (suite *GptTestSuite) SetupTest() {
	httpmock.Activate()
	suite.ctrl = gomock.NewController(suite.T())

}

func (suite *GptTestSuite) TearDownTest() {
	httpmock.DeactivateAndReset()
}

func (suite *GptTestSuite) TestGptWithoutFunctionCall() {
	engine := mocks.NewMockTemplateEngine(suite.ctrl)
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
	json, _ := json.Marshal(body)
	jsonString := string(json)
	httpmock.RegisterResponder("POST", "http://localhost:8080", httpmock.NewStringResponder(http.StatusOK, jsonString))

	aiFunctions := make([]functions.IFunction, 0)
	functionStore := make(functions.FunctionStore)
	client := NewGptClient(
		&aiFunctions,
		engine,
		functionStore,
		Config{
			Endpoint: "http://localhost:8080",
			ApiKey:   "123",
		},
	)

	prompt := "Prompt"
	content, history, err := client.Generate(&prompt, []dto.MessageResponseDto{})

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "Mock Data", *content)
	assert.Equal(suite.T(), 3, len(history))

	content, history, err = client.Generate(&prompt, history)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "Mock Data", *content)
	assert.Equal(suite.T(), 5, len(history))
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(GptTestSuite))
}

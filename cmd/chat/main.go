package main

import (
	"fmt"
	"github.com/google/logger"
	functions2 "github.com/meta-metopia/go-packages/cmd/chat/functions"
	"github.com/meta-metopia/go-packages/cmd/chat/input"
	"github.com/meta-metopia/go-packages/cmd/chat/output"
	"github.com/meta-metopia/go-packages/cmd/chat/template"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
	"io"
	"os"
)

type Model struct {
	Name              string  `json:"name"`
	PromptPricing     float64 `json:"prompt_pricing"`
	CompletionPricing float64 `json:"completion_pricing"`
}

var AvailableModels = []Model{
	{
		Name:              "gpt-3.5-turbo",
		PromptPricing:     0.0005,
		CompletionPricing: 0.0015,
	},
	{
		Name:              "gpt-4-turbo-preview",
		PromptPricing:     0.01,
		CompletionPricing: 0.03,
	},
}

func deleteLastLine() {
	fmt.Printf("\033[1A\033[K")
}

func stringPtr(s string) *string {
	return &s
}

func calculatePricing(model Model, history []dto.MessageResponseDto) (total float64, completionToken int, promptToken int) {
	for _, message := range history {
		if message.Usage == nil {
			continue
		}
		promptPrice := model.PromptPricing / 1000 * float64(message.Usage.PromptToken)
		completionPrice := model.CompletionPricing / 1000 * float64(message.Usage.CompletionToken)
		completionToken += message.Usage.CompletionToken
		promptToken += message.Usage.PromptToken

		total += promptPrice + completionPrice
	}
	return total, completionToken, promptToken
}

func main() {
	logger.Init("Chatbot", true, false, io.Discard)
	inputClient := input.NewPromptInput()
	outputClient := output.NewAzureSpeechOutput("zh-CN-XiaoyouNeural")
	gptFunctions := []functions.FunctionInterface{
		functions2.NewGetAllMenuFunction(),
		functions2.NewCompleteOrderFunction(),
		functions2.NewAddDishFunction(),
	}

	functionStore := functions.FunctionStore{}
	model := AvailableModels[0]
	config := gpt.Config{
		Endpoint: "https://api.openai.com/v1/chat/completions",
		Model:    model.Name,
		ApiKey:   os.Getenv("OPENAPI_KEY"),
		Prompt:   `你是一個點餐機器人，你可以幫用戶去點餐。`,
	}
	templateEngine := template.NewEngine()

	gptClient := gpt.NewGptClient(&gptFunctions, templateEngine, functionStore, config)
	history := make([]dto.MessageResponseDto, 0)

	for prompt, err := range inputClient.Run {
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Generating response...")
		generate, newHistory, err := gptClient.Generate(&prompt, history)
		if err != nil {
			return
		}
		history = newHistory
		deleteLastLine()
		totalPricing, completionToken, promptToken := calculatePricing(model, history)
		fmt.Printf("(Total pricing: $%.5f, Prompt Token: %d, Completion Token: %d)\n", totalPricing, promptToken, completionToken)

		for _, response := range generate {
			if len(response.Content) > 0 {
				fmt.Println("Generating output...")
				err := outputClient.Run(response.Content)
				if err != nil {
					logger.Fatal(err)
					return
				}
				deleteLastLine()
			}
		}
	}

}

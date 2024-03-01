package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/google/logger"
	functions2 "github.com/meta-metopia/go-packages/cmd/chat/functions"
	"github.com/meta-metopia/go-packages/cmd/chat/input"
	plugins2 "github.com/meta-metopia/go-packages/cmd/chat/plugins"
	"github.com/meta-metopia/go-packages/cmd/chat/template"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/plugin"
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

func saveHistoryToFile(history []dto.Message, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		logger.Fatal(err)
	}
	defer file.Close()

	indent, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return
	}

	_, err = file.Write(indent)
}

func calculatePricing(model Model, history []dto.Message) (total float64, completionToken int, promptToken int) {
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
	//outputClient := output.NewAzureSpeechOutput(os.Getenv("VOICE_NAME"))
	gptFunctions := []functions.FunctionInterface{
		functions2.NewAddDishFunction(),
		functions2.NewCompleteOrderFunction(),
		functions2.NewGetAllMenuFunction(),
	}

	plugins := []plugin.Interface{
		plugins2.NewAzurePlugin(os.Getenv("VOICE_NAME")),
	}

	functionStore := functions.FunctionStore{}
	model := AvailableModels[0]
	templateEngine := template.NewEngine()
	config := gpt.Config{
		Endpoint:  "https://api.openai.com/v1/chat/completions",
		Model:     model.Name,
		ApiKey:    os.Getenv("OPENAPI_KEY"),
		Prompt:    `你是一個問答機器人`,
		Functions: &gptFunctions,
		Store:     functionStore,
		Template:  templateEngine,
		Plugins:   &plugins,
	}

	gptClient := gpt.NewGptClient(config)
	history := make([]dto.Message, 0)

	for prompt, err := range inputClient.Run {
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Generating response...")

		for response, err := range gptClient.GenerateIterator(&prompt, history) {
			saveHistoryToFile(history, "history.json")
			if err != nil {
				fmt.Println(err)
				return
			}

			history = response.FullHistory
			totalPricing, completionToken, promptToken := calculatePricing(model, history)
			fmt.Printf(color.RedString("Usage: ")+"Total pricing: $%.5f, Prompt Token: %d, Completion Token: %d\n", totalPricing, promptToken, completionToken)
			if len(response.NewResponses) == 0 {
				continue
			}

			for _, response := range response.NewResponses {
				if len(response.Content) > 0 {
					var content string
					content = response.Content
					if response.Name != nil {
						content = "調用函數" + *response.Name + ": " + content
					}
					//err := outputClient.Run(content)
					fmt.Println(content)
					//if err != nil {
					//	logger.Fatal(err)
					//	return
					//}
				}
			}
		}
	}
}

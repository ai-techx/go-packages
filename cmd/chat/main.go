package main

import (
	"fmt"
	"github.com/google/logger"
	"github.com/meta-metopia/go-packages/cmd/chat/input"
	"github.com/meta-metopia/go-packages/cmd/chat/output"
	"github.com/meta-metopia/go-packages/cmd/chat/template"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/functions"
	"os"
)

func deleteLastLine() {
	fmt.Printf("\033[1A\033[K")
}

func stringPtr(s string) *string {
	return &s
}

func main() {
	inputClient := input.NewPromptInput()
	outputClient := output.NewAzureSpeechOutput("zh-CN-XiaoyouNeural")
	gptFunctions := make([]functions.FunctionInterface, 0)
	functionStore := functions.FunctionStore{}
	config := gpt.Config{
		Endpoint: "https://api.openai.com/v1/chat/completions",
		Model:    "gpt-3.5-turbo",
		ApiKey:   os.Getenv("OPENAPI_KEY"),
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

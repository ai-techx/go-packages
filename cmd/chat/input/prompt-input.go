package input

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

type PromptInput struct {
}

func NewPromptInput() Input {
	return &PromptInput{}
}

func (p *PromptInput) Run(yield func(input string, err error) bool) {
	// infinite loop
	for {
		prompt := promptui.Prompt{
			Label: "Input",
		}

		result, err := prompt.Run()
		if err != nil {
			if !yield("", err) {
				break
			}
		} else {
			if result == "exit" {
				break
			}
			if !yield(result, nil) {
				break
			}
		}
	}
	fmt.Println(color.RedString("Goodbye!"))
	return
}

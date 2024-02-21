package main

type Model struct {
	Name string `json:"name"`
}

var AvailableModels = []Model{
	{
		Name: "gpt-3.5-turbo",
	},
	{
		Name: "gpt-4-turbo-preview",
	},
}

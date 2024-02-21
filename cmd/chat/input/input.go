package input

type Input interface {
	//Run will convert user input to string and yield it to the chat
	Run(yield func(input string, err error) bool)
}

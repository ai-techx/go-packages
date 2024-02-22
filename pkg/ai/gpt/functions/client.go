package functions

type FunctionClient struct {
}

func (m FunctionClient) OnInit() error {
	return nil
}

func (m FunctionClient) OnClose() error {
	return nil
}

func (m FunctionClient) OnAfterGptRespond(yield func(FunctionGptResponse, error) bool) {

}

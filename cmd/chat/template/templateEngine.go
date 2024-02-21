package template

import template "github.com/meta-metopia/go-packages/pkg/ai"

type Engine struct {
}

// NewEngine returns a new instance of Engine.
func NewEngine() template.Engine {
	return &Engine{}
}

// Render renders a string using the template engine.
func (e *Engine) Render(arg0 string) (string, error) {
	return arg0, nil
}

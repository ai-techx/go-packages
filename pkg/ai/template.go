package template

//go:generate mockgen -destination=mocks/template.go -package=mocks . TemplateEngine
type Engine interface {
	Render(template string) (string, error)
}

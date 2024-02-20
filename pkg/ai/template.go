package template

//go:generate mockgen -destination=mock_template.go -package=template . Engine
type Engine interface {
	Render(template string) (string, error)
}

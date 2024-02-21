package output

type Output interface {
	Run(input string) error
}

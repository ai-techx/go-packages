package errors

type ErrorInterface interface {
	Error() string
	Code() ErrorCode
}

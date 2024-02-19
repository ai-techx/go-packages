package errors

type DocumentNotFound struct {
	code ErrorCode
}

// NewDocumentNotFound creates a new DocumentNotFound
func NewDocumentNotFound() *DocumentNotFound {
	return &DocumentNotFound{
		code: ErrorDocumentNotFound,
	}
}

func (e *DocumentNotFound) Error() string {
	return "The given resource was not found"
}

func (e *DocumentNotFound) Code() ErrorCode {
	return e.code
}

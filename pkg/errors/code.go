package errors

// ErrorCode 0-999: General errors,
// 1000-1999 400 Bad Request,
// 2000-2999 401 Unauthorized,
// 3000-3999 403 Forbidden,
// 4000-4999 404 Not Found,
// 5000-5999 500 Internal Server Error
type ErrorCode int

const (
	ErrorMissingAPIKey    ErrorCode = 2000
	ErrorInvalidAPIKey    ErrorCode = 2001
	ErrorDocumentNotFound ErrorCode = 4000
)

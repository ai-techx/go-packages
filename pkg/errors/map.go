package errors

import "net/http"

// MapErrorCodeToHTTPStatus maps an error code to an HTTP status code.
func MapErrorCodeToHTTPStatus(code ErrorCode) int {
	if code >= 1000 && code < 2000 {
		return http.StatusBadRequest
	}
	if code >= 2000 && code < 3000 {
		return http.StatusUnauthorized
	}
	if code >= 3000 && code < 4000 {
		return http.StatusForbidden
	}
	if code >= 4000 && code < 5000 {
		return http.StatusNotFound
	}
	if code >= 5000 && code < 6000 {
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}

func MapErrorToHTTPStatus(err error) int {
	if e, ok := err.(ErrorInterface); ok {
		return MapErrorCodeToHTTPStatus(e.Code())
	}
	return http.StatusInternalServerError
}

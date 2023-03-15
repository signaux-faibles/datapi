// Package utils contient le code technique commun
package utils

// Jerror interface for JSON errors
type Jerror interface {
	Error() string
	Code() int
}

// JSONerror enables returning enriched errors with JSON status
type JSONerror struct {
	error string
	code  int
}

// Error() provide the classical error status
func (j JSONerror) Error() string {
	return j.error
}

// Code provides JSON error code to be returned
func (j JSONerror) Code() int {
	return j.code
}

// NewJSONerror return JSON error from string
func NewJSONerror(code int, e string) JSONerror {
	return JSONerror{error: e, code: code}
}

// ErrorToJSON retourne une JSONerror avec un code http
func ErrorToJSON(code int, e error) JSONerror {
	return JSONerror{error: e.Error(), code: code}
}

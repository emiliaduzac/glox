package errors

import (
	"fmt"
)

type LoxError struct {
	Line    int
	Message string
}

func NewLoxError(line int, message string) *LoxError {
	return &LoxError{
		Line:    line,
		Message: message,
	}
}

// Error implementa la interfaz error para LoxError, devolviendo un mensaje de error formateado.
func (e *LoxError) Error() string {
	return fmt.Sprintf("[line %d] Error: %s", e.Line, e.Message)
}

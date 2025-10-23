package types

import (
	"net/http"
)

type ErrorResponse struct {
	ErrorInfo ErrorContent `json:"error"`
}

type ErrorContent struct {
	Message        string `json:"message"`
	Code           string `json:"code"`
	HTTPCode       int    `json:"-"`
	DisplayMessage string `json:"displayMessage"`
}

func (e *ErrorResponse) Error() string {
	return e.ErrorInfo.Message
}

func CustomError(message, code string, httpCode int, displayMessage string) *ErrorResponse {
	return &ErrorResponse{
		ErrorInfo: ErrorContent{
			Message:        message,
			Code:           code,
			HTTPCode:       httpCode,
			DisplayMessage: displayMessage,
		},
	}
}

func NewNotFoundError(message, code, displayMessage string) *ErrorResponse {
	return &ErrorResponse{
		ErrorInfo: ErrorContent{
			Message:        message,
			Code:           code,
			HTTPCode:       http.StatusNotFound,
			DisplayMessage: displayMessage,
		},
	}
}

func InternalServerError(message string) *ErrorResponse {
	return &ErrorResponse{
		ErrorInfo: ErrorContent{
			Message:        message,
			Code:           "INTERNAL_ERROR",
			HTTPCode:       http.StatusInternalServerError,
			DisplayMessage: "An internal error occurred",
		},
	}
}

func BadRequestError(message, code, displayMessage string) *ErrorResponse {
	return &ErrorResponse{
		ErrorInfo: ErrorContent{
			Message:        message,
			Code:           code,
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: displayMessage,
		},
	}
}

func UnauthorizedError(message, code, displayMessage string) *ErrorResponse {
	return &ErrorResponse{
		ErrorInfo: ErrorContent{
			Message:        message,
			Code:           code,
			HTTPCode:       http.StatusForbidden,
			DisplayMessage: displayMessage,
		},
	}
}

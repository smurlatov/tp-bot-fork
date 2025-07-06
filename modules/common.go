package modules

type APIError interface {
	GetCode() string
	GetMessage() string
	Error() string
}

type BaseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *BaseError) GetCode() string {
	return e.Code
}

func (e *BaseError) GetMessage() string {
	return e.Message
}

func (e *BaseError) Error() string {
	return e.Code + ": " + e.Message
}

func NewError(code, message string) APIError {
	return &BaseError{
		Code:    code,
		Message: message,
	}
}

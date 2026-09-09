package httperror

import "judge-backend/api"

var (
	InternalServerError = New(500, "internal server error")
)

type Error struct {
	Code    int
	Message string
	Details []string
}

func New(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func From(err error) *Error {
	if err == nil {
		return InternalServerError
	}

	if e, ok := err.(*Error); ok {
		return e
	}

	return New(500, err.Error())
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Response() api.SharedAPIErrorResponse {
	message := e.Message
	if e.Code == 500 {
		message = "internal server error"
	}

	return api.SharedAPIErrorResponse{
		StatusCode: e.Code,
		Details:    e.Details,
		Message:    message,
		Success:    false,
	}
}

func (e *Error) WithDetails(details interface{}) *Error {
	newError := *e
	if errors, ok := details.([]interface{}); ok {
		errs := make([]string, len(errors))
		for i, v := range errors {
			errs[i] = v.(string)
		}
		newError.Details = append(newError.Details, errs...)
		return &newError
	}

	if errors, ok := details.([]string); ok {
		newError.Details = append(newError.Details, errors...)
		return &newError
	}

	if err, ok := details.(string); ok {
		newError.Details = append(newError.Details, err)
		return &newError
	}

	if err, ok := details.(error); ok {
		newError.Details = append(newError.Details, err.Error())
		return &newError
	}

	return &newError
}

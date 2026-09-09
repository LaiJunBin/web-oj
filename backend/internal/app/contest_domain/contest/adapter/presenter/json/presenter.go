package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"reflect"
)

type JSONPresenter struct {
}

func New() JSONPresenter {
	return JSONPresenter{}
}

func (p JSONPresenter) Output(output interface{}) api.SharedAPIResponse {
	if reflect.ValueOf(output).IsNil() {
		return api.SharedAPIResponse{
			Success: true,
		}
	}

	return api.SharedAPIResponse{
		Success: true,
		Data:    output,
	}
}

func (p JSONPresenter) Error(err error) api.SharedAPIErrorResponse {
	return httperror.From(err).Response()
}

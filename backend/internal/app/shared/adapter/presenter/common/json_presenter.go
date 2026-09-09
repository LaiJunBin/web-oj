package commonpresenter

import (
	"fmt"
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"reflect"
	"regexp"
	"strings"
)

type JSONPresenter struct {
}

func NewJSONPresenter() *JSONPresenter {
	return &JSONPresenter{}
}

func (p *JSONPresenter) Output(output interface{}) api.SharedAPIResponse {
	if reflect.ValueOf(output).IsNil() {
		return api.SharedAPIResponse{
			Success: true,
		}
	}

	outputValue := reflect.ValueOf(output).Elem()
	outputType := outputValue.Type()

	fields := make([]reflect.StructField, outputType.NumField())
	for i := 0; i < outputType.NumField(); i++ {
		field := outputType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			regex := regexp.MustCompile(`([a-z0-9])([A-Z])`)
			jsonTag = regex.ReplaceAllString(field.Name, "${1}_${2}")
			jsonTag = strings.ToLower(jsonTag)
		}
		fields[i] = reflect.StructField{
			Name: field.Name,
			Type: field.Type,
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, jsonTag)),
		}
	}

	newStructType := reflect.StructOf(fields)
	newStruct := reflect.New(newStructType).Elem()
	for i := 0; i < newStruct.NumField(); i++ {
		newStruct.Field(i).Set(outputValue.Field(i))
	}

	output = newStruct.Interface()

	return api.SharedAPIResponse{
		Success: true,
		Data:    output,
	}
}

func (p *JSONPresenter) Error(err error) api.SharedAPIErrorResponse {
	return httperror.From(err).Response()
}

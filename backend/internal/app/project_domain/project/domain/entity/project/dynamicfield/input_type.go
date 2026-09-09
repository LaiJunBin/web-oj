package dynamicfield

import "errors"

var (
	ErrorInvalidInputType = errors.New("input type is invalid")

	TEXT    = declareInputType("text")
	FILE    = declareInputType("file")
	BOOLEAN = declareInputType("boolean")
)

var inputTypes = []InputType{}

func GetInputTypes() []InputType {
	return inputTypes
}

type InputType struct {
	value string
}

func newInputType(value string) (InputType, error) {
	inputType := InputType{value: value}
	if err := inputType.validate(); err != nil {
		return InputType{}, err
	}

	return inputType, nil
}

func mustNewInputType(value string) InputType {
	inputType, err := newInputType(value)
	if err != nil {
		panic(err)
	}

	return inputType
}

func declareInputType(value string) InputType {
	inputType := mustNewInputType(value)
	inputTypes = append(inputTypes, inputType)
	return inputType
}

func (i InputType) validate() error {
	if i.value == "" {
		return ErrorInvalidInputType
	}

	return nil
}

func GetTypeByValue(value string) (InputType, error) {
	for _, inputType := range inputTypes {
		if inputType.value == value {
			return inputType, nil
		}
	}

	return InputType{}, ErrorInvalidInputType
}

func (i InputType) String() string {
	return i.value
}

func (i InputType) Equals(other InputType) bool {
	return i.value == other.value
}

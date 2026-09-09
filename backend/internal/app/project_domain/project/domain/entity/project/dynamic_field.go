package project

import "judge-backend/internal/app/project_domain/project/domain/entity/project/dynamicfield"

type DynamicField struct {
	inputType    dynamicfield.InputType
	label        dynamicfield.Label
	name         dynamicfield.Name
	defaultValue *string
	isRequired   bool
}

type DynamicFieldConfig struct {
	Type       string
	Label      string
	Name       string
	DefaultVal *string
	IsRequired bool
}

func NewDynamicField(config *DynamicFieldConfig) (DynamicField, error) {
	label, err := dynamicfield.NewLabel(config.Label)
	if err != nil {
		return DynamicField{}, err
	}

	name, err := dynamicfield.NewName(config.Name)
	if err != nil {
		return DynamicField{}, err
	}

	t, err := dynamicfield.GetTypeByValue(config.Type)
	if err != nil {
		return DynamicField{}, err
	}

	return DynamicField{
		inputType:    t,
		label:        label,
		name:         name,
		defaultValue: config.DefaultVal,
		isRequired:   config.IsRequired,
	}, nil
}

func (df DynamicField) InputType() dynamicfield.InputType {
	return df.inputType
}

func (df DynamicField) Label() dynamicfield.Label {
	return df.label
}

func (df DynamicField) Name() dynamicfield.Name {
	return df.name
}

func (df DynamicField) DefaultValue() *string {
	return df.defaultValue
}

func (df DynamicField) IsRequired() bool {
	return df.isRequired
}

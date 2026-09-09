package project

import (
	sharedtypes "judge-backend/internal/app/shared/types"
)

type Project struct {
	id            ID
	name          Name
	description   Description
	image         Image
	testImage     Image
	envVars       Envs
	testEnvVars   Envs
	dynamicFields []DynamicField
	resultParser  string

	sharedtypes.TimeStamps
}

type ProjectConfig struct {
	ID            int32
	Name          string
	Description   string
	Image         *string
	TestImage     *string
	Env           string
	TestEnv       string
	DynamicFields []*DynamicFieldConfig
	ResultParser  string
}

func New(config *ProjectConfig) (*Project, error) {
	id, err := NewID(config.ID)
	if err != nil {
		return nil, err
	}

	name, err := NewName(config.Name)
	if err != nil {
		return nil, err
	}

	envVars, err := ParseEnvs(config.Env)
	if err != nil {
		return nil, err
	}

	testEnvVars, err := ParseEnvs(config.TestEnv)
	if err != nil {
		return nil, err
	}

	project := &Project{
		id:           id,
		name:         name,
		description:  NewDescription(config.Description),
		envVars:      envVars,
		testEnvVars:  testEnvVars,
		resultParser: config.ResultParser,
	}

	if err := project.SetDynamicFields(config.DynamicFields); err != nil {
		return nil, err
	}

	if config.Image != nil {
		project.image = NewImage(*config.Image)
	}

	if config.TestImage != nil {
		project.testImage = NewImage(*config.TestImage)
	}

	return project, nil
}

func MustNew(config *ProjectConfig) *Project {
	project, err := New(config)
	if err != nil {
		panic(err)
	}

	return project
}

func (p *Project) ID() ID {
	return p.id
}

func (p *Project) Name() Name {
	return p.name
}

func (p *Project) Description() Description {
	return p.description
}

func (p *Project) Image() Image {
	return p.image
}

func (p *Project) TestImage() Image {
	return p.testImage
}

func (p *Project) EnvVars() Envs {
	return p.envVars
}

func (p *Project) TestEnvVars() Envs {
	return p.testEnvVars
}

func (p *Project) DynamicFields() []DynamicField {
	return p.dynamicFields
}

func (p *Project) ResultParser() string {
	return p.resultParser
}

func (p *Project) SetResultParser(resultParser string) {
	p.resultParser = resultParser
}

func (p *Project) SetID(id int32) error {
	idValue, err := NewID(id)
	if err != nil {
		return err
	}

	p.id = idValue
	return nil
}

func (p *Project) SetName(name string) error {
	n, err := NewName(name)
	if err != nil {
		return err
	}

	p.name = n
	return nil
}

func (p *Project) SetDescription(description string) {
	p.description = NewDescription(description)
}

func (p *Project) SetImage(image string) {
	p.image = NewImage(image)
}

func (p *Project) SetTestImage(image string) {
	p.testImage = NewImage(image)
}

func (p *Project) SetEnv(env string) error {
	envVars, err := ParseEnvs(env)
	if err != nil {
		return err
	}

	p.envVars = envVars
	return nil
}

func (p *Project) SetTestEnv(env string) error {
	envVars, err := ParseEnvs(env)
	if err != nil {
		return err
	}

	p.testEnvVars = envVars
	return nil
}

func (p *Project) SetDynamicFields(configs []*DynamicFieldConfig) error {
	ds := make([]DynamicField, 0, len(configs))
	for _, dfc := range configs {
		df, err := NewDynamicField(dfc)
		if err != nil {
			return err
		}
		ds = append(ds, df)
	}

	p.dynamicFields = ds
	return nil
}

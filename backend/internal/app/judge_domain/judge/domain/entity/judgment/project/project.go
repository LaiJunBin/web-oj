package project

import "strings"

type Project struct {
	id           int32
	image        Image
	testImage    Image
	resultParser string
	limits       Limits
}

// Limits are the resource caps a judgment's containers run under, as
// configured on the contest project. A nil field means the contest project
// does not override the deployment-wide default, which only the runtime knows
// — so the fallback is applied there, and the domain carries the intent
// ("unset") rather than a number it would have to guess.
type Limits struct {
	// Pids caps processes. nil = use the deployment default, 0 = no cap.
	Pids *int32
	// MemoryMB caps memory. nil = use the deployment default, 0 = no cap.
	MemoryMB *int32
}

type ProjectConfig struct {
	ID            int32
	ImageTag      string
	TestImageTag  string
	Env           string
	TestEnv       string
	MountPath     string
	ResultParser  string
	PidsLimit     *int32
	MemoryLimitMB *int32
}

func NewProject(config *ProjectConfig) *Project {
	return &Project{
		id: config.ID,
		image: Image{
			Tag:       config.ImageTag,
			Env:       splitAndRemoveEmpty(config.Env, "\n"),
			MountPath: config.MountPath,
		},
		testImage: Image{
			Tag: config.TestImageTag,
			Env: splitAndRemoveEmpty(config.TestEnv, "\n"),
		},
		resultParser: config.ResultParser,
		limits: Limits{
			Pids:     config.PidsLimit,
			MemoryMB: config.MemoryLimitMB,
		},
	}
}

func splitAndRemoveEmpty(s string, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func (p *Project) ID() int32 {
	return p.id
}

func (p *Project) ImageTag() string {
	return p.image.Tag
}

func (p *Project) TestImageTag() string {
	return p.testImage.Tag
}

func (p *Project) Env() []string {
	return p.image.Env
}

func (p *Project) TestEnv() []string {
	return p.testImage.Env
}

func (p *Project) MountPath() string {
	return p.image.MountPath
}

func (p *Project) ResultParser() string {
	return p.resultParser
}

func (p *Project) Limits() Limits {
	return p.limits
}

package project

import (
	"errors"
	"strings"
)

var (
	ErrorInvalidEnvKey = errors.New("invalid env key, must not be empty")
	ErrorInvalidEnv    = errors.New("invalid env")
)

type Env struct {
	key   string
	value string
}

type Envs []Env

func NewEnv(key, value string) (Env, error) {
	e := Env{key: key, value: value}
	if err := e.validate(); err != nil {
		return Env{}, err
	}

	return e, nil
}

func MustNewEnv(key, value string) Env {
	e, err := NewEnv(key, value)
	if err != nil {
		panic(err)
	}

	return e
}

func ParseEnvs(env string) (Envs, error) {
	var envs []Env
	lines := strings.Split(env, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		sp := strings.SplitN(line, "=", 2)
		if len(sp) != 2 {
			return nil, ErrorInvalidEnv
		}

		env, err := NewEnv(sp[0], sp[1])
		if err != nil {
			return nil, err
		}

		envs = append(envs, env)
	}

	return envs, nil
}

func (e Env) validate() error {
	if e.key == "" {
		return ErrorInvalidEnvKey
	}

	return nil
}

func (e Env) Key() string {
	return e.key
}

func (e Env) Value() string {
	return e.value
}

func (e Env) Equals(other Env) bool {
	return e == other
}

func (e Env) String() string {
	return e.key + "=" + e.value
}

func (envs Envs) String() string {
	var sb strings.Builder
	for _, env := range envs {
		sb.WriteString(env.String())
		sb.WriteString("\n")
	}

	return sb.String()
}

func (envs Envs) Strings() []string {
	var ss []string
	for _, env := range envs {
		ss = append(ss, env.String())
	}

	return ss
}

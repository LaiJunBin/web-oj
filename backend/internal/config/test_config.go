package config

import (
	"path/filepath"
	"runtime"
)

// NewTestConfig loads the TRACKED unit-test fixture,
// internal/config/config/config_test.yaml. Unit tests are hermetic: they
// depend on nothing generated and on no developer-local file. The path is
// anchored to this source file rather than the CWD, because go test runs
// every package from its own directory.
func NewTestConfig() *Config {
	_, self, _, _ := runtime.Caller(0)
	configRootDir = filepath.Join(filepath.Dir(self), "config")
	configName = "config_test"
	return NewConfig()
}

// NewFeatureConfig loads the GENERATED config_test.yaml at the backend root.
// The feature suites' scripts write real local values into it before running,
// which is why that file deliberately stays outside version control — and why
// unit tests must not read it (they did, once: a fresh clone panicked).
func NewFeatureConfig() *Config {
	configRootDir = rootDir()
	configName = "config_test"
	return NewConfig()
}

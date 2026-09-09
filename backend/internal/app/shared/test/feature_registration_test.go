package test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"judge-backend/internal/projectpath"
)

// Feature files are not discovered — each suite's *_test.go names the features
// it runs, one TestFeature call per file. A .feature added without that call
// does not fail; it does NOTHING, and with TARGET_TEST filtering the rest away
// the run prints "ok" over zero scenarios. That exact combination shipped a
// bundle feature that had never executed while its endpoint returned 501.
//
// This test lives under internal/ because that is what CI runs as unit tests;
// the feature suites themselves need containers and are a separate job.
func TestEveryFeatureFileIsRegistered(t *testing.T) {
	featuresRoot := filepath.Join(projectpath.Root, "tests", "features")

	// TestFeature(t, "name", ...) — the name is the .feature file's basename.
	registered := regexp.MustCompile(`TestFeature\(t,\s*"([^"]+)"`)

	err := filepath.Walk(featuresRoot, func(dir string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || filepath.Base(dir) != "features" {
			return err
		}

		suiteDir := filepath.Dir(dir)

		names := map[string]bool{}
		entries, err := os.ReadDir(suiteDir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			body, err := os.ReadFile(filepath.Join(suiteDir, e.Name()))
			if err != nil {
				return err
			}
			for _, m := range registered.FindAllStringSubmatch(string(body), -1) {
				names[m[1]] = true
			}
		}

		featureFiles, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, f := range featureFiles {
			if !strings.HasSuffix(f.Name(), ".feature") {
				continue
			}
			name := strings.TrimSuffix(f.Name(), ".feature")
			if !names[name] {
				t.Errorf("%s/features/%s.feature 存在，但 %s 底下沒有任何 _test.go 以 TestFeature 註冊它 — 這個檔案永遠不會被執行",
					suiteDir, name, suiteDir)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", featuresRoot, err)
	}
}

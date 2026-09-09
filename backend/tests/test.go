package main

import (
	"fmt"
	"judge-backend/internal/config"
	"judge-backend/tests/setup"
	"os"
	"os/exec"
)

func main() {
	// os.Exit skips deferred cleanup, so run the suite in a helper
	// and exit with its code afterwards.
	os.Exit(run())
}

func run() int {
	args := os.Args[1:]
	// scenarios reset the database each run, so suites easily exceed
	// go test's default 10m timeout
	goTestArgs := []string{"test", "-failfast", "-p", "1", "-timeout", "60m"}
	if os.Getenv("FEATURE_TEST_VERBOSE") == "1" {
		goTestArgs = append(goTestArgs, "-v")
	}

	config := config.NewConfig()
	setup := setup.NewSetup(config)

	defer setup.TerminateInfrastructure()
	if err := setup.SetupInfrastructure(); err != nil {
		panic(err)
	}

	cmd := exec.Command("go", append(goTestArgs, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

package testutils

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type Context struct {
	Context *godog.ScenarioContext
	API     *APIFeature
	DB      *gorm.DB
	RDB     *redis.Client
	MDB     *mongo.Client
}

type ScenarioInitializer interface {
	InitializeScenario(ctx *Context)
}

func TestFeature(t *testing.T, feature string, initializer ScenarioInitializer) {
	targetTest := getEnv("TARGET_TEST")
	if targetTest != "" && targetTest != feature {
		fmt.Println("Skipping test for feature: " + feature)
		return
	}

	wd, _ := os.Getwd()
	path := wd + "/features/" + feature + ".feature"

	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			ctx := &Context{
				Context: sc,
				API:     NewAPIFeature(),
				DB:      GetDB(),
				RDB:     GetRedisClient(),
				MDB:     GetMongoDB(),
			}
			initializer.InitializeScenario(ctx)
		},
		Options: &godog.Options{
			Format:        "pretty",
			Paths:         []string{path},
			TestingT:      t,
			Strict:        true,
			StopOnFailure: true,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return ""
	}
	return value
}

package testutils

import (
	"context"

	"judge-backend/database"
	"judge-backend/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
)

var mdb = database.NewMongoDB(config.NewFeatureConfig())

func GetMongoDB() *mongo.Client {
	return mdb
}

func ResetMongoDB() {
	mdb.Database("judge").Drop(context.Background())
}

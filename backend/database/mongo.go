package database

import (
	"context"
	"fmt"
	"judge-backend/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoDB(config *config.Config) *mongo.Client {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/",
		config.Database.Mongo.Username,
		config.Database.Mongo.Password,
		config.Database.Mongo.Host,
		config.Database.Mongo.Port,
	)

	ctx := context.Background()
	mdb, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	return mdb
}

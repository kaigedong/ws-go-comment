package dbops

import (
	"context"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	DBCollections = &dbCollection{}
)

type dbCollection struct {
	UserComment        *mongo.Collection
	Projects           *mongo.Collection
	ProjectSubscribers *mongo.Collection
}

func InitDBCollection() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbClient, err := mongo.Connect(ctx, options.Client().ApplyURI(viper.GetString("mongodbURL")))
	if err != nil {
		logrus.Error(os.Stderr, err)
		os.Exit(1)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = dbClient.Ping(ctx, readpref.Primary()); err != nil {
		logrus.Error(os.Stderr, err)
		os.Exit(1)
	}

	// Define collections
	dadaDB := dbClient.Database(viper.GetString("mongodbName"))
	DBCollections.UserComment = dadaDB.Collection("user_comments")
	DBCollections.Projects = dadaDB.Collection("projects")
	DBCollections.ProjectSubscribers = dadaDB.Collection("project_subscriber")
}

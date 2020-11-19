package comment_package

import (
	"comment/dbops"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AllComments(projectUniqID string) ([]*UserComment, error) {

	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()
	options := options.Find()
	options.SetSort(bson.D{{"id", -1}})

	cursor, err := dbops.DBCollections.UserComment.Find(ctx, bson.M{"project_uniq_id": projectUniqID}, options)
	if err != nil {
		return nil, err
	}

	out := []*UserComment{}
	for cursor.Next(context.TODO()) {
		var elem UserComment
		if err := cursor.Decode(&elem); err != nil {
			logrus.WithFields(logrus.Fields{"Decode": cursor}).Error(err)
			continue
		}
		out = append(out, &elem)
	}

	return out, nil
}

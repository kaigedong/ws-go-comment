package comment_package

import (
	"context"
	"crypto/rand"
	"math"
	"math/big"
	"sync"
	"time"

	"comment/dbops"

	"github.com/sirupsen/logrus"
)

var (
	UserComments          = make(chan *UserComment, 1024)
	newCommentSubscribers sync.Map // i64 -> chan *UserComment
)

type UserComment struct {
	ProjectUniqID string `json:"project_uniq_id" query:"project_uniq_id" form:"project_uniq_id" bson:"project_uniq_id"`
	UserAddr      string `json:"user_addr" query:"user_addr" form:"user_addr" bson:"user_addr"`
	Comment       string `json:"comment" query:"comment" form:"comment" bson:"comment"`
	Timestamp     string `json:"timestamp" query:"timestamp" form:"timestamp" bson:"timestamp"`
}

func CommentDaemon() {
	go func() {
		for {
			if err := addNewComment(UserComments); err != nil {
				logrus.Error(err)
			}
		}
	}()
}

func addNewComment(uc <-chan *UserComment) error {

	for userComment := range uc {
		// Store userComment to db
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := dbops.DBCollections.UserComment.InsertOne(ctx, userComment); err != nil {
			return err
		}

		// Write userComment to ws
		newCommentSubscribers.Range(func(key, value interface{}) bool {
			value.(chan *UserComment) <- userComment
			return true
		})
	}
	return nil
}

func randint64() (int64, error) {
	val, err := rand.Int(rand.Reader, big.NewInt(int64(math.MaxInt64)))
	if err != nil {
		return 0, err
	}
	return val.Int64(), nil
}

func SubscribeNewComment() (<-chan *UserComment, int64) {
	var id int64
	c := make(chan *UserComment, 64)
	for {
		randomID, err := randint64()
		if err != nil {
			continue
		}
		if _, ok := newCommentSubscribers.Load(randomID); !ok {
			break
		}
	}
	newCommentSubscribers.Store(id, c)

	return c, id
}

func UnsubscribeNewComment(id int64) {
	if ch, ok := newCommentSubscribers.Load(id); ok {
		close(ch.(chan *UserComment))
		newCommentSubscribers.Delete(id)
	}
}

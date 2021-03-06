package mongo

import (
	"context"
	"github.com/pkg/errors"
	"github.com/vctrl/terpila-bot/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TolerancesMongo struct {
	col *mongo.Collection
}

func (t *TolerancesMongo) Add(ctx context.Context, tol *db.Tolerance) error {
	_, err := t.col.InsertOne(ctx, tol)
	if err != nil {
		return errors.Wrap(err, "insert tolerance")
	}

	return nil
}

func (t *TolerancesMongo) GetCountByUser(ctx context.Context, userID int64) (int64, error) {
	cnt, err := t.col.CountDocuments(ctx, bson.M{"_id": userID})
	if err != nil {
		return 0, errors.Wrap(err, "count tolerances")
	}

	return cnt, nil
}

func NewTolerancesMongo(client *mongo.Client) db.Tolerances {
	return &TolerancesMongo{
		col: client.Database(db.DBName).Collection(db.TolerancesColName),
	}
}

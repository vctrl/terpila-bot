package mongo

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/vctrl/terpila-bot/db"
)

type TerpiloidMongo struct {
	db *sql.DB
}

func (t *TerpiloidMongo) Add(context.Context, *db.Terpiloid) error {
	return errors.New("not implemented")
}

func (t *TerpiloidMongo) Get(context.Context) (*db.Terpiloid, error) {
	return nil, errors.New("not implemented")
}
func NewTerpiloidsMongo() db.Terpiloids {
	return &TerpiloidMongo{}
}

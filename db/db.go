package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

var (
	TolerancesColName = "tolerances"
	TerpiloidsColName = "terpiloids"
	DBName            = "terpila_bot"
)

type Terpiloid struct {
}

type Tolerance struct {
	ID        uuid.UUID
	UserID    int64
	CreatedAt time.Time
}

type Terpiloids interface {
	Add(context.Context, *Terpiloid) error
	Get(context.Context) (*Terpiloid, error)
}

type Tolerances interface {
	Add(ctx context.Context, tol *Tolerance) error
	GetCountByUser(ctx context.Context, userID int64) (int64, error)
}

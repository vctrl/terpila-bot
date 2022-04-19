package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vctrl/terpila-bot/db"
)

func NewTolerance(id uuid.UUID, userID int64) *db.Tolerance {
	return &db.Tolerance{
		ID:        id,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
}

type TolerancesMemory struct {
	data []*db.Tolerance
}

func (t *TolerancesMemory) Add(ctx context.Context, tol *db.Tolerance) error {
	t.data = append(t.data, tol)

	return nil
}

func (t *TolerancesMemory) GetCountByUser(ctx context.Context, userID int64) (int64, error) {
	cnt := int64(0)
	for _, t := range t.data {
		if t.UserID == userID {
			cnt++
		}
	}

	return cnt, nil
}

func NewTolerancesMemory() db.Tolerances {
	return &TolerancesMemory{
		data: make([]*db.Tolerance, 0, 10),
	}
}

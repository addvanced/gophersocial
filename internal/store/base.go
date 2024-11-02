package store

import "time"

type BaseEntityer interface {
	GetID() int64
	GetCreatedAt() time.Time
}

type BaseEntity struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (b BaseEntity) GetID() int64 {
	return b.ID
}

func (b BaseEntity) GetCreatedAt() time.Time {
	return b.CreatedAt
}

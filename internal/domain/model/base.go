package model

import "time"

type Entity interface {
	GetID() string
	SetID(id string)
	SetCreatedAt(t time.Time)
	SetUpdatedAt(t time.Time)
}

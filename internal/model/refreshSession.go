package model

import "time"

type RefreshSession struct {
	UserID    int64
	Token     string
	ExpiresIn time.Time
}

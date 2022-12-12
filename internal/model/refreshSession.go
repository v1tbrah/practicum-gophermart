package model

import "time"

type RefreshSession struct {
	ExpiresIn time.Time
	Token     string
	UserID    int64
}

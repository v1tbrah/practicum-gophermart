package model

type RefreshSession struct {
	UserID    int64
	Token     string
	ExpiresIn int64
}

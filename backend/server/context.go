package server

import (
	"context"
)

type contextKey string

const userIDKey contextKey = "userID"

func contextWithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

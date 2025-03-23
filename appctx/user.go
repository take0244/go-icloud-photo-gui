package appctx

import (
	"context"
)

const (
	userKey contextKey = "user"
)

type ContextUser struct {
	ID string
}

func WithUser(ctx context.Context, usr ContextUser) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

func User(ctx context.Context) *ContextUser {
	usr, ok := ctx.Value(userKey).(ContextUser)
	if ok {
		return &usr
	}

	return nil
}

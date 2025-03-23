package appctx

import (
	"context"
)

type contextKey string

func NewAppContext() context.Context {
	ctx := context.Background()
	return ctx
}

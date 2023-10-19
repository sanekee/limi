package limi

import (
	"context"
	"sync"
)

type ctxKey string

var limiContextKey ctxKey = "limi context"

type limiContext struct {
	URLParams map[string]string
	lock      sync.RWMutex
}

func NewContext(ctx context.Context) context.Context {
	lCtx := &limiContext{
		URLParams: make(map[string]string),
	}
	return context.WithValue(ctx, limiContextKey, lCtx)
}

func GetURLParam(ctx context.Context, key string) string {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return ""
	}
	lCtx.lock.RLock()
	defer lCtx.lock.RUnlock()

	return lCtx.URLParams[key]
}

func SetURLParam(ctx context.Context, key, val string) {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return
	}
	lCtx.lock.Lock()
	defer lCtx.lock.Unlock()

	lCtx.URLParams[key] = val
}

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

	routingPath string
	router      any
}

func NewContext(ctx context.Context) context.Context {
	_, ok := ctx.Value(limiContextKey).(*limiContext)
	if ok {
		return ctx
	}

	lCtx := &limiContext{
		URLParams: make(map[string]string),
	}
	return context.WithValue(ctx, limiContextKey, lCtx)
}

func IsContextSet(ctx context.Context) bool {
	_, ok := ctx.Value(limiContextKey).(*limiContext)
	return ok
}

func Reset(ctx context.Context) {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return
	}

	lCtx.lock.Lock()
	defer lCtx.lock.Unlock()

	for k := range lCtx.URLParams {
		delete(lCtx.URLParams, k)
	}

	lCtx.routingPath = ""
	lCtx.router = nil
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

func GetRoutingPath(ctx context.Context) string {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return ""
	}
	lCtx.lock.RLock()
	defer lCtx.lock.RUnlock()

	return lCtx.routingPath
}

func SetRoutingPath(ctx context.Context, path string) {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return
	}
	lCtx.lock.Lock()
	defer lCtx.lock.Unlock()

	lCtx.routingPath = path
}

func GetRouter(ctx context.Context) any {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return ""
	}
	lCtx.lock.RLock()
	defer lCtx.lock.RUnlock()

	return lCtx.router
}

func SetRouter(ctx context.Context, router any) {
	lCtx, ok := ctx.Value(limiContextKey).(*limiContext)
	if !ok {
		return
	}
	lCtx.lock.Lock()
	defer lCtx.lock.Unlock()

	lCtx.router = router
}

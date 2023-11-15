package common

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type CtxKey string

func WithLogger(ctx context.Context, logger *log.Entry) context.Context {
	return context.WithValue(ctx, CtxKey("logger"), logger)
}

func GetLogger(ctx context.Context) *log.Entry {
	logger, ok := ctx.Value(CtxKey("logger")).(*log.Entry)
	if !ok {
		return log.WithField("daemon", "kubedtnd")
	}
	return logger
}

func WithCtxValue(ctx context.Context, key string, value interface{}) context.Context {
	return context.WithValue(ctx, CtxKey(key), value)
}

func GetCtxValue(ctx context.Context, key string) interface{} {
	return ctx.Value(CtxKey(key))
}

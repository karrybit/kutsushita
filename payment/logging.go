package payment

import (
	"time"

	"go.uber.org/zap"
)

type loggingMiddleware struct {
	next   Service
	logger *zap.Logger
}

func LoggingMiddleware(logger *zap.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{next: next, logger: logger}
	}
}

func (mw *loggingMiddleware) Authorise(amount float32) (auth Authorisation, err error) {
	defer func(begin time.Time) {
		mw.logger.Info("method Authorise", zap.Bool("result", auth.Authorised), zap.Duration("took", time.Since(begin)))
	}(time.Now())
	return mw.next.Authorise(amount)
}

func (mw *loggingMiddleware) Health() (health []Health) {
	defer func(begin time.Time) {
		mw.logger.Info("method Health", zap.Int("result", len(health)), zap.Duration("took", time.Since(begin)))
	}(time.Now())
	return mw.next.Health()
}

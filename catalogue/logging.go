package catalogue

import (
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) Middleware {
	return func(next Service) Service {
		return loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger *zap.Logger
}

func (mw loggingMiddleware) List(tags []string, order string, pageNum, pageSize int) (socks []Sock, err error) {
	defer func(begin time.Time) {
		mw.logger.Info("method List", zap.Strings("tags", tags), zap.String("order", order), zap.Int("pageNum", pageNum), zap.Int("pageSize", pageSize), zap.Int("result", len(socks)), zap.Error(err), zap.Duration("took", time.Since(begin)))
	}(time.Now())
	return mw.next.List(tags, order, pageNum, pageSize)
}

func (mw loggingMiddleware) Count(tags []string) (n int, err error) {
	defer func(begin time.Time) {
		mw.logger.Info("method Count", zap.Strings("tags", tags), zap.Int("result", n), zap.Error(err), zap.Duration("took", time.Since(begin)))
	}(time.Now())
	return mw.next.Count(tags)
}

func (mw loggingMiddleware) Get(id string) (s Sock, err error) {
	defer func(begin time.Time) {
		mw.logger.Info("method Get", zap.String("id", id), zap.String("sock", s.ID), zap.Error(err), zap.Duration("took", time.Since(begin)))
	}(time.Now())
	return mw.next.Get(id)
}

func (mw loggingMiddleware) Tags() (tags []string, err error) {
	defer func(begin time.Time) {
		mw.logger.Info("method Tags", zap.Int("result", len(tags)), zap.Error(err), zap.Duration("took", time.Since(begin)))
	}(time.Now())
	return mw.next.Tags()
}

func (mw loggingMiddleware) Health() (health []Health) {
	defer func(begin time.Time) {
		mw.logger.Info("method Health", zap.Int("result", len(health)), zap.Duration("took", time.Since(begin)))
	}(time.Now())
	return mw.next.Health()
}

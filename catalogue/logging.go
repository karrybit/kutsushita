package catalogue

import (
	"log"
	"strings"
	"time"
)

func LoggingMiddleware(logger *log.Logger) Middleware {
	return func(next Service) Service {
		return loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger *log.Logger
}

func (mw loggingMiddleware) List(tags []string, order string, pageNum, pageSize int) (socks []Sock, err error) {
	defer func(begin time.Time) {
		mw.logger.Println("method", "List", "tags", strings.Join(tags, ", "), "order", order, "pageNum", pageNum, "pageSize", pageSize, "result", len(socks), "err", err, "took", time.Since(begin))
	}(time.Now())
	return mw.next.List(tags, order, pageNum, pageSize)
}

func (mw loggingMiddleware) Count(tags []string) (n int, err error) {
	defer func(begin time.Time) {
		mw.logger.Println("method", "Count", "tags", strings.Join(tags, ", "), "result", n, "err", err, "took", time.Since(begin))
	}(time.Now())
	return mw.next.Count(tags)
}

func (mw loggingMiddleware) Get(id string) (s Sock, err error) {
	defer func(begin time.Time) {
		mw.logger.Println("method", "Get", "id", id, "sock", s.ID, "err", err, "took", time.Since(begin))
	}(time.Now())
	return mw.next.Get(id)
}

func (mw loggingMiddleware) Tags() (tags []string, err error) {
	defer func(begin time.Time) {
		mw.logger.Println("method", "Tags", "result", len(tags), "err", err, "took", time.Since(begin))
	}(time.Now())
	return mw.next.Tags()
}

func (mw loggingMiddleware) Health() (health []Health) {
	defer func(begin time.Time) {
		mw.logger.Println("method", "Health", "result", len(health), "took", time.Since(begin))
	}(time.Now())
	return mw.next.Health()
}

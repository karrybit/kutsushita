package payment

import (
	"log"
	"time"
)

type loggingMiddleware struct {
	next   Service
	logger *log.Logger
}

func LoggingMiddleware(logger *log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{next: next, logger: logger}
	}
}

func (mw *loggingMiddleware) Authorise(amount float32) (auth Authorisation, err error) {
	defer func(begin time.Time) {
		mw.logger.Println("method", "Authorise", "result", auth.Authorised, "took", time.Since(begin))
	}(time.Now())
	return mw.next.Authorise(amount)
}

func (mw *loggingMiddleware) Health() (health []Health) {
	defer func(begin time.Time) {
		mw.logger.Println("method", "Health", "result", len(health), "took", time.Since(begin))
	}(time.Now())
	return mw.next.Health()
}

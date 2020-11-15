package api

import (
	"context"
	"time"
	"user/users"

	"go.uber.org/zap"
)

// Middleware decorates a service.
type Middleware func(Service) Service

// LoggingMiddleware logs method calls, parameters, results, and elapsed time.
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

func (mw loggingMiddleware) Login(ctx context.Context, username string, password string) (user users.User, err error) {
	defer func(begin time.Time) {
		mw.logger.Info(
			"method Login",
			zap.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return mw.next.Login(ctx, username, password)
}

func (mw loggingMiddleware) Register(ctx context.Context, username string, password string, email string, first string, last string) (string, error) {
	defer func(begin time.Time) {
		mw.logger.Info(
			"method Register",
			zap.String("username", username),
			zap.String("email", email),
			zap.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return mw.next.Register(ctx, username, password, email, first, last)
}

func (mw loggingMiddleware) PostUser(ctx context.Context, user users.User) (id string, err error) {
	defer func(begin time.Time) {
		mw.logger.Info(
			"method PostUser",
			zap.String("username", user.Username),
			zap.String("email", user.Email),
			zap.String("result", id),
			zap.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return mw.next.PostUser(ctx, user)
}

func (mw loggingMiddleware) GetUsers(ctx context.Context, id string) (users []users.User, err error) {
	defer func(begin time.Time) {
		who := id
		if who == "" {
			who = "all"
		}
		mw.logger.Info(
			"method GetUsers",
			zap.String("id", who),
			zap.Int("result", len(users)),
			zap.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return mw.next.GetUsers(ctx, id)
}

func (mw loggingMiddleware) PostAddress(ctx context.Context, userAddress users.Address, id string) (string, error) {
	defer func(begin time.Time) {
		mw.logger.Info(
			"method PostAddress",
			zap.String("streed", userAddress.Street),
			zap.String("number", userAddress.Number),
			zap.String("id", id),
			zap.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return mw.next.PostAddress(ctx, userAddress, id)
}

func (mw loggingMiddleware) GetAddresses(ctx context.Context, id string) (userAddresses []users.Address, err error) {
	defer func(begin time.Time) {
		who := id
		if who == "" {
			who = "all"
		}
		mw.logger.Info(
			"method GetAddress",
			zap.String("id", who),
			zap.Int("result", len(userAddresses)),
			zap.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return mw.next.GetAddresses(ctx, id)
}

func (mw loggingMiddleware) PostCard(ctx context.Context, userCard users.Card, id string) (string, error) {
	defer func(begin time.Time) {
		cc := userCard
		cc.MaskCC()
		mw.logger.Info(
			"method PostCard",
			zap.String("card", cc.LongNum),
			zap.String("user", id),
			zap.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return mw.next.PostCard(ctx, userCard, id)
}

func (mw loggingMiddleware) GetCards(ctx context.Context, id string) (userCards []users.Card, err error) {
	defer func(begin time.Time) {
		who := id
		if who == "" {
			who = "all"
		}
		mw.logger.Info(
			"method GetCards",
			zap.String("id", who),
			zap.Int("result", len(userCards)),
			zap.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return mw.next.GetCards(ctx, id)
}

func (mw loggingMiddleware) Delete(ctx context.Context, entity string, id string) (err error) {
	defer func(begin time.Time) {
		mw.logger.Info(
			"method Delete",
			zap.String("entity", entity),
			zap.String("id", id),
			zap.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return mw.next.Delete(ctx, entity, id)
}

func (mw loggingMiddleware) Health(ctx context.Context) (health []Health) {
	defer func(begin time.Time) {
		mw.logger.Info(
			"method Health",
			zap.Int("result", len(health)),
			zap.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return mw.next.Health(ctx)
}

// type instrumentingService struct {
// 	requestCount   metrics.Counter
// 	requestLatency metrics.Histogram
// 	Service
// }

// // NewInstrumentingService returns an instance of an instrumenting Service.
// func NewInstrumentingService(requestCount metrics.Counter, requestLatency metrics.Histogram, s Service) Service {
// 	return &instrumentingService{
// 		requestCount:   requestCount,
// 		requestLatency: requestLatency,
// 		Service:        s,
// 	}
// }

// func (s *instrumentingService) Login(username, password string) (users.User, error) {
// 	defer func(begin time.Time) {
// 		s.requestCount.With("method", "login").Add(1)
// 		s.requestLatency.With("method", "login").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return s.Service.Login(username, password)
// }

// func (s *instrumentingService) Register(username, password, email, first, last string) (string, error) {
// 	defer func(begin time.Time) {
// 		s.requestCount.With("method", "register").Add(1)
// 		s.requestLatency.With("method", "register").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return s.Service.Register(username, password, email, first, last)
// }

// func (s *instrumentingService) PostUser(user users.User) (string, error) {
// 	defer func(begin time.Time) {
// 		s.requestCount.With("method", "postUser").Add(1)
// 		s.requestLatency.With("method", "postUser").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return s.Service.PostUser(user)
// }

// func (s *instrumentingService) GetUsers(id string) (u []users.User, err error) {
// 	defer func(begin time.Time) {
// 		s.requestCount.With("method", "getUsers").Add(1)
// 		s.requestLatency.With("method", "getUsers").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return s.Service.GetUsers(id)
// }

// func (s *instrumentingService) PostAddress(add users.Address, id string) (string, error) {
// 	defer func(begin time.Time) {
// 		s.requestCount.With("method", "postAddress").Add(1)
// 		s.requestLatency.With("method", "postAddress").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return s.Service.PostAddress(add, id)
// }

// func (s *instrumentingService) GetAddresses(id string) ([]users.Address, error) {
// 	defer func(begin time.Time) {
// 		s.requestCount.With("method", "getAddresses").Add(1)
// 		s.requestLatency.With("method", "getAddresses").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return s.Service.GetAddresses(id)
// }

// func (s *instrumentingService) PostCard(card users.Card, id string) (string, error) {
// 	defer func(begin time.Time) {
// 		s.requestCount.With("method", "postCard").Add(1)
// 		s.requestLatency.With("method", "postCard").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return s.Service.PostCard(card, id)
// }

// func (s *instrumentingService) GetCards(id string) ([]users.Card, error) {
// 	defer func(begin time.Time) {
// 		s.requestCount.With("method", "getCards").Add(1)
// 		s.requestLatency.With("method", "getCards").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return s.Service.GetCards(id)
// }

// func (s *instrumentingService) Delete(entity, id string) error {
// 	defer func(begin time.Time) {
// 		s.requestCount.With("method", "delete").Add(1)
// 		s.requestLatency.With("method", "delete").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return s.Service.Delete(entity, id)
// }

// func (s *instrumentingService) Health() []Health {
// 	defer func(begin time.Time) {
// 		s.requestCount.With("method", "health").Add(1)
// 		s.requestLatency.With("method", "health").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return s.Service.Health()
// }

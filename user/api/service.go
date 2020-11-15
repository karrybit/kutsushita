package api

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"time"

	"user/db"
	"user/users"
)

var (
	ErrUnauthorized = errors.New("Unauthorized")
)

// Service is the user service, providing operations for users to login, register, and retrieve customer information.
type Service interface {
	Login(ctx context.Context, username string, password string) (users.User, error) // GET /login
	Register(ctx context.Context, username string, password string, email string, first string, last string) (string, error)
	GetUsers(ctx context.Context, id string) ([]users.User, error)
	PostUser(ctx context.Context, user users.User) (string, error)
	GetAddresses(ctx context.Context, id string) ([]users.Address, error)
	PostAddress(ctx context.Context, userAddress users.Address, userID string) (string, error)
	GetCards(ctx context.Context, id string) ([]users.Card, error)
	PostCard(ctx context.Context, userCard users.Card, userID string) (string, error)
	Delete(ctx context.Context, entity string, id string) error
	Health(ctx context.Context) []Health // GET /health
}

// NewFixedService returns a simple implementation of the Service interface,
func NewFixedService() Service {
	return &fixedService{}
}

type fixedService struct{}

type Health struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Time    string `json:"time"`
}

func (s *fixedService) Login(ctx context.Context, username string, password string) (users.User, error) {
	user, err := db.GetUserByName(ctx, username)
	if err != nil {
		return users.New(), err
	}
	if user.Password != calculatePassHash(password, user.Salt) {
		return users.New(), ErrUnauthorized
	}

	db.GetUserAttributes(ctx, &user)
	user.MaskCCs()
	return user, nil
}

func (s *fixedService) Register(ctx context.Context, username string, password string, email string, first string, last string) (string, error) {
	user := users.New()
	user.Username = username
	user.Password = calculatePassHash(password, user.Salt)
	user.Email = email
	user.FirstName = first
	user.LastName = last
	err := db.CreateUser(ctx, &user)
	return user.UserID, err
}

func (s *fixedService) GetUsers(ctx context.Context, id string) ([]users.User, error) {
	if id == "" {
		users, err := db.GetUsers(ctx)
		for i, user := range users {
			user.AddLinks()
			users[i] = user
		}
		return users, err
	}
	user, err := db.GetUser(ctx, id)
	user.AddLinks()
	return []users.User{user}, err
}

func (s *fixedService) PostUser(ctx context.Context, user users.User) (string, error) {
	user.NewSalt()
	user.Password = calculatePassHash(user.Password, user.Salt)
	err := db.CreateUser(ctx, &user)
	return user.UserID, err
}

func (s *fixedService) GetAddresses(ctx context.Context, id string) ([]users.Address, error) {
	if id == "" {
		addresses, err := db.GetAddresses(ctx)
		for i, address := range addresses {
			address.AddLinks()
			addresses[i] = address
		}
		return addresses, err
	}
	address, err := db.GetAddress(ctx, id)
	address.AddLinks()
	return []users.Address{address}, err
}

func (s *fixedService) PostAddress(ctx context.Context, userAddress users.Address, userID string) (string, error) {
	err := db.CreateAddress(ctx, &userAddress, userID)
	return userAddress.ID, err
}

func (s *fixedService) GetCards(ctx context.Context, id string) ([]users.Card, error) {
	if id == "" {
		cards, err := db.GetCards(ctx)
		for i, card := range cards {
			card.AddLinks()
			cards[i] = card
		}
		return cards, err
	}
	card, err := db.GetCard(ctx, id)
	card.AddLinks()
	return []users.Card{card}, err
}

func (s *fixedService) PostCard(ctx context.Context, userCard users.Card, userID string) (string, error) {
	err := db.CreateCard(ctx, &userCard, userID)
	return userCard.ID, err
}

func (s *fixedService) Delete(ctx context.Context, entity string, id string) error {
	return db.Delete(ctx, entity, id)
}

func (s *fixedService) Health(ctx context.Context) []Health {
	var health []Health
	dbstatus := "OK"

	if err := db.Ping(ctx); err != nil {
		dbstatus = "err"
	}

	app := Health{"user", "OK", time.Now().String()}
	db := Health{"user-db", dbstatus, time.Now().String()}

	health = append(health, app)
	health = append(health, db)

	return health
}

func calculatePassHash(pass, salt string) string {
	h := sha1.New()
	io.WriteString(h, salt)
	io.WriteString(h, pass)
	return fmt.Sprintf("%x", h.Sum(nil))
}

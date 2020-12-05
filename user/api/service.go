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
	Login(ctx context.Context, username string, password string) (*users.User, error) // GET /login
	Register(ctx context.Context, username string, password string, email string, first string, last string) (string, error)
	GetUsers(ctx context.Context, id string) (*[]*users.User, error)
	PostUser(ctx context.Context, user *users.User) (string, error)
	GetAddresses(ctx context.Context, id string) (*[]*users.Address, error)
	PostAddress(ctx context.Context, userAddress *users.Address, userID string) (string, error)
	GetCards(ctx context.Context, id string) (*[]*users.Card, error)
	PostCard(ctx context.Context, userCard *users.Card, userID string) (string, error)
	Delete(ctx context.Context, entity string, id string) error
	Health(ctx context.Context) *[]*Health // GET /health
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

func (s *fixedService) Login(ctx context.Context, username string, password string) (*users.User, error) {
	user, err := db.GetUserByName(ctx, username)
	if err != nil {
		return nil, err
	}
	pass, err := calculatePassHash(password, user.Salt)
	if err != nil {
		return nil, err
	}
	if user.Password != pass {
		return nil, ErrUnauthorized
	}

	if err := db.GetUserAttributes(ctx, user); err != nil {
		return nil, err
	}

	user.MaskCCs()
	return user, nil
}

func (s *fixedService) Register(ctx context.Context, username string, password string, email string, first string, last string) (string, error) {
	user := users.New()
	user.Username = username
	pass, err := calculatePassHash(password, user.Salt)
	if err != nil {
		return "", err
	}
	user.Password = pass
	user.Email = email
	user.FirstName = first
	user.LastName = last
	if err := db.CreateUser(ctx, &user); err != nil {
		return "", err
	}

	return user.UserID, nil
}

func (s *fixedService) GetUsers(ctx context.Context, id string) (us *[]*users.User, err error) {
	if id == "" {
		if us, err = db.GetUsers(ctx); err != nil {
			return nil, err
		}
	} else {
		user, err := db.GetUser(ctx, id)
		if err != nil {
			return nil, err
		}
		*us = append(*us, user)
	}

	for i := range *us {
		(*us)[i].AddLinks()
	}
	return us, err
}

func (s *fixedService) PostUser(ctx context.Context, user *users.User) (string, error) {
	user.NewSalt()
	pass, err := calculatePassHash(user.Password, user.Salt)
	if err != nil {
		return "", err
	}
	user.Password = pass
	if err := db.CreateUser(ctx, user); err != nil {
		return "", err
	}

	return user.UserID, nil
}

func (s *fixedService) GetAddresses(ctx context.Context, id string) (addresses *[]*users.Address, err error) {
	if id == "" {
		if addresses, err = db.GetAddresses(ctx); err != nil {
			return addresses, err
		}
	} else {
		address, err := db.GetAddress(ctx, id)
		if err != nil {
			return addresses, err
		}
		*addresses = append(*addresses, address)
	}

	for i := range *addresses {
		(*addresses)[i].AddLinks()
	}

	return addresses, nil
}

func (s *fixedService) PostAddress(ctx context.Context, userAddress *users.Address, userID string) (string, error) {
	err := db.CreateAddress(ctx, userAddress, userID)
	return userAddress.ID, err
}

func (s *fixedService) GetCards(ctx context.Context, id string) (cards *[]*users.Card, err error) {
	if id == "" {
		if cards, err = db.GetCards(ctx); err != nil {
			return nil, err
		}
	} else {
		card, err := db.GetCard(ctx, id)
		if err != nil {
			return nil, err
		}
		*cards = append(*cards, card)
	}

	for i := range *cards {
		(*cards)[i].AddLinks()
	}
	return cards, nil
}

func (s *fixedService) PostCard(ctx context.Context, userCard *users.Card, userID string) (string, error) {
	err := db.CreateCard(ctx, userCard, userID)
	return userCard.ID, err
}

func (s *fixedService) Delete(ctx context.Context, entity string, id string) error {
	return db.Delete(ctx, entity, id)
}

func (s *fixedService) Health(ctx context.Context) *[]*Health {
	health := new([]*Health)
	dbstatus := "OK"

	if err := db.Ping(ctx); err != nil {
		dbstatus = "err"
	}

	app := &Health{"user", "OK", time.Now().String()}
	db := &Health{"user-db", dbstatus, time.Now().String()}

	*health = append(*health, app)
	*health = append(*health, db)

	return health
}

func calculatePassHash(pass, salt string) (string, error) {
	h := sha1.New()

	if _, err := io.WriteString(h, salt); err != nil {
		return "", err
	}
	if _, err := io.WriteString(h, pass); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

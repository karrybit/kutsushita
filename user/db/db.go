package db

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"user/users"
)

type Database interface {
	Init() error

	GetUserByName(ctx context.Context, name string) (users.User, error)
	GetUser(ctx context.Context, id string) (users.User, error)
	GetUsers(ctx context.Context) ([]users.User, error)
	CreateUser(ctx context.Context, user *users.User) error

	GetUserAttributes(ctx context.Context, user *users.User) error

	GetCard(ctx context.Context, id string) (users.Card, error)
	GetCards(ctx context.Context) ([]users.Card, error)
	CreateCard(ctx context.Context, userCard *users.Card, userID string) error

	GetAddress(ctx context.Context, id string) (users.Address, error)
	GetAddresses(ctx context.Context) ([]users.Address, error)
	CreateAddress(ctx context.Context, userAddress *users.Address, userID string) error

	Delete(ctx context.Context, entity string, id string) error
	Ping(ctx context.Context) error
}

var (
	database              string
	DefaultDb             Database
	DBTypes               = map[string]Database{}
	ErrNoDatabaseFound    = "No database with name %v registerd"
	ErrNoDatabaseSelected = errors.New("No DB selected")
)

func init() {
	flag.StringVar(&database, "database", os.Getenv("USER_DATABASE"), "Database to use, Mongodb or ...")
}

func Init() error {
	if database == "" {
		return ErrNoDatabaseSelected
	}
	err := Set()
	if err != nil {
		return err
	}
	return DefaultDb.Init()
}

func Set() error {
	if v, ok := DBTypes[database]; ok {
		DefaultDb = v
		return nil
	}
	return fmt.Errorf(ErrNoDatabaseFound, database)
}

func Register(name string, db Database) {
	DBTypes[name] = db
}

func CreateUser(ctx context.Context, user *users.User) error {
	return DefaultDb.CreateUser(ctx, user)
}

func GetUserByName(ctx context.Context, name string) (users.User, error) {
	u, err := DefaultDb.GetUserByName(ctx, name)
	if err == nil {
		u.AddLinks()
	}
	return u, err
}

func GetUser(ctx context.Context, id string) (users.User, error) {
	u, err := DefaultDb.GetUser(ctx, id)
	if err == nil {
		u.AddLinks()
	}
	return u, err
}

func GetUsers(ctx context.Context) ([]users.User, error) {
	us, err := DefaultDb.GetUsers(ctx)
	for k := range us {
		us[k].AddLinks()
	}
	return us, err
}

func GetUserAttributes(ctx context.Context, user *users.User) error {
	err := DefaultDb.GetUserAttributes(ctx, user)
	if err != nil {
		return err
	}
	for k := range user.Addresses {
		user.Cards[k].AddLinks()
	}
	return nil
}

func CreateAddress(ctx context.Context, userAddress *users.Address, userID string) error {
	return DefaultDb.CreateAddress(ctx, userAddress, userID)
}

func GetAddress(ctx context.Context, id string) (users.Address, error) {
	a, err := DefaultDb.GetAddress(ctx, id)
	if err == nil {
		a.AddLinks()
	}
	return a, err
}

func GetAddresses(ctx context.Context) ([]users.Address, error) {
	as, err := DefaultDb.GetAddresses(ctx)
	for k := range as {
		as[k].AddLinks()
	}
	return as, err
}

func CreateCard(ctx context.Context, userCard *users.Card, userID string) error {
	return DefaultDb.CreateCard(ctx, userCard, userID)
}

func GetCard(ctx context.Context, id string) (users.Card, error) {
	return DefaultDb.GetCard(ctx, id)
}

func GetCards(ctx context.Context) ([]users.Card, error) {
	cs, err := DefaultDb.GetCards(ctx)
	for k := range cs {
		cs[k].AddLinks()
	}
	return cs, err
}

func Delete(ctx context.Context, entity string, id string) error {
	return DefaultDb.Delete(ctx, entity, id)
}

func Ping(ctx context.Context) error {
	return DefaultDb.Ping(ctx)
}

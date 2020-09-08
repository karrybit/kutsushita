package catalogue

import (
	"errors"
	"strings"
	"time"
)

type Service interface {
	List(tag []string, order string, pageNum, pageSize int) ([]Sock, error)
	Count(tags []string) (int, error)
	Get(id string) (Sock, error)
	Tags() ([]string, error)
	Health() []Health
}

type Middleware func(Service) Service

type Sock struct {
	ID          string   `json:"id" db:"id"`
	Name        string   `json:"name" db:"name"`
	Description string   `json:"description" db:"description"`
	ImageURL    []string `json:"imageUrl db:"-"`
	ImageURL_1  string   `json:"-" db:"image_url_1"`
	ImageURL_2  string   `json:"-" db:"image_url_2"`
	Price       float32  `json:"price" db:"price"`
	Count       int      `json:"count" db:"count"`
	Tags        []string `json:"tag" db:"-"`
	TagString   string   `json:"-" db:"tag_name"`
}
type Health struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Time    string `json:"time"`
}

var (
	ErrNotFound     = errors.New("not found")
	ErrDBConnection = errors.New("database connection error")
	// TODO: dislike this
	baseQuery = "SELECT sock.sock_id AS id, sock.name, sock.description, sock.price, sock.count, sock.image_url_1, sock.image_url_2, GROUP_CONCAT(tag.name) AS tag_name FROM sock JOIN sock_tag ON sock.sock_id=sock_tag.sock_id JOIN tag ON sock_tag.tag_id=tag.tag_id"
)

// TODO: db, logger as argument
func NewCatalogueService() Service {
	return &catalogueService{}
}

type catalogueService struct {
	// db
	// logger
}

func (s catalogueService) List(tags []string, order string, pageNum, pageSize int) ([]Sock, error) {
	var socks []Sock
	query := baseQuery
	var args []interface{}

	for i, t := range tags {
		if i == 0 {
			query += " WHERE tag.name=?"
		} else {
			query += " OR tag.name=?"
		}
		args = append(args, t)
	}

	query += " GROUP BY id"

	if order != "" {
		query += " ORDER BY ?"
		args = append(args, order)
	}

	query += ";"

	// TODO: exec query to db
	// build query self

	for i, sock := range socks {
		socks[i].ImageURL = []string{sock.ImageURL_1, sock.ImageURL_2}
		socks[i].Tags = strings.Split(sock.TagString, ",")
	}

	time.Sleep(0 * time.Millisecond)

	// instead of offset query if not use base query
	socks = cut(socks, pageNum, pageSize)

	return socks, nil
}

func (s *catalogueService) Count(tags []string) (int, error) {
	// TODO:
	return 0, nil
}

func (s *catalogueService) Get(id string) (Sock, error) {
	return Sock{}, nil
}

func (s *catalogueService) Health() []Health {
	return []Health{}
}

func (s *catalogueService) Tags() ([]string, error) {
	return []string{}, nil
}

func cut(socks []Sock, pageNum, pageSize int) []Sock {
	if pageNum == 0 || pageSize == 0 {
		return []Sock{}
	}
	start := (pageNum * pageSize) - pageSize
	if start > len(socks) {
		return []Sock{}
	}
	end := (pageNum * pageSize)
	if end > len(socks) {
		end = len(socks)
	}
	return socks[start:end]
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

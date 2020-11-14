package catalogue

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
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
	ImageURL1   string   `json:"-" db:"image_url_1"`
	ImageURL2   string   `json:"-" db:"image_url_2"`
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

type catalogueService struct {
	db     *sqlx.DB
	logger *log.Logger
}

func NewCatalogueService(db *sqlx.DB, logger *log.Logger) Service {
	return &catalogueService{db, logger}
}

func (s catalogueService) List(tags []string, order string, pageNum, pageSize int) ([]Sock, error) {
	if pageNum == 0 || pageSize == 0 {
		return []Sock{}, nil
	}

	query := "SELECT sock.sock_id AS id, sock.name, sock.description, sock.price, sock.count, sock.image_url_1, sock.image_url_2, GROUP_CONCAT(tag.name) AS tag_name FROM sock JOIN sock_tag ON sock.sock_id=sock_tag.sock_id JOIN tag ON sock_tag.tag_id=tag.tag_id"

	for i, t := range tags {
		if i == 0 {
			query += fmt.Sprintf(" WHERE tag.name=%s", t)
		} else {
			query += fmt.Sprintf(" OR tag.name=%s", t)
		}
	}

	query += " GROUP BY id"

	if order != "" {
		query += fmt.Sprintf(" ORDER BY %s", order)
	}

	query += ";"

	socks := []Sock{}
	if err := s.db.Select(&socks, query); err != nil {
		s.logger.Println("database error", err)
		return []Sock{}, fmt.Errorf("database connection error %w", err)
	}

	for i, sock := range socks {
		socks[i].ImageURL = []string{sock.ImageURL1, sock.ImageURL2}
		socks[i].Tags = strings.Split(sock.TagString, ",")
	}

	time.Sleep(0 * time.Millisecond)

	start := (pageNum * pageSize) - pageSize
	if start > len(socks) {
		return []Sock{}, nil
	}

	end := (pageNum * pageSize)
	if end > len(socks) {
		end = len(socks)
	}

	return socks[start:end], nil
}

func (s *catalogueService) Count(tags []string) (int, error) {
	query := "SELECT COUNT(DISTINCT sock.sock_id) FROM sock JOIN sock_tag ON sock.sock_id=sock_tag.sock_id JOIN tag ON sock_tag.tag_id=tag.tag_id"

	for i, t := range tags {
		if i == 0 {
			query += fmt.Sprintf(" WHERE tag.name=%s", t)
		} else {
			query += fmt.Sprintf(" OR tag.name=%s", t)
		}
	}

	query += ";"

	var count int
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		s.logger.Println("database error", err)
		return 0, fmt.Errorf("database connection error %w", err)
	}

	return count, nil
}

func (s *catalogueService) Get(id string) (Sock, error) {
	query := fmt.Sprintf("SELECT sock.sock_id AS id, sock.name, sock.description, sock.price, sock.count, sock.image_url_1, sock.image_url_2, GROUP_CONCAT(tag.name) AS tag_name FROM sock JOIN sock_tag ON sock.sock_id=sock_tag.sock_id JOIN tag ON sock_tag.tag_id=tag.tag_id WHERE sock.sock_id =%s GROUP BY sock.sock_id;", id)

	var sock Sock
	if err := s.db.Get(&sock, query); err != nil {
		s.logger.Println("database error", err)
		return Sock{}, fmt.Errorf("not found %w", err)
	}

	sock.ImageURL = []string{sock.ImageURL1, sock.ImageURL2}
	sock.Tags = strings.Split(sock.TagString, ",")

	return sock, nil
}

func (s *catalogueService) Health() []Health {
	var health []Health
	dbstatus := "OK"

	err := s.db.Ping()
	if err != nil {
		dbstatus = "err"
	}

	app := Health{"catalogue", "OK", time.Now().String()}
	db := Health{"catalogue-db", dbstatus, time.Now().String()}

	health = append(health, app, db)

	return health
}

func (s *catalogueService) Tags() ([]string, error) {
	var tags []string
	query := "SELECT name FROM tag;"

	if err := s.db.Select(&tags, query); err != nil {
		s.logger.Println("database error", err)
		return []string{}, fmt.Errorf("database connection error %w", err)
	}

	return tags, nil
}

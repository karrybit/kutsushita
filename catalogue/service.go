package catalogue

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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

var (
	ErrNotFound     = errors.New("not found")
	ErrDBConnection = errors.New("database connection error")
	// TODO dislike this
	baseQuery = "SELECT sock.sock_id AS id, sock.name, sock.description, sock.price, sock.count, sock.image_url_1, sock.image_url_2, GROUP_CONCAT(tag.name) AS tag_name FROM sock JOIN sock_tag ON sock.sock_id=sock_tag.sock_id JOIN tag ON sock_tag.tag_id=tag.tag_id"
)

type catalogueService struct {
	db     *sql.DB
	logger *log.Logger
}

// TODO db, logger as argument
func NewCatalogueService(db *sql.DB, logger *log.Logger) Service {
	return &catalogueService{db, logger}
}

func (s catalogueService) List(tags []string, order string, pageNum, pageSize int) ([]Sock, error) {
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

	rows, err := s.db.Query(query)
	if err != nil {
		// TODO log
		// TODO wrap error
		return []Sock{}, ErrDBConnection
	}

	socks := []Sock{}
	for rows.Next() {
		sock := Sock{}
		if err = rows.Scan(&sock); err != nil {
			// TODO log
			panic(err)
		}
		socks = append(socks, sock)
	}

	for i, sock := range socks {
		socks[i].ImageURL = []string{sock.ImageURL1, sock.ImageURL2}
		socks[i].Tags = strings.Split(sock.TagString, ",")
	}

	time.Sleep(0 * time.Millisecond)

	// instead of offset query if not use base query
	socks = cut(socks, pageNum, pageSize)

	return socks, nil
}

func (s *catalogueService) Count(tags []string) (int, error) {
	query := "SELECT COUNT(DISTINCT sock.sock_id) FROM sock JOIN sock_tag ON sock.sock_id=sock_tag.sock_id JOIN tag ON sock_tag.tag_id=tag.tag_id"

	var args []interface{}

	for i, t := range tags {
		if i == 0 {
			query += " WHERE tag.name=?"
		} else {
			query += " OR tag.name=?"
		}
		args = append(args, t)
	}

	query += ";"

	sel, err := s.db.Prepare(query)
	if err != nil {
		// TODO log
		return 0, ErrDBConnection
	}
	defer sel.Close()

	var count int
	err = sel.QueryRow(args...).Scan(&count)
	if err != nil {
		// TODO log
		// TODO wrap error
		return 0, ErrDBConnection
	}

	return count, nil
}

func (s *catalogueService) Get(id string) (Sock, error) {
	query := baseQuery + fmt.Sprintf(" WHERE sock.sock_id =%s GROUP BY sock.sock_id;", id)

	var sock Sock
	err := s.db.QueryRow(query).Scan(&sock)
	if err != nil {
		// TODO log
		// TODO wrap error
		return Sock{}, ErrNotFound
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

	health = append(health, app)
	health = append(health, db)

	return health
}

func (s *catalogueService) Tags() ([]string, error) {
	var tags []string
	query := "SELECT name FROM tag;"

	rows, err := s.db.Query(query)
	if err != nil {
		// TODO; log
		// TODO wrap error
		return []string{}, ErrDBConnection
	}

	var tag string
	for rows.Next() {
		err = rows.Scan(&tag)
		if err != nil {
			//TODO log
			continue
		}
		tags = append(tags, tag)
	}

	return tags, nil
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

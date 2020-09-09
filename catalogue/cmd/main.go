package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"catalogue"

	_ "github.com/go-sql-driver/mysql"
)

const ServiceName = "catalogue"

func main() {
	// ctx := context.Background()

	// TODO opentracing

	// TODO db
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		// TODO log
		os.Exit(1)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		// TODO log
	}

	_ = catalogue.NewCatalogueService(db)

	// TODO launch server

	errc := make(chan error)

	go func() {
		errc <- http.ListenAndServe(":80", nil)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	fmt.Println("exit", <-errc)
}

package main

import (
	"database/sql"
	"fmt"
	"log"
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

	var logger log.Logger

	// TODO opentracing

	// TODO db
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		logger.Println("err", err)
		os.Exit(1)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.Println("Error", "Unable to connect to Database", "DSN")
	}

	service := catalogue.NewCatalogueService(db, &logger)
	service = catalogue.LoggingMiddleware(&logger)(service)

	// TODO launch server

	errc := make(chan error)

	go func() {
		logger.Println("transport", "HTTP", "port")
		errc <- http.ListenAndServe(":80", nil)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	fmt.Println("exit", <-errc)
}

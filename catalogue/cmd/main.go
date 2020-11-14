package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"catalogue"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const ServiceName = "catalogue"

func main() {
	var (
		port   = flag.String("port", "80", "Port to bind HTTP listener") // TODO(pb): should be -addr, default ":80"
		images = flag.String("images", "./images/", "Image path")
		_      = flag.String("DSN", "catalogue_user:default_password@tcp(catalogue-db:3306)/socksdb", "Data Source Name: [username[:password]@][protocol[(address)]]/dbname")
		_      = flag.String("zipkin", os.Getenv("ZIPKIN"), "Zipkin address")
	)
	flag.Parse()

	fmt.Fprintf(os.Stderr, "images: %q\n", *images)
	abs, err := filepath.Abs(*images)
	fmt.Fprintf(os.Stderr, "Abs(images): %q (%v)\n", abs, err)
	pwd, err := os.Getwd()
	fmt.Fprintf(os.Stderr, "Getwd: %q (%v)\n", pwd, err)
	files, _ := filepath.Glob(*images + "/*")
	fmt.Fprintf(os.Stderr, "ls: %q\n", files)

	// ctx := context.Background()

	logger := zap.L()

	// TODO opentelemetry

	db, err := sqlx.Connect("mysql", "user:password@/dbname")
	if err != nil {
		logger.Fatal("Error", zap.Error(err))
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		logger.Error("Error", zap.Error(err))
	}

	service := catalogue.NewCatalogueService(db, logger)
	service = catalogue.LoggingMiddleware(logger)(service)

	app := catalogue.MakeHTTPHandler(service, *images)

	errc := make(chan error)

	go func() {
		logger.Info("transport HTTP port")
		errc <- app.Listen(":" + *port)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	logger.Info("exit", zap.Error(<-errc))
}

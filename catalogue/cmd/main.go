package main

import (
	"catalogue"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const ServiceName = "catalogue"

func main() {
	// ctx := context.Background()

	// TODO: opentracing

	// TODO: db

	// TODO: service
	_ = catalogue.NewCatalogueService()

	// TODO: launch server

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

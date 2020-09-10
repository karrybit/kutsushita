package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"payment"
	"syscall"
)

const ServiceName = "payment"

func main() {
	var (
		port          = flag.String("port", "8080", "Port to bind HTTP listener")
		_             = flag.String("zipkin", os.Getenv("ZIPKIN"), "Zipkin address")
		declineAmount = flag.Float64("decline", 105, "Decline payments over certain amount")
	)
	flag.Parse()

	// TODO tracer

	logger := log.Logger{}
	service := payment.NewAuthorisationService(float32(*declineAmount))
	service = payment.LoggingMiddleware(&logger)(service)

	router := payment.MakeHTTPHandler(service)

	errc := make(chan error)
	go func() {
		logger.Println("transport", "HTTP", "port", *port)
		errc <- http.ListenAndServe(":"+*port, router)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	logger.Println("exit", <-errc)
}

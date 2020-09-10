package main

import (
	"context"
	"flag"
	"fmt"
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

	ctx := context.Background()
	handler, logger := payment.WireUp(ctx, float32(*declineAmount), ServiceName)

	errc := make(chan error)
	go func() {
		logger.Println("transport", "HTTP", "port", *port)
		errc <- http.ListenAndServe(":"+*port, handler)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	logger.Println("exit", <-errc)
}

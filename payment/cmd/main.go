package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"payment"
	"syscall"

	"go.uber.org/zap"
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

	logger := zap.L()
	service := payment.NewAuthorisationService(float32(*declineAmount))
	service = payment.LoggingMiddleware(logger)(service)

	router := payment.MakeHTTPHandler(service)

	errc := make(chan error)
	go func() {
		logger.Info("transport HTTP", zap.String("port", *port))
		errc <- router.Listen(":" + *port)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	logger.Info("exit", zap.Error(<-errc))
}

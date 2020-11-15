package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"user/db"
	"user/db/mongodb"

	"go.uber.org/zap"
)

var (
	zip  string
	port string
)

const (
	ServiceName = "user"
)

func main() {
	flag.StringVar(&zip, "zipkin", os.Getenv("ZIPKIN"), "Zipkin address")
	flag.StringVar(&port, "port", "8084", "Port on which to run")
	db.Register("mongodb", &mongodb.Mongo{})

	flag.Parse()

	logger := zap.L()

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logger.Fatal("", zap.Error(err))
	}
	// localAddr := conn.LocalAddr().(*net.UDPAddr)
	// host := strings.Split(localAddr.String(), ":")[0]
	defer conn.Close()

	// TODO: tracer

	dbconn := false
	for !dbconn {
		err := db.Init()
		if err != nil {
			if err == db.ErrNoDatabaseSelected {
				logger.Fatal("", zap.Error(err))
			}
			logger.Error("", zap.Error(err))
		} else {
			dbconn = true
		}
	}

	// fieldKeys := []string{"method"}

	// service := api.NewFixedService()
	// service = api.LoggingMiddleware(logger)(service)

	// endpoints := api.MakeEndpoints(service, tracer)

	// router := api.MakeHTTPHandler(endpoints, logger, tracer)

	// TODO: httpMiddleware
	// TODO: handler

	errc := make(chan error)
	go func() {
		logger.Info("transport HTTP", zap.String("port", port))
		// errc <- http.ListenAndServe(fmt.Sprintf(":%v", port), handler)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	logger.Info("exit", zap.Error(<-errc))
}

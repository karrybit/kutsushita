package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"user/api"
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

	service := api.NewFixedService()
	service = api.LoggingMiddleware(logger)(service)
	router := api.MakeHTTPHandler(service, logger)

	// TODO: httpMiddleware
	// TODO: handler

	errc := make(chan error)
	go func() {
		logger.Info("transport HTTP", zap.String("port", port))
		errc <- router.Listen(":" + port)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	logger.Info("exit", zap.Error(<-errc))
}

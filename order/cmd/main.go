package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"order"

	"go.uber.org/zap"
)

var (
	zip  string
	port string
)

const (
	ServiceName = "order"
)

func main() {
	flag.StringVar(&zip, "zipkin", os.Getenv("ZIPKIN"), "Zipkin address")
	flag.StringVar(&port, "port", "8084", "Port on which to run")

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
		err := order.InitDB()
		if err != nil {
			logger.Error("", zap.Error(err))
		} else {
			dbconn = true
		}
	}

	service := order.NewService(logger)
	router := order.MakeHTTPHandler(service)

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

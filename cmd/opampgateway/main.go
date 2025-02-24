package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/opampgateway/gateway"
	"go.uber.org/zap"
)

func main() {
	configFlag := flag.String("config", "", "Path to a gateway configuration file")
	flag.Parse()

	cfg := &gateway.Config{}

	cfg, err := cfg.Load(*configFlag)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	gateway := gateway.NewGateway(logger, cfg)
	if err := gateway.Start(context.Background()); err != nil {
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	if err := gateway.Stop(context.Background()); err != nil {
		log.Fatal(err)
	}
}

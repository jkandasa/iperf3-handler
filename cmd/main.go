package main

import (
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-cmd/cmd"
	"github.com/jkandasa/iperf3-handler/pkg/handler"
	"github.com/jkandasa/iperf3-handler/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	httpAddress = ":8080"
)

func initLogger() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func main() {
	initLogger()

	go shutdownHook()      // start shutdown hook
	go startIPerf3Server() // start the iperf3 server

	// start the http handler
	zap.L().Info("starting http lister", zap.String("address", httpAddress))
	err := http.ListenAndServe(httpAddress, handler.NewHandler())
	if err != nil {
		zap.L().Fatal("error on starting listener", zap.Error(err))
	}
}

func startIPerf3Server() {
	options := strings.Split(types.IPerf3ServerCommand, " ")
	serverCmd := cmd.NewCmd(options[0], options[1:]...)
	zap.L().Info("starting iperf3 server", zap.Strings("commands", options))
	statusChan := serverCmd.Start()
	<-statusChan

	status := serverCmd.Status()
	if status.Error != nil {
		zap.L().Fatal("error on starting iperf3 server", zap.Error(status.Error), zap.Strings("stdout", status.Stdout), zap.Strings("stderr", status.Stderr))
	}
}

func shutdownHook() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// waiting for signal
	sig := <-sigs
	close(sigs)

	zap.L().Info("shutdown initiated..", zap.Any("signal", sig))
	os.Exit(0)
}

package main

import (
	"go.uber.org/zap"
	"time"
	"weaveTest/internal/config"
	"weaveTest/internal/server"
	"weaveTest/internal/server/github/client"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// since we need a logger instance
	client.InitRateLimiter(10, time.Minute, logger)

	err = config.InitEnvFromFile("./.env.json")
	if err != nil {
		logger.Error("failed to load env file", zap.Error(err))
		panic(err)
	}

	// TODO: add command line args for port and githubToken
	serverInstance := server.NewServerWithDefaultPort(logger)
	serverInstance.Start()
}

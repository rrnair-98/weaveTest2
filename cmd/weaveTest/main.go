package main

import (
	"go.uber.org/zap"
	"weaveTest/internal/config"
	server "weaveTest/internal/server"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	err = config.InitEnvFromFile("./.env.json")
	if err != nil {
		logger.Error("failed to load env file", zap.Error(err))
		panic(err)
	}

	// TODO: add command line args for port and githubToken
	serverInstance := server.NewServerWithDefaultPort(logger)
	serverInstance.Start()
}

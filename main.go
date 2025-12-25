package main

import (
	"context"
	"edgelink-api/internal/infrastructure/config"
	"edgelink-api/internal/pkg/logger"
	"edgelink-api/internal/startup"
	"flag"
	"log/slog"
	"os/signal"
	"syscall"
	"time"
)

type appConfig struct {
	configPath string
}

func parseFlag() appConfig {
	var configPath string
	flag.StringVar(&configPath, "c", "./config/config.yaml", "set yaml file path")
	flag.Parse()
	return appConfig{
		configPath: configPath,
	}
}

func main() {
	appCfg := parseFlag()

	if err := logger.InitLogger(logger.LogConfig{
		Level:         slog.LevelInfo,
		LogFmt:        logger.LogTextFormat,
		FilePath:      "./run.log",
		ShowLogSource: true,
	}); err != nil {
		panic(err)
	}

	config.InitConfigWithViper(appCfg.configPath)
	container, srv := startup.InitApp()
	startup.RunWebServer(srv)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()
	startup.Stop(context.Background(), srv, container, 10*time.Second)
}

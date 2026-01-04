package main

import (
	"context"
	"edgelink-api/internal/bootstrap"
	"edgelink-api/internal/infrastructure/config"
	"edgelink-api/internal/pkg/logger"
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

	bizCtx, cancel := context.WithCancel(context.Background())

	config.InitConfigWithViper(appCfg.configPath)
	container, srv := bootstrap.InitApp()
	bootstrap.RunWebServer(srv)
	bootstrap.InitDataLoader(bizCtx, container.RedisCmd)

	// wait quit signal
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()
	bootstrap.Stop(bizCtx, srv, container, 10*time.Second)
	cancel() // todo 这里还可以优化一下退出逻辑
}

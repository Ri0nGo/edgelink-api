package main

import (
	"context"
	"edgelink-api/internal/bootstrap"
	"edgelink-api/internal/infrastructure/config"
	"edgelink-api/internal/pkg/logger"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
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

	config.InitConfigWithViper(appCfg.configPath)

	if err := logger.InitLogger(logger.LogConfig{
		Level:         viper.GetString("logger.level"),
		LogFmt:        logger.LogFormat(viper.GetString("logger.format")),
		FilePath:      viper.GetString("logger.filepath"),
		ShowLogSource: viper.GetBool("logger.show_source"),
	}); err != nil {
		panic(err)
	}

	bizCtx, cancel := context.WithCancel(context.Background())

	container, srv := bootstrap.InitApp(viper.GetString("app.addr"))
	bootstrap.Bootstrap(bizCtx, container, srv)

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

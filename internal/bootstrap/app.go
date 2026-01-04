package bootstrap

import (
	"context"
	"edgelink-api/internal/ioc"
	"edgelink-api/internal/pkg/logger"
	"errors"
	"net/http"
	"time"
)

func InitApp(addr ...string) (*ioc.Container, *http.Server) {
	container := ioc.InitWebServer()
	address := ":8080"
	if len(addr) > 0 {
		address = addr[0]
	}
	srv := &http.Server{
		Addr:    address,
		Handler: container.Engine,
	}
	return container, srv
}

func RunWebServer(srv *http.Server) {
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
}

func Stop(ctx context.Context, srv *http.Server, ctr *ioc.Container, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	logger.Info("shutting down http server...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error:", err)
	}

	ctr.Close()
	logger.Info("shutdown completed")
}

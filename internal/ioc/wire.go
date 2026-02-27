//go:build wireinject
// +build wireinject

package ioc

import (
	"edgelink-api/internal/api"
	"edgelink-api/internal/infrastructure/cache"
	"edgelink-api/internal/infrastructure/db"
	"edgelink-api/internal/pkg/logger"
	"edgelink-api/internal/repo"
	"edgelink-api/internal/router"
	"edgelink-api/internal/svc"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	baseSet = wire.NewSet(
		db.InitDB,
		cache.InitRedisCache,
	)
)

type Container struct {
	Engine   *gin.Engine
	RedisCmd redis.Cmdable
	MysqlDB  *gorm.DB
}

func (c *Container) Close() {
	if c.RedisCmd != nil {
		if r, ok := c.RedisCmd.(io.Closer); ok {
			_ = r.Close()
			logger.Info("close redis completed")
		}
	}

	if c.MysqlDB != nil {
		db, err := c.MysqlDB.DB()
		if err == nil {
			_ = db.Close()
			logger.Info("close mysql completed")
		}
	}
}

func InitWebServer() *Container {
	wire.Build(
		baseSet,

		// repo layer
		repo.NewThingModelRepo,
		repo.NewThingModelPropRepo,
		repo.NewProductRepo,
		repo.NewDeviceRepo,
		repo.NewHistoryDataRepo,

		// svc layer
		svc.NewThingModelSvc,
		svc.NewProductSvc,
		svc.NewDeviceSvc,
		svc.NewMetricSvc,

		// api layer
		api.NewThingModelApi,
		api.NewProductApi,
		api.NewDeviceApi,
		api.NewMetricApi,

		router.LoadRegistryRouters,
		router.InitRouter,

		wire.Struct(new(Container), "*"),
	)
	return nil
}

package router

import (
	"edgelink-api/internal/api"

	"github.com/gin-gonic/gin"
)

type RegistryRouter interface {
	RegistryRouter(g *gin.RouterGroup)
}

func InitRouter(routers []RegistryRouter) *gin.Engine {
	engine := gin.Default()
	//engine.Use(mdls...)

	group := engine.Group("/api/edgelink")
	for _, router := range routers {
		router.RegistryRouter(group)
	}

	return engine
}

func LoadRegistryRouters(
	tmApi *api.ThingModelApi,
	productApi *api.ProductApi,
	deviceApi *api.DeviceApi,
) []RegistryRouter {
	return []RegistryRouter{
		tmApi,
		productApi,
		deviceApi,
	}
}

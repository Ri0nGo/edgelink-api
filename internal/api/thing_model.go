package api

import (
	"edgelink-api/internal/svc"

	"github.com/gin-gonic/gin"
)

type ThingModelApi struct {
	tmSvc svc.IThingModelSvc
}

func (a *ThingModelApi) RegistryRouter(g *gin.RouterGroup) {
	group := g.Group("/thing_model")
	group.GET("/list", a.GetThingModelList)
}

func (a *ThingModelApi) GetThingModelList(ctx *gin.Context) {
	ctx.JSON(200, "ok")
}

func NewThingModelApi(tmSvc svc.IThingModelSvc) *ThingModelApi {
	return &ThingModelApi{tmSvc: tmSvc}
}

package api

import (
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/pkg/ginx/response"
	"edgelink-api/internal/pkg/logger"
	"edgelink-api/internal/svc"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ThingModelApi struct {
	tmSvc svc.IThingModelSvc
}

func (a *ThingModelApi) RegistryRouter(g *gin.RouterGroup) {
	group := g.Group("/thing_model")
	group.POST("/create", a.CreateThingModel)
	group.POST("/update", a.UpdateThingModel)
	group.POST("/delete", a.DeleteThingModel)
	group.GET("/:id", a.GetThingModelDetail)
	group.GET("/list", a.GetThingModelList)
}

func (a *ThingModelApi) CreateThingModel(ctx *gin.Context) {
	var req dto.ReqThingModel
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("add thing model err", "err", err)
		response.Failed(ctx, response.RespCodeParamErr)
		return
	}

	if err := a.tmSvc.CreateThingModel(ctx.Request.Context(), &req); err != nil {
		logger.Error("create thing model err", "err", err)
		response.Failed(ctx, response.RespCodeInternalErr)
		return
	}
	response.Success(ctx)
}

func (a *ThingModelApi) UpdateThingModel(ctx *gin.Context) {
	var req dto.ReqThingModel
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("update thing model err", "err", err)
		response.Failed(ctx, response.RespCodeParamErr)
		return
	}

	if err := a.tmSvc.UpdateThingModel(ctx.Request.Context(), &req); err != nil {
		logger.Error("update thing model err", "err", err)
		response.Failed(ctx, response.RespCodeInternalErr)
		return
	}
	response.Success(ctx)
}

func (a *ThingModelApi) DeleteThingModel(ctx *gin.Context) {
	var req dto.ReqId
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("delete thing model err", "err", err)
		response.Failed(ctx, response.RespCodeParamErr)
		return
	}

	if err := a.tmSvc.DeleteThingModel(ctx.Request.Context(), req.Id); err != nil {
		logger.Error("delete thing model err", "err", err)
		if errors.Is(err, gorm.ErrRecordNotFound) { // 触发次数多了说明不对劲，可以做监控
			response.Success(ctx)
			return
		}
		response.Failed(ctx, response.RespCodeInternalErr)
		return
	}
	response.Success(ctx)
}

func (a *ThingModelApi) GetThingModelDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Failed(ctx, response.RespCodeParamErr)
		return
	}

	tmDao, err := a.tmSvc.GetThingModelById(ctx.Request.Context(), id)
	if err != nil {
		logger.Error("get thing model err", "err", err)
		response.Failed(ctx, response.RespCodeInternalErr)
		return
	}
	response.Success(ctx, tmDao)
}

func (a *ThingModelApi) GetThingModelList(ctx *gin.Context) {
	var req dto.RespThingModelList
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.Failed(ctx, response.RespCodeParamErr)
		return
	}

	page, err := a.tmSvc.GetThingModelList(ctx.Request.Context(), req.Search, req.Page)
	if err != nil {
		response.Failed(ctx, response.RespCodeInternalErr)
		return
	}
	response.Success(ctx, page)
}

func NewThingModelApi(tmSvc svc.IThingModelSvc) *ThingModelApi {
	return &ThingModelApi{tmSvc: tmSvc}
}

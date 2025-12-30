package api

import (
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/pkg/ginx/handler"
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

	// thing model property
	// todo 待实现
	propGroup := group.Group("/prop")
	propGroup.POST("/create") // 新增属性后，需要手动同步属性到设备
	propGroup.POST("/update") // 更新属性，注意同步设备监听
	propGroup.POST("/delete") // 检查是否有产品在使用，如果有在使用的话，则不允许删除
	propGroup.POST("/list")
}

func (a *ThingModelApi) CreateThingModel(ctx *gin.Context) {
	var req dto.ReqThingModel
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("add thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.tmSvc.CreateThingModel(ctx.Request.Context(), &req); err != nil {
		logger.Error("create thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *ThingModelApi) UpdateThingModel(ctx *gin.Context) {
	var req dto.ReqThingModel
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("update thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.tmSvc.UpdateThingModel(ctx.Request.Context(), &req); err != nil {
		logger.Error("update thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *ThingModelApi) DeleteThingModel(ctx *gin.Context) {
	var req dto.ReqId
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("delete thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.tmSvc.DeleteThingModel(ctx.Request.Context(), req.Id); err != nil {
		logger.Error("delete thing model err", "err", err)
		if errors.Is(err, gorm.ErrRecordNotFound) { // 触发次数多了说明不对劲，可以做监控
			handler.Success(ctx)
			return
		}
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *ThingModelApi) GetThingModelDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	tmDao, err := a.tmSvc.GetThingModelById(ctx.Request.Context(), id)
	if err != nil {
		logger.Error("get thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	response.Success(ctx, tmDao)
}

func (a *ThingModelApi) GetThingModelList(ctx *gin.Context) {
	var req dto.RespThingModelList
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	page, err := a.tmSvc.GetThingModelList(ctx.Request.Context(), req.Search, req.Page)
	if err != nil {
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	response.Success(ctx, page)
}

func NewThingModelApi(tmSvc svc.IThingModelSvc) *ThingModelApi {
	return &ThingModelApi{tmSvc: tmSvc}
}

package api

import (
	"edgelink-api/internal/api/dto"
	bizErr "edgelink-api/internal/pkg/bizerr"
	"edgelink-api/internal/pkg/ginx/handler"
	"edgelink-api/internal/pkg/ginx/response"
	"edgelink-api/internal/pkg/logger"
	"edgelink-api/internal/svc"
	"errors"
	"fmt"
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
	propGroup := group.Group("/prop")
	propGroup.POST("/create", a.CreateThingModelProp) // 新增属性后，需要手动同步属性到设备
	propGroup.POST("/update", a.UpdateThingModelProp) // 更新属性，注意同步设备监听
	propGroup.POST("/delete", a.DeleteThingModelProp) // 检查是否有产品在使用，如果有在使用的话，则不允许删除
	propGroup.GET("/list", a.GetThingModelPropList)
}

func (a *ThingModelApi) CreateThingModel(ctx *gin.Context) {
	var req dto.ReqThingModel
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("add thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	var keyMap = make(map[string]struct{})
	for _, funcType := range req.FuncTypes {
		_, ok := keyMap[funcType.Key]
		if ok {
			handler.HandlerError(ctx, response.RespCodeParamErr, bizErr.NewBizError(fmt.Sprintf("标识符: %s 重复", funcType.Key)))
		}
		keyMap[funcType.Key] = struct{}{}
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

	c := ctx.Request.Context()
	tmDao, err := a.tmSvc.GetThingModelById(c, id)
	if err != nil {
		logger.Error("get thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	props, err := a.tmSvc.GetThingModelProps(c, id)
	if err != nil {
		logger.Error("get thing model prop list err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	tmDao.Props = props
	handler.Success(ctx, tmDao)
}

func (a *ThingModelApi) GetThingModelList(ctx *gin.Context) {
	var req dto.ReqPageSearch
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	page, err := a.tmSvc.GetThingModelList(ctx.Request.Context(), req.Search, req.Page)
	if err != nil {
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx, page)
}

// ---------------- 物模型属性 ---------------- //

func (a *ThingModelApi) GetThingModelPropList(ctx *gin.Context) {
	var req dto.ReqPageSearch
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	page, err := a.tmSvc.GetThingModelPropList(ctx, req.ModelId, req.Search, req.Page)
	if err != nil {
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx, page)
}

func (a *ThingModelApi) CreateThingModelProp(ctx *gin.Context) {
	var req dto.ReqThingModelProp
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("add thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.tmSvc.CreateThingModelProp(ctx.Request.Context(), &req); err != nil {
		logger.Error("create thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *ThingModelApi) UpdateThingModelProp(ctx *gin.Context) {
	var req dto.ReqThingModelProp
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("update thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.tmSvc.UpdateThingModelProp(ctx.Request.Context(), &req); err != nil {
		logger.Error("update thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *ThingModelApi) DeleteThingModelProp(ctx *gin.Context) {
	var req dto.ReqId
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("delete thing model err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.tmSvc.DeleteThingModelProp(ctx.Request.Context(), req.Id); err != nil {
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

func NewThingModelApi(tmSvc svc.IThingModelSvc) *ThingModelApi {
	return &ThingModelApi{tmSvc: tmSvc}
}

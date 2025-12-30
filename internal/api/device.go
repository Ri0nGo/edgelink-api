package api

import (
	"edgelink-api/internal/api/dto"
	bizErr "edgelink-api/internal/pkg/bizerr"
	"edgelink-api/internal/pkg/ginx/handler"
	"edgelink-api/internal/pkg/ginx/response"
	"edgelink-api/internal/pkg/logger"
	"edgelink-api/internal/svc"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeviceApi struct {
	deviceSvc svc.IDeviceSvc
}

func (a *DeviceApi) RegistryRouter(g *gin.RouterGroup) {
	group := g.Group("/device")
	group.POST("/create", a.CreateDevice)
	group.POST("/update", a.UpdateDevice)
	group.POST("/delete", a.DeleteDevice)
	group.GET("/:id", a.GetDeviceDetail)
	group.GET("/list", a.GetDeviceList)

	// 设备属性
	propGroup := group.Group("/prop")
	propGroup.POST("/update", a.UpdateDeviceProp)
	propGroup.POST("/delete", a.DeleteDeviceProp)
}

func (a *DeviceApi) CreateDevice(ctx *gin.Context) {
	var req dto.ReqDevice
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("add Device err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if req.Key == "" {
		handler.HandlerError(ctx, response.RespCodeParamErr, bizErr.NewBizError("设备标识符不能为空"))
		return
	}

	if err := a.deviceSvc.CreateDevice(ctx.Request.Context(), &req); err != nil {
		logger.Error("create Device err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *DeviceApi) UpdateDevice(ctx *gin.Context) {
	var req dto.ReqDevice
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("update Device err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if req.Key != "" {
		handler.HandlerError(ctx, response.RespCodeParamErr, bizErr.NewBizError("设备标识符不能修改"))
		return
	}

	if err := a.deviceSvc.UpdateDevice(ctx.Request.Context(), &req); err != nil {
		logger.Error("update Device err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *DeviceApi) DeleteDevice(ctx *gin.Context) {
	var req dto.ReqId
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("delete Device err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.deviceSvc.DeleteDevice(ctx.Request.Context(), req.Id); err != nil {
		logger.Error("delete Device err", "err", err)
		if errors.Is(err, gorm.ErrRecordNotFound) { // 触发次数多了说明不对劲，可以做监控
			handler.Success(ctx)
			return
		}
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *DeviceApi) GetDeviceDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	deviceDao, err := a.deviceSvc.GetDeviceById(ctx.Request.Context(), id)
	if err != nil {
		logger.Error("get Device err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx, deviceDao)
}

func (a *DeviceApi) GetDeviceList(ctx *gin.Context) {
	var page dto.Page
	if err := ctx.ShouldBindQuery(&page); err != nil {
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	page, err := a.deviceSvc.GetDeviceList(ctx.Request.Context(), page)
	if err != nil {
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx, page)
}

// ---------------- device property ---------------- //

func (a *DeviceApi) UpdateDeviceProp(ctx *gin.Context) {
	var req dto.ReqDeviceProp
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("update Device err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.deviceSvc.UpdateDeviceProp(ctx.Request.Context(), &req); err != nil {
		logger.Error("update Device err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *DeviceApi) DeleteDeviceProp(ctx *gin.Context) {
	var req dto.ReqIds
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("delete Device err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.deviceSvc.DeleteDeviceProps(ctx.Request.Context(), &req); err != nil {
		logger.Error("delete Device err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func NewDeviceApi(deviceSvc svc.IDeviceSvc) *DeviceApi {
	return &DeviceApi{deviceSvc: deviceSvc}
}

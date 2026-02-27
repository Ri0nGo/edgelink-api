package api

import (
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/pkg/ginx/handler"
	"edgelink-api/internal/pkg/ginx/response"
	"edgelink-api/internal/pkg/logger"
	"edgelink-api/internal/svc"

	"github.com/gin-gonic/gin"
)

type MetricApi struct {
	metricSvc svc.IMetricSvc
}

func NewMetricApi(metricSvc svc.IMetricSvc) *MetricApi {
	return &MetricApi{metricSvc: metricSvc}
}

func (a *MetricApi) RegistryRouter(g *gin.RouterGroup) {
	dataGroup := g.Group("/data")
	{
		dataGroup.POST("/timeseries", a.GetTimeseriesData) // 获取时序数据
	}

}

func (a *MetricApi) GetTimeseriesData(ctx *gin.Context) {
	var req dto.ReqTimeSeriesData
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("get time series data err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	data, err := a.metricSvc.GetTimeSeriesHistoryData(ctx.Request.Context(),
		req.DeviceIds, req.PropertyIds, req.Begin, req.End)
	if err != nil {
		logger.Error("get time series data err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx, data)
}

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

type ProductApi struct {
	tmSvc svc.IProductSvc
}

func (a *ProductApi) RegistryRouter(g *gin.RouterGroup) {
	group := g.Group("/product")
	group.POST("/create", a.CreateProduct)
	group.POST("/update", a.UpdateProduct)
	group.POST("/delete", a.DeleteProduct)
	group.GET("/:id", a.GetProductDetail)
	group.GET("/list", a.GetProductList)
}

func (a *ProductApi) CreateProduct(ctx *gin.Context) {
	var req dto.ReqProduct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("add product err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.tmSvc.CreateProduct(ctx.Request.Context(), &req); err != nil {
		logger.Error("create product err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *ProductApi) UpdateProduct(ctx *gin.Context) {
	var req dto.ReqProduct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("update product err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.tmSvc.UpdateProduct(ctx.Request.Context(), &req); err != nil {
		logger.Error("update product err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *ProductApi) DeleteProduct(ctx *gin.Context) {
	var req dto.ReqId
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("delete product err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	if err := a.tmSvc.DeleteProduct(ctx.Request.Context(), req.Id); err != nil {
		logger.Error("delete product err", "err", err)
		if errors.Is(err, gorm.ErrRecordNotFound) { // 触发次数多了说明不对劲，可以做监控
			handler.Success(ctx)
			return
		}
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx)
}

func (a *ProductApi) GetProductDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	tmDao, err := a.tmSvc.GetProductById(ctx.Request.Context(), id)
	if err != nil {
		logger.Error("get product err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx, tmDao)
}

func (a *ProductApi) GetProductList(ctx *gin.Context) {
	var req dto.ReqPageSearch
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	page, err := a.tmSvc.GetProductList(ctx.Request.Context(), req.Search, req.Page)
	if err != nil {
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx, page)
}

func NewProductApi(tmSvc svc.IProductSvc) *ProductApi {
	return &ProductApi{tmSvc: tmSvc}
}

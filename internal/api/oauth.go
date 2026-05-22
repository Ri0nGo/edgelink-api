package api

import (
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/pkg/ginx/handler"
	"edgelink-api/internal/pkg/ginx/response"
	"edgelink-api/internal/pkg/logger"
	"edgelink-api/internal/svc"

	"github.com/gin-gonic/gin"
)

type OAuthApi struct {
	oauthSvc svc.IOAuthSvc
}

func NewOAuthApi(oauthSvc svc.IOAuthSvc) *OAuthApi {
	return &OAuthApi{oauthSvc: oauthSvc}
}

func (a *OAuthApi) RegistryRouter(g *gin.RouterGroup) {
	group := g.Group("/oauth")
	group.GET("/info", a.GetOAuthInfo)
	group.POST("/token", a.GetUserToken)
	group.POST("/refresh_token", a.RefreshUserToken)
	group.GET("/userinfo", a.GetUserInfo)
}

func (a *OAuthApi) GetOAuthInfo(ctx *gin.Context) {
	info, err := a.oauthSvc.GetOAuthInfo(ctx.Query("oauth2_key"), ctx.Query("state"))
	if err != nil {
		logger.Error("get oauth info err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx, info)
}

func (a *OAuthApi) GetUserToken(ctx *gin.Context) {
	var req dto.ReqOAuthToken
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("get oauth token err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	token, err := a.oauthSvc.GetUserToken(ctx.Request.Context(), req.OAuth2Key, req.Code)
	if err != nil {
		logger.Error("get oauth token err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx, token)
}

func (a *OAuthApi) RefreshUserToken(ctx *gin.Context) {
	var req dto.ReqOAuthRefreshToken
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("refresh oauth token err", "err", err)
		handler.HandlerError(ctx, response.RespCodeParamErr, err)
		return
	}

	token, err := a.oauthSvc.RefreshUserToken(ctx.Request.Context(), req.OAuth2Key, req.RefreshToken)
	if err != nil {
		logger.Error("refresh oauth token err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx, token)
}

func (a *OAuthApi) GetUserInfo(ctx *gin.Context) {
	accessToken := ctx.Query("access_token")
	user, err := a.oauthSvc.GetUserInfo(ctx.Request.Context(), ctx.Query("oauth2_key"), accessToken)
	if err != nil {
		logger.Error("get oauth user info err", "err", err)
		handler.HandlerError(ctx, response.RespCodeInternalErr, err)
		return
	}
	handler.Success(ctx, user)
}

package svc

import (
	"context"
	"strings"
	"time"

	"edgelink-api/internal/api/dto"
	bizErr "edgelink-api/internal/pkg/bizerr"

	oauth2Client "github.com/Ri0nGo/gokit/iam/oauth2Client"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

const oauthOpenIDKeyPrefix = "edgelink:oauth:openid:"

type IOAuthSvc interface {
	GetOAuthInfo(state string) (*dto.OAuthInfo, error)
	GetUserToken(ctx context.Context, code string) (*oauth2Client.TokenResponse, error)
	RefreshUserToken(ctx context.Context, refreshToken string) (*oauth2Client.TokenResponse, error)
	GetUserInfo(ctx context.Context, accessToken string) (*oauth2Client.UserInfo, error)
}

type OAuthSvc struct {
	enabled     bool
	clientID    string
	redirectURI string
	scopes      []string
	redis       redis.Cmdable
	client      *oauth2Client.Client
}

func NewOAuthSvc(redis redis.Cmdable) IOAuthSvc {
	enabled := viper.GetBool("oauth.enabled")
	clientID := viper.GetString("oauth.client_id")
	redirectURI := viper.GetString("oauth.redirect_uri")
	scopes := getOAuthScopes()

	var client *oauth2Client.Client
	if enabled {
		var err error
		client, err = oauth2Client.New(oauth2Client.Config{
			BaseURL:      viper.GetString("oauth.auth_base_url"),
			ClientID:     clientID,
			ClientSecret: viper.GetString("oauth.client_secret"),
			RedirectURI:  redirectURI,
			Scopes:       scopes,
			Timeout:      time.Duration(viper.GetInt("oauth.timeout_seconds")) * time.Second,
		})
		if err != nil {
			panic(err)
		}
	}

	return &OAuthSvc{
		enabled:     enabled,
		clientID:    clientID,
		redirectURI: redirectURI,
		scopes:      scopes,
		redis:       redis,
		client:      client,
	}
}

func (s *OAuthSvc) GetOAuthInfo(state string) (*dto.OAuthInfo, error) {
	info := &dto.OAuthInfo{
		Enabled:      s.enabled,
		ClientID:     s.clientID,
		RedirectURI:  s.redirectURI,
		ResponseType: "code",
		Scope:        strings.Join(s.scopes, " "),
	}
	if !s.enabled {
		return info, nil
	}
	if s.client == nil {
		return nil, bizErr.NewBizError("OAuth2客户端未初始化")
	}

	authURL, err := s.client.AuthCodeURL(state)
	if err != nil {
		return nil, err
	}
	info.AuthURL = authURL
	return info, nil
}

func (s *OAuthSvc) GetUserToken(ctx context.Context, code string) (*oauth2Client.TokenResponse, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	token, err := s.client.GetUserToken(ctx, code)
	if err != nil {
		return nil, bizErr.NewBizError(err.Error())
	}
	if err := s.saveOpenID(ctx, token); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *OAuthSvc) RefreshUserToken(ctx context.Context, refreshToken string) (*oauth2Client.TokenResponse, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	token, err := s.client.RefreshUserToken(ctx, refreshToken)
	if err != nil {
		return nil, bizErr.NewBizError(err.Error())
	}
	if err := s.saveOpenID(ctx, token); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *OAuthSvc) GetUserInfo(ctx context.Context, accessToken string) (*oauth2Client.UserInfo, error) {
	if err := s.ensureEnabled(); err != nil {
		return nil, err
	}
	openID, err := s.getOpenID(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	user, err := s.client.GetUserInfo(ctx, accessToken, openID)
	if err != nil {
		return nil, bizErr.NewBizError(err.Error())
	}
	return user, nil
}

func (s *OAuthSvc) saveOpenID(ctx context.Context, token *oauth2Client.TokenResponse) error {
	if token == nil || strings.TrimSpace(token.AccessToken) == "" || strings.TrimSpace(token.OpenID) == "" {
		return bizErr.NewBizError("OAuth2 token缺少access_token或openid")
	}
	if s.redis == nil {
		return bizErr.NewBizError("Redis未初始化")
	}
	ttl := time.Duration(token.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = 2 * time.Hour
	}
	if err := s.redis.Set(ctx, oauthOpenIDKeyPrefix+token.AccessToken, token.OpenID, ttl).Err(); err != nil {
		return err
	}
	return nil
}

func (s *OAuthSvc) getOpenID(ctx context.Context, accessToken string) (string, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return "", bizErr.NewBizError("access_token不能为空")
	}
	if s.redis == nil {
		return "", bizErr.NewBizError("Redis未初始化")
	}
	openID, err := s.redis.Get(ctx, oauthOpenIDKeyPrefix+accessToken).Result()
	if err != nil {
		return "", bizErr.NewBizError("access_token无效或已过期")
	}
	return openID, nil
}

func (s *OAuthSvc) ensureEnabled() error {
	if !s.enabled {
		return bizErr.NewBizError("OAuth2未启用")
	}
	if s.client == nil {
		return bizErr.NewBizError("OAuth2客户端未初始化")
	}
	return nil
}

func getOAuthScopes() []string {
	scopes := viper.GetStringSlice("oauth.scopes")
	if len(scopes) > 0 {
		return scopes
	}
	return strings.Fields(viper.GetString("oauth.scope"))
}

package svc

import (
	"context"
	"fmt"
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
	GetOAuthInfo(oauth2Key, state string) (*dto.OAuthInfo, error)
	GetUserToken(ctx context.Context, oauth2Key, code string) (*oauth2Client.TokenResponse, error)
	RefreshUserToken(ctx context.Context, oauth2Key, refreshToken string) (*oauth2Client.TokenResponse, error)
	GetUserInfo(ctx context.Context, oauth2Key, accessToken string) (*oauth2Client.UserInfo, error)
}

type oauthClientConfig struct {
	key         string
	enabled     bool
	clientID    string
	redirectURI string
	scopes      []string
	client      *oauth2Client.Client
}

type OAuthSvc struct {
	redis   redis.Cmdable
	clients map[string]*oauthClientConfig
}

func NewOAuthSvc(redis redis.Cmdable) IOAuthSvc {
	clients := make(map[string]*oauthClientConfig)
	clientConfigs := viper.GetStringMap("oauth.clients")
	for oauth2Key := range clientConfigs {
		clients[oauth2Key] = mustBuildOAuthClient(oauth2Key, viper.Sub("oauth.clients."+oauth2Key))
	}

	// Backward compatibility for local configs that have not been migrated yet.
	if len(clients) == 0 && strings.TrimSpace(viper.GetString("oauth.client_id")) != "" {
		clients["default"] = mustBuildOAuthClient("default", viper.Sub("oauth"))
	}

	return &OAuthSvc{redis: redis, clients: clients}
}

func mustBuildOAuthClient(oauth2Key string, cfg *viper.Viper) *oauthClientConfig {
	if cfg == nil {
		panic(fmt.Sprintf("OAuth2配置不存在: %s", oauth2Key))
	}

	clientCfg := &oauthClientConfig{
		key:         oauth2Key,
		enabled:     cfg.GetBool("enabled"),
		clientID:    cfg.GetString("client_id"),
		redirectURI: cfg.GetString("redirect_uri"),
		scopes:      getOAuthScopes(cfg),
	}
	if !clientCfg.enabled {
		return clientCfg
	}

	client, err := oauth2Client.New(oauth2Client.Config{
		BaseURL:      cfg.GetString("auth_base_url"),
		ClientID:     clientCfg.clientID,
		ClientSecret: cfg.GetString("client_secret"),
		RedirectURI:  clientCfg.redirectURI,
		Scopes:       clientCfg.scopes,
		Timeout:      time.Duration(cfg.GetInt("timeout_seconds")) * time.Second,
	})
	if err != nil {
		panic(err)
	}
	clientCfg.client = client
	return clientCfg
}

func (s *OAuthSvc) GetOAuthInfo(oauth2Key, state string) (*dto.OAuthInfo, error) {
	clientCfg, err := s.getClient(oauth2Key)
	if err != nil {
		return nil, err
	}

	info := &dto.OAuthInfo{
		Enabled:      clientCfg.enabled,
		OAuth2Key:    clientCfg.key,
		ClientID:     clientCfg.clientID,
		RedirectURI:  clientCfg.redirectURI,
		ResponseType: "code",
		Scope:        strings.Join(clientCfg.scopes, " "),
	}
	if !clientCfg.enabled {
		return info, nil
	}
	if clientCfg.client == nil {
		return nil, bizErr.NewBizError("OAuth2客户端未初始化")
	}

	authURL, err := clientCfg.client.AuthCodeURL(state)
	if err != nil {
		return nil, err
	}
	info.AuthURL = authURL
	return info, nil
}

func (s *OAuthSvc) GetUserToken(ctx context.Context, oauth2Key, code string) (*oauth2Client.TokenResponse, error) {
	clientCfg, err := s.ensureEnabled(oauth2Key)
	if err != nil {
		return nil, err
	}
	token, err := clientCfg.client.GetUserToken(ctx, code)
	if err != nil {
		return nil, bizErr.NewBizError(err.Error())
	}
	if err := s.saveOpenID(ctx, token); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *OAuthSvc) RefreshUserToken(ctx context.Context, oauth2Key, refreshToken string) (*oauth2Client.TokenResponse, error) {
	clientCfg, err := s.ensureEnabled(oauth2Key)
	if err != nil {
		return nil, err
	}
	token, err := clientCfg.client.RefreshUserToken(ctx, refreshToken)
	if err != nil {
		return nil, bizErr.NewBizError(err.Error())
	}
	if err := s.saveOpenID(ctx, token); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *OAuthSvc) GetUserInfo(ctx context.Context, oauth2Key, accessToken string) (*oauth2Client.UserInfo, error) {
	clientCfg, err := s.ensureEnabled(oauth2Key)
	if err != nil {
		return nil, err
	}
	openID, err := s.getOpenID(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	user, err := clientCfg.client.GetUserInfo(ctx, accessToken, openID)
	if err != nil {
		return nil, bizErr.NewBizError(err.Error())
	}
	return user, nil
}

func (s *OAuthSvc) getClient(oauth2Key string) (*oauthClientConfig, error) {
	oauth2Key = strings.TrimSpace(oauth2Key)
	if oauth2Key == "" {
		return nil, bizErr.NewBizError("oauth2_key不能为空")
	}
	clientCfg, ok := s.clients[oauth2Key]
	if !ok {
		return nil, bizErr.NewBizError(fmt.Sprintf("OAuth2配置不存在: %s", oauth2Key))
	}
	return clientCfg, nil
}

func (s *OAuthSvc) ensureEnabled(oauth2Key string) (*oauthClientConfig, error) {
	clientCfg, err := s.getClient(oauth2Key)
	if err != nil {
		return nil, err
	}
	if !clientCfg.enabled {
		return nil, bizErr.NewBizError("OAuth2未启用")
	}
	if clientCfg.client == nil {
		return nil, bizErr.NewBizError("OAuth2客户端未初始化")
	}
	return clientCfg, nil
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

func getOAuthScopes(cfg *viper.Viper) []string {
	scopes := cfg.GetStringSlice("scopes")
	if len(scopes) > 0 {
		return scopes
	}
	return strings.Fields(cfg.GetString("scope"))
}

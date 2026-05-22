package dto

type OAuthInfo struct {
	Enabled      bool   `json:"enabled"`
	OAuth2Key    string `json:"oauth2_key"`
	AuthURL      string `json:"auth_url"`
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	ResponseType string `json:"response_type"`
	Scope        string `json:"scope"`
}

type ReqOAuthToken struct {
	OAuth2Key string `json:"oauth2_key" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

type ReqOAuthRefreshToken struct {
	OAuth2Key    string `json:"oauth2_key" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

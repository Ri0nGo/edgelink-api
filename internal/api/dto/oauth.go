package dto

type OAuthInfo struct {
	Enabled      bool   `json:"enabled"`
	AuthURL      string `json:"auth_url"`
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	ResponseType string `json:"response_type"`
	Scope        string `json:"scope"`
}

type ReqOAuthToken struct {
	Code string `json:"code" binding:"required"`
}

type ReqOAuthRefreshToken struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

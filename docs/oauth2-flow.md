# OAuth2 登录流程

本文档说明 EdgeLink API 接入 IAM OAuth2 的接口、配置和前后端交互流程。

## 基本约定

EdgeLink API 对前端暴露统一前缀：

```text
/api/edgelink/oauth
```

IAM 服务端接口由 `github.com/Ri0nGo/gokit/iam/oauth2Client` 封装调用，前端不直接接触 `client_secret`。

OAuth2 回调地址配置为前端页面地址，例如：

```text
http://localhost:5173/oauth/callback
```

该地址必须与 IAM 中授权应用配置的 `redirect_uri` 完全一致。

## 配置

配置文件位置：

```text
config/config.yaml
```

示例：

```yaml
oauth:
  enabled: true
  auth_base_url: "http://localhost:8080"
  client_id: "edgelink"
  client_secret: "edgelink-secret"
  redirect_uri: "http://localhost:5173/oauth/callback"
  scopes:
    - basic
  timeout_seconds: 3
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| `enabled` | 是否启用 OAuth2 登录 |
| `auth_base_url` | IAM 服务端地址 |
| `client_id` | IAM 授权应用 ID |
| `client_secret` | IAM 授权应用密钥，仅后端使用 |
| `redirect_uri` | IAM 登录成功后的前端回调地址 |
| `scopes` | 授权范围 |
| `timeout_seconds` | 调用 IAM 接口的超时时间 |

## 接口清单

### 查询 OAuth2 登录信息

```http
GET /api/edgelink/oauth/info?state=xxx
```

用途：前端登录页获取 IAM 授权跳转地址。

响应示例：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "enabled": true,
    "auth_url": "http://localhost:8080/api/v1/oauth/authorize?client_id=edgelink&redirect_uri=http%3A%2F%2Flocalhost%3A5173%2Foauth%2Fcallback&response_type=code&scope=basic&state=xxx",
    "client_id": "edgelink",
    "redirect_uri": "http://localhost:5173/oauth/callback",
    "response_type": "code",
    "scope": "basic"
  }
}
```

说明：

- `auth_url` 由后端使用 `oauth2Client.AuthCodeURL(state)` 生成。
- `client_secret` 不会返回给前端。
- `state` 建议由前端生成并存入 `sessionStorage`，用于回调时校验，防止 CSRF。

### 使用 code 换 token

```http
POST /api/edgelink/oauth/token
Content-Type: application/json

{
  "code": "iam-returned-code"
}
```

用途：前端 OAuth2 回调页拿到 IAM 返回的 `code` 后，交给 EdgeLink API 换取 token。

响应示例：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "access_token": "xxx",
    "expires_in": 7200,
    "refresh_token": "yyy",
    "openid": "zzz",
    "scope": "basic"
  }
}
```

后端处理：

- EdgeLink API 调用 IAM token 接口。
- 请求 IAM 时由后端携带 `client_id` 和 `client_secret`。
- 换 token 成功后，EdgeLink API 将 `access_token -> openid` 写入 Redis。
- Redis key 前缀为 `edgelink:oauth:openid:`。
- Redis TTL 默认使用 IAM 返回的 `expires_in`。

### 使用 refresh_token 刷新 token

```http
POST /api/edgelink/oauth/refresh_token
Content-Type: application/json

{
  "refresh_token": "yyy"
}
```

用途：`access_token` 过期前或过期后，使用 `refresh_token` 换取新的 `access_token`。

响应示例：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "access_token": "new-access-token",
    "expires_in": 7200,
    "refresh_token": "yyy",
    "openid": "zzz",
    "scope": "basic"
  }
}
```

后端处理：

- EdgeLink API 调用 IAM refresh token 接口。
- IAM 返回新 `access_token` 后，EdgeLink API 重新写入 `new_access_token -> openid` 到 Redis。
- 前端应使用新的 `access_token` 替换本地旧 token。

### 查询用户信息

```http
GET /api/edgelink/oauth/userinfo?access_token=xxx
```

用途：前端用 `access_token` 查询当前 IAM 用户信息。

当前仅支持 query 参数传 token：

```text
?access_token=xxx
```

不支持：

```text
Authorization: Bearer xxx
```

响应示例：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "openid": "zzz",
    "username": "admin",
    "display_name": "系统管理员",
    "status": 1,
    "roles": ["admin"]
  }
}
```

后端处理：

- 最新 IAM `userinfo` 接口要求同时传 `access_token` 和 `openid`。
- 前端只传 `access_token`。
- EdgeLink API 先从 Redis 查询 `access_token -> openid`。
- EdgeLink API 再调用 IAM `userinfo?access_token=xxx&openid=zzz`。
- 如果 Redis 中找不到对应 openid，接口会返回 `access_token无效或已过期`。

## 完整登录流程

1. 前端登录页生成随机 `state`。
2. 前端将 `state` 存入 `sessionStorage`。
3. 前端调用 `GET /api/edgelink/oauth/info?state=xxx`。
4. EdgeLink API 返回 `auth_url`。
5. 前端跳转到 `auth_url`。
6. 用户在 IAM 完成登录。
7. IAM 校验授权应用、用户状态和 `redirect_uri`。
8. IAM 重定向到前端回调页：`/oauth/callback?code=xxx&state=xxx`。
9. 前端回调页读取 URL 中的 `code` 和 `state`。
10. 前端校验回调 `state` 是否等于 `sessionStorage` 中保存的值。
11. 校验通过后，前端调用 `POST /api/edgelink/oauth/token`，请求体传 `code`。
12. EdgeLink API 使用后端配置的 `client_secret` 调 IAM 换 token。
13. EdgeLink API 返回 `access_token`、`refresh_token`、`openid`、`expires_in`。
14. EdgeLink API 将 `access_token -> openid` 写入 Redis。
15. 前端保存 token 信息并进入系统。
16. 前端需要用户信息时调用 `GET /api/edgelink/oauth/userinfo?access_token=xxx`。

## Token 刷新流程

1. 前端记录 `access_token` 的过期时间。
2. 前端在过期前 1 到 5 分钟主动调用 `POST /api/edgelink/oauth/refresh_token`。
3. 或者当前端收到 token 过期类错误时，被动调用刷新接口。
4. EdgeLink API 使用 `refresh_token` 调 IAM 获取新 `access_token`。
5. EdgeLink API 将新 `access_token -> openid` 写入 Redis。
6. 前端用新 token 替换旧 token。
7. 如果刷新失败，前端清空登录态并重新跳转 IAM 登录。

## 与 IAM 的内部交互

EdgeLink API 当前通过 `gokit` 使用这些 IAM OAuth2 接口：

```text
GET /api/v1/oauth/authorize
GET /api/v1/oauth/token
GET /api/v1/oauth/refresh_token
GET /api/v1/oauth/userinfo
```

IAM 最新服务端行为：

- `authorize` 校验用户登录 Cookie 和授权应用，成功后重定向到前端回调地址。
- `token` 返回统一响应包裹，`data` 中包含 `access_token`、`refresh_token`、`openid`。
- `refresh_token` 使用长期 refresh token 换新 access token。
- `userinfo` 要求 `access_token` 和 `openid` 同时存在。

## 前端建议

- `state` 必须随机生成并在回调页校验。
- `client_secret` 永远不要放到前端。
- 登录成功后保存 `access_token`、`refresh_token`、`expires_in`。
- 建议根据 `expires_in` 主动刷新 token。
- 如果使用浏览器本地存储 token，需要注意 XSS 风险。
- 更安全的长期方案是后端管理 refresh token，并通过 HttpOnly Cookie 维护登录态。

## 常见失败场景

| 场景 | 可能原因 |
| --- | --- |
| `oauth/info` 返回未启用 | `oauth.enabled=false` |
| IAM 登录后未回调前端 | IAM 应用配置的 `redirect_uri` 与请求不一致 |
| 使用 code 换 token 失败 | code 过期、重复使用、应用密钥错误 |
| `userinfo` 提示 token 无效或已过期 | Redis 中没有 `access_token -> openid` 映射，或 access token 已过期 |
| refresh token 失败 | refresh token 无效、过期或不属于当前应用 |

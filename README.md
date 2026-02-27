# EdgeLink-API

连接边缘设备的 API 服务

## 项目简介

EdgeLink-API 是一个基于 Go 语言开发的边缘设备连接管理平台，提供设备管理、物模型管理、产品管理以及时序数据查询等功能。通过 MQTT 协议与边缘设备进行通信，支持设备的配置管理和数据持久化。

## 主要功能

- **设备管理**: 设备的创建、更新、删除、查询，支持设备属性配置
- **产品管理**: 产品定义与管理，关联物模型
- **物模型管理**: 定义设备功能类型和属性，支持属性的增删改查
- **时序数据**: 查询设备历史数据
- **MQTT 通信**: 与边缘设备进行消息通信
- **数据持久化**: 支持 MySQL 数据库和 Redis 缓存

## 技术栈

- **语言**: Go 1.25.4
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL
- **缓存**: Redis
- **消息队列**: MQTT (Paho)
- **依赖注入**: Google Wire
- **配置管理**: Viper
- **日志**: 结构化日志

## 项目结构

```
edgelink-api/
├── cmd/                      # 应用程序入口
├── config/                   # 配置文件
│   ├── config.yaml          # 开发环境配置
│   └── prod.yaml            # 生产环境配置
├── internal/                 # 内部包
│   ├── api/                 # API 层 (Handler)
│   │   ├── device.go        # 设备 API
│   │   ├── product.go       # 产品 API
│   │   ├── thing_model.go   # 物模型 API
│   │   ├── metric.go        # 指标 API
│   │   └── dto/             # 数据传输对象
│   ├── bootstrap/           # 应用启动和停止
│   ├── dataloader/          # 数据加载器
│   ├── infrastructure/      # 基础设施层
│   │   ├── cache/           # Redis 缓存
│   │   ├── config/          # 配置加载
│   │   └── db/              # 数据库连接
│   ├── ioc/                 # 依赖注入配置
│   ├── model/               # 数据模型
│   ├── pkg/                 # 公共包
│   │   ├── bizerr/          # 业务错误
│   │   ├── ginx/            # Gin 扩展
│   │   ├── logger/          # 日志
│   │   └── paginate/        # 分页
│   ├── repo/                # 数据访问层
│   ├── router/              # 路由配置
│   ├── svc/                 # 业务逻辑层
│   └── utils/               # 工具函数
├── scripts/                  # 脚本文件
│   ├── lint/                # 代码规范配置
│   ├── mysql/               # 数据库脚本
│   └── setup.sh             # 环境初始化脚本
├── test/                     # 测试文件
├── main.go                   # 程序入口
├── go.mod                    # Go 模块定义
├── Makefile                  # Make 命令
└── run.log                   # 运行日志
```

## 快速开始

### 环境要求

- Go 1.25.4+
- MySQL 5.7+
- Redis 6.0+
- MQTT Broker (如 EMQX、Mosquitto)

### 安装依赖

```bash
# 安装 Go 依赖
go mod tidy

# 安装开发工具
make setup
```

### 配置

编辑 `config/config.yaml` 文件，配置数据库、Redis 和 MQTT 连接信息:

```yaml
databases:
  edgelink:
    host: 127.0.0.1
    port: 3306
    username: root
    password: 123456
    dbname: edgelink

redis:
  addr: "127.0.0.1:6379"
  password: "123456"
  db: 0

mqtt:
  host: 127.0.0.1
  port: 1883
  username: admin
  password: 123456
```

### 运行

```bash
# 使用 Makefile 运行
make run

# 或直接运行
go run main.go

# 指定配置文件路径
go run main.go -c ./config/config.yaml
```

服务默认启动在 `:8080` 端口

### API 路由

所有 API 路由前缀为 `/api/edgelink`

| 模块 | 路由前缀 | 说明 |
|------|----------|------|
| 物模型 | `/thing_model` | 物模型及其属性管理 |
| 产品 | `/product` | 产品管理 |
| 设备 | `/device` | 设备管理及属性配置 |
| 数据 | `/data` | 时序数据查询 |

## 开发命令

```bash
# 代码格式化
go fmt ./...

# 依赖整理
make tidy

# 代码规范检查
make lint

# 生成依赖注入代码
make wire

# 运行测试
go test ./...
```

## 架构分层

- **API 层** (`internal/api`): 处理 HTTP 请求，参数校验，响应格式化
- **Service 层** (`internal/svc`): 业务逻辑处理
- **Repo 层** (`internal/repo`): 数据访问，数据库操作
- **Model 层** (`internal/model`): 数据模型定义
- **Infrastructure 层** (`internal/infrastructure`): 基础设施 (数据库、缓存、配置)

## 依赖注入

项目使用 [Google Wire](https://github.com/google/wire) 进行依赖注入。修改依赖后需要重新生成:

```bash
make wire
```

## 日志

日志配置在配置文件中指定，支持以下配置项:

- `logger.level`: 日志级别 (debug, info, warn, error)
- `logger.format`: 日志格式 (json, console)
- `logger.filepath`: 日志文件路径
- `logger.show_source`: 是否显示日志来源

## 测试

```bash
# 运行所有测试
go test ./...

# 运行指定包测试
go test ./internal/svc/...

# 带覆盖率
go test -cover ./...
```

## 许可证

MIT License

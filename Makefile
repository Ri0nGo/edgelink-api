# 清理项目依赖
.PHONY: tidy
tidy:
	@go mod tidy -v

# 初始化项目环境
.PHONY: setup
setup:
	@sh ./scripts/setup.sh

# 代码规范检查
.PHONY: lint
lint:
	@golangci-lint run -c ./scripts/lint/.golangci.yaml ./...


.PHONY: wire
wire:
	wire gen internal/ioc/wire.go
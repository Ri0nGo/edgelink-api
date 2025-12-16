#!/bin/sh

echo "安装 golangci-lint..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

echo "安装 mockgen..."
go install go.uber.org/mock/mockgen@latest

echo "安装 wire..."
go install github.com/google/wire/cmd/wire@latest

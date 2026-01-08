package bootstrap

import (
	"context"
	"edgelink-api/internal/ioc"
	"net/http"
)

func Bootstrap(ctx context.Context, container *ioc.Container, srv *http.Server) {
	RunWebServer(srv)

	// 初始化数据接收器，处理器，存储器
	dataLoaderC := InitDataLoader(ctx, container.MysqlDB, container.RedisCmd)

	// 初始化通知器（设备配置和属性通知）
	InitNotify(ctx, container.MysqlDB, container.RedisCmd, dataLoaderC)

}

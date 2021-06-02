package easytcp

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/sirupsen/logrus"
)

// SetLogger 设置日志对象
func SetLogger(log *logrus.Logger) {
	logger.Default = log
}

// NewTcp 创建 tcp server
func NewTcp(opt server.TcpOption) *server.TcpServer {
	return server.NewTcp(opt)
}

func NewUdp(opt server.UdpOption) *server.UdpServer {
	return server.NewUdp(opt)
}

// RegisterRoute 注册消息路由
func RegisterRoute(msgId uint, handler router.HandlerFunc, middleware ...router.MiddlewareFunc) {
	router.Instance().Register(msgId, handler, middleware...)
}

// RegisterMiddleware 注册全局中间件
func RegisterMiddleware(m ...router.MiddlewareFunc) {
	router.Instance().RegisterMiddleware(m...)
}

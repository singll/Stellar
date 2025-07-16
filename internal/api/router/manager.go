package router

import (
	"github.com/gin-gonic/gin"
	"github.com/StellarServer/internal/api/middleware"
)

// RouteManager 统一路由管理器
type RouteManager struct {
	engine   *gin.Engine
	apiGroup *gin.RouterGroup
	groups   map[string]*gin.RouterGroup
}

// NewRouteManager 创建新的路由管理器
func NewRouteManager(engine *gin.Engine) *RouteManager {
	// 创建API v1组
	apiV1 := engine.Group("/api/v1")
	
	return &RouteManager{
		engine:   engine,
		apiGroup: apiV1,
		groups:   make(map[string]*gin.RouterGroup),
	}
}

// ApplyGlobalMiddleware 应用全局中间件
func (rm *RouteManager) ApplyGlobalMiddleware() {
	rm.engine.Use(middleware.Recovery())
	rm.engine.Use(middleware.RequestLogger())
	rm.engine.Use(middleware.CORS())
	rm.engine.Use(middleware.Security())
}

// RegisterGroup 注册路由组
func (rm *RouteManager) RegisterGroup(name string, path string, handlers ...RouteHandler) *gin.RouterGroup {
	group := rm.apiGroup.Group(path)
	rm.groups[name] = group
	
	// 注册所有处理器到该组
	for _, handler := range handlers {
		handler.RegisterRoutes(group)
	}
	
	return group
}

// RegisterTopLevel 注册顶级路由（如健康检查）
func (rm *RouteManager) RegisterTopLevel(method, path string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		rm.engine.GET(path, handler)
	case "POST":
		rm.engine.POST(path, handler)
	case "PUT":
		rm.engine.PUT(path, handler)
	case "DELETE":
		rm.engine.DELETE(path, handler)
	}
}

// GetGroup 获取已注册的路由组
func (rm *RouteManager) GetGroup(name string) *gin.RouterGroup {
	return rm.groups[name]
}

// RegisterAuthRoutes 注册需要认证的路由组
func (rm *RouteManager) RegisterAuthGroup(name string, path string, handlers ...RouteHandler) *gin.RouterGroup {
	group := rm.apiGroup.Group(path)
	group.Use(middleware.AuthMiddleware()) // 应用认证中间件
	rm.groups[name] = group
	
	// 注册所有处理器到该组
	for _, handler := range handlers {
		handler.RegisterRoutes(group)
	}
	
	return group
}

// RegisterPublicRoutes 注册公开路由组（无需认证）
func (rm *RouteManager) RegisterPublicGroup(name string, path string, handlers ...RouteHandler) *gin.RouterGroup {
	group := rm.apiGroup.Group(path)
	rm.groups[name] = group
	
	// 注册所有处理器到该组
	for _, handler := range handlers {
		handler.RegisterRoutes(group)
	}
	
	return group
}

// GetEngine 获取Gin引擎
func (rm *RouteManager) GetEngine() *gin.Engine {
	return rm.engine
}

// RouteHandler 路由处理器接口
type RouteHandler interface {
	RegisterRoutes(group *gin.RouterGroup)
}
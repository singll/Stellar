package app

import (
	"github.com/StellarServer/internal/config"
	"github.com/StellarServer/internal/database"
	"github.com/StellarServer/internal/pkg/container"
	"github.com/StellarServer/internal/pkg/logger"
	"github.com/StellarServer/internal/repository"
)

// Application 应用程序结构
type Application struct {
	Container *container.Container
	Config    *config.Config
	DB        *database.DB
	Repo      *repository.Repository
}

// NewApplication 创建新的应用程序实例
func NewApplication(cfg *config.Config) (*Application, error) {
	app := &Application{
		Container: container.NewContainer(),
		Config:    cfg,
	}

	// 初始化日志
	logConfig := logger.Config{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}
	err := logger.Init(logConfig)
	if err != nil {
		return nil, err
	}

	// 初始化数据库
	db, err := database.NewDB(cfg)
	if err != nil {
		return nil, err
	}
	app.DB = db

	// 初始化仓储
	repo := repository.NewRepository(db)
	app.Repo = repo

	// 注册服务到容器
	err = app.registerServices()
	if err != nil {
		return nil, err
	}

	return app, nil
}

// registerServices 注册服务到依赖注入容器
func (app *Application) registerServices() error {
	// 注册配置
	app.Container.Register("config", app.Config)

	// 注册数据库
	app.Container.Register("database", app.DB)

	// 注册仓储
	app.Container.Register("repository", app.Repo)

	// 注册业务服务
	app.registerBusinessServices()

	// 注册API处理器
	app.registerHandlers()

	return nil
}

// registerBusinessServices 注册业务服务
func (app *Application) registerBusinessServices() {
	// TODO: 注册具体的业务服务
	// 例如：
	// app.Container.RegisterSingleton("userService", func(c *container.Container) (interface{}, error) {
	//     repo := c.MustGet("repository").(*repository.Repository)
	//     return services.NewUserService(repo), nil
	// })
}

// registerHandlers 注册API处理器
func (app *Application) registerHandlers() {
	// TODO: 注册API处理器
	// 例如：
	// app.Container.RegisterSingleton("userHandler", func(c *container.Container) (interface{}, error) {
	//     userService := c.MustGet("userService").(services.UserService)
	//     return handlers.NewUserHandler(userService), nil
	// })
}

// Shutdown 关闭应用程序
func (app *Application) Shutdown() error {
	logger.Info("Shutting down application", nil)

	// 关闭数据库连接
	if app.DB != nil {
		if err := app.DB.Close(); err != nil {
			logger.Error("Failed to close database", map[string]interface{}{
				"error": err,
			})
			return err
		}
	}

	logger.Info("Application shutdown complete", nil)
	return nil
}
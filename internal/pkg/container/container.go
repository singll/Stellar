package container

import (
	"fmt"
	"reflect"
	"sync"
)

// Container 依赖注入容器
type Container struct {
	mu        sync.RWMutex
	services  map[string]interface{}
	factories map[string]func(*Container) (interface{}, error)
	singletons map[string]interface{}
}

// NewContainer 创建新的容器
func NewContainer() *Container {
	return &Container{
		services:   make(map[string]interface{}),
		factories:  make(map[string]func(*Container) (interface{}, error)),
		singletons: make(map[string]interface{}),
	}
}

// Register 注册服务实例
func (c *Container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

// RegisterFactory 注册服务工厂
func (c *Container) RegisterFactory(name string, factory func(*Container) (interface{}, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[name] = factory
}

// RegisterSingleton 注册单例服务
func (c *Container) RegisterSingleton(name string, factory func(*Container) (interface{}, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[name] = func(container *Container) (interface{}, error) {
		// 检查单例是否已创建
		if instance, exists := c.singletons[name]; exists {
			return instance, nil
		}
		
		// 创建新实例
		instance, err := factory(container)
		if err != nil {
			return nil, err
		}
		
		// 存储单例
		c.singletons[name] = instance
		return instance, nil
	}
}

// Get 获取服务
func (c *Container) Get(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 先查找已注册的服务实例
	if service, exists := c.services[name]; exists {
		return service, nil
	}

	// 查找工厂方法
	if factory, exists := c.factories[name]; exists {
		return factory(c)
	}

	return nil, fmt.Errorf("service '%s' not found", name)
}

// MustGet 获取服务，失败时panic
func (c *Container) MustGet(name string) interface{} {
	service, err := c.Get(name)
	if err != nil {
		panic(err)
	}
	return service
}

// Resolve 解析服务到指定类型
func (c *Container) Resolve(name string, target interface{}) error {
	service, err := c.Get(name)
	if err != nil {
		return err
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	serviceValue := reflect.ValueOf(service)
	targetValue.Elem().Set(serviceValue)
	return nil
}

// Inject 依赖注入到结构体
func (c *Container) Inject(target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetType := targetValue.Elem().Type()
	targetValue = targetValue.Elem()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetValue.Field(i)

		// 检查是否有inject标签
		injectTag := field.Tag.Get("inject")
		if injectTag == "" {
			continue
		}

		// 如果字段不可设置，跳过
		if !fieldValue.CanSet() {
			continue
		}

		// 获取服务
		service, err := c.Get(injectTag)
		if err != nil {
			return fmt.Errorf("failed to inject field '%s': %v", field.Name, err)
		}

		// 设置字段值
		serviceValue := reflect.ValueOf(service)
		if !serviceValue.Type().AssignableTo(fieldValue.Type()) {
			return fmt.Errorf("service '%s' is not assignable to field '%s'", injectTag, field.Name)
		}

		fieldValue.Set(serviceValue)
	}

	return nil
}

// Call 调用函数并注入依赖
func (c *Container) Call(fn interface{}, args ...interface{}) ([]reflect.Value, error) {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	// 准备参数
	var params []reflect.Value
	
	// 添加额外参数
	for _, arg := range args {
		params = append(params, reflect.ValueOf(arg))
	}

	// 自动注入剩余参数
	for i := len(params); i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		
		// 尝试通过类型名称查找服务
		serviceName := paramType.String()
		service, err := c.Get(serviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parameter %d: %v", i, err)
		}
		
		params = append(params, reflect.ValueOf(service))
	}

	// 调用函数
	return fnValue.Call(params), nil
}

// Has 检查服务是否存在
func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	_, exists := c.services[name]
	if exists {
		return true
	}
	
	_, exists = c.factories[name]
	return exists
}

// Remove 移除服务
func (c *Container) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.services, name)
	delete(c.factories, name)
	delete(c.singletons, name)
}

// Clear 清空所有服务
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.services = make(map[string]interface{})
	c.factories = make(map[string]func(*Container) (interface{}, error))
	c.singletons = make(map[string]interface{})
}

// 全局容器实例
var defaultContainer = NewContainer()

// 全局函数，使用默认容器
func Register(name string, service interface{}) {
	defaultContainer.Register(name, service)
}

func RegisterFactory(name string, factory func(*Container) (interface{}, error)) {
	defaultContainer.RegisterFactory(name, factory)
}

func RegisterSingleton(name string, factory func(*Container) (interface{}, error)) {
	defaultContainer.RegisterSingleton(name, factory)
}

func Get(name string) (interface{}, error) {
	return defaultContainer.Get(name)
}

func MustGet(name string) interface{} {
	return defaultContainer.MustGet(name)
}

func Resolve(name string, target interface{}) error {
	return defaultContainer.Resolve(name, target)
}

func Inject(target interface{}) error {
	return defaultContainer.Inject(target)
}

func Has(name string) bool {
	return defaultContainer.Has(name)
}

func Remove(name string) {
	defaultContainer.Remove(name)
}

func Clear() {
	defaultContainer.Clear()
}
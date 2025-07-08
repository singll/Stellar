# StellarServer 插件系统

StellarServer 插件系统是一个灵活、可扩展的框架，允许开发者通过插件扩展 StellarServer 的功能。

## 设计目标

- **灵活性**：支持多种插件类型和功能
- **可扩展性**：允许轻松添加新的插件类型和功能
- **安全性**：提供沙箱环境，确保插件运行安全
- **多语言支持**：支持使用不同编程语言开发插件
- **版本管理**：支持插件版本管理和依赖处理
- **易用性**：简化插件开发和使用流程

## 架构概述

插件系统由以下核心组件组成：

1. **插件接口**：定义插件的标准接口和生命周期方法
2. **插件注册表**：管理已安装的插件
3. **插件加载器**：负责加载不同类型的插件
4. **插件管理器**：提供插件的安装、卸载、启用和禁用功能
5. **插件沙箱**：提供安全的插件运行环境
6. **插件元数据存储**：存储插件的元数据信息

## 插件类型

StellarServer 支持以下类型的插件：

- **漏洞扫描插件**：用于检测安全漏洞
- **资产发现插件**：用于发现网络资产
- **端口扫描插件**：用于扫描开放端口
- **子域名枚举插件**：用于枚举子域名
- **敏感信息检测插件**：用于检测敏感信息
- **监控插件**：用于监控网页变更
- **工具类插件**：提供各种辅助功能

## 插件格式

StellarServer 支持以下格式的插件：

- **Go 插件**：使用 Go 语言开发的插件（.so 文件）
- **脚本插件**：使用 Python、JavaScript 等脚本语言开发的插件
- **YAML 插件**：使用 YAML 定义的简单插件

## 插件开发指南

### 创建 Go 插件

1. 创建一个实现 `Plugin` 接口的结构体：

```go
package main

import (
    "github.com/StellarServer/internal/plugin"
)

type MyPlugin struct {
    info   plugin.PluginInfo
    config map[string]interface{}
}

var Plugin = &MyPlugin{
    info: plugin.PluginInfo{
        ID:          "my_plugin",
        Name:        "My Plugin",
        Version:     "1.0.0",
        Type:        plugin.TypeUtility,
        Author:      "Your Name",
        Description: "My awesome plugin",
        // ...其他字段
    },
    config: make(map[string]interface{}),
}

// 实现 Plugin 接口的方法
func (p *MyPlugin) Info() plugin.PluginInfo {
    return p.info
}

func (p *MyPlugin) Init(config map[string]interface{}) error {
    p.config = config
    return nil
}

func (p *MyPlugin) Execute(ctx plugin.PluginContext) (plugin.PluginResult, error) {
    // 实现插件功能
    return plugin.PluginResult{
        Success: true,
        Data:    "Hello from my plugin!",
    }, nil
}

func (p *MyPlugin) Validate(params map[string]interface{}) error {
    // 验证参数
    return nil
}

func (p *MyPlugin) Cleanup() error {
    // 清理资源
    return nil
}
```

2. 构建插件：

```bash
go build -buildmode=plugin -o my_plugin.so my_plugin.go
```

### 创建 Python 插件

创建一个 Python 脚本，实现以下函数：

```python
#!/usr/bin/env python3

import json
import sys

# 插件信息
PLUGIN_INFO = {
    "id": "my_python_plugin",
    "name": "My Python Plugin",
    "version": "1.0.0",
    "type": "utility",
    # ...其他字段
}

def info():
    """返回插件信息"""
    return PLUGIN_INFO

def init(config_json):
    """初始化插件"""
    config = json.loads(config_json)
    # 初始化逻辑
    return {"success": True}

def execute(params_json):
    """执行插件"""
    params = json.loads(params_json)
    # 执行逻辑
    result = {
        "success": True,
        "data": "Hello from Python plugin!",
    }
    return json.dumps(result)

def validate(params_json):
    """验证参数"""
    params = json.loads(params_json)
    # 验证逻辑
    return json.dumps({"success": True})

def cleanup():
    """清理资源"""
    # 清理逻辑
    return json.dumps({"success": True})

if __name__ == "__main__":
    # 命令行接口
    if len(sys.argv) < 2:
        print("Usage: my_plugin.py <command> [args]")
        sys.exit(1)
    
    command = sys.argv[1]
    
    if command == "info":
        print(json.dumps(info()))
    elif command == "init":
        config_json = sys.stdin.read()
        print(json.dumps(init(config_json)))
    elif command == "execute":
        params_json = sys.stdin.read()
        print(execute(params_json))
    elif command == "validate":
        params_json = sys.stdin.read()
        print(validate(params_json))
    elif command == "cleanup":
        print(cleanup())
    else:
        print(json.dumps({
            "success": False,
            "error": f"Unknown command: {command}"
        }))
        sys.exit(1)
```

### 创建 YAML 插件

创建一个 YAML 文件，定义插件：

```yaml
id: my_yaml_plugin
name: My YAML Plugin
version: 1.0.0
type: utility
author: Your Name
description: My awesome YAML plugin
# ...其他字段

params:
  - name: name
    type: string
    description: Your name
    required: true
    default: World

script:
  language: python
  code: |
    def execute(params):
        name = params.get('name', 'World')
        return {
            'success': True,
            'data': f"Hello, {name}!",
        }

    def validate(params):
        if 'name' not in params:
            return {
                'success': False,
                'error': "Missing required parameter: name"
            }
        return {'success': True}
```

## 插件安装和管理

### 安装插件

```go
// 创建插件管理器
registry := plugin.NewRegistry()
metadataStore := plugin.NewMongoMetadataStore(db, "plugins")
manager := plugin.NewManager(registry, metadataStore)

// 添加插件目录
manager.AddPluginDirectory("/path/to/plugins")

// 加载插件
if err := manager.LoadPlugins(); err != nil {
    log.Fatalf("加载插件失败: %v", err)
}

// 安装插件
pluginID, err := manager.InstallPlugin("/path/to/my_plugin.so", "/path/to/plugins")
if err != nil {
    log.Fatalf("安装插件失败: %v", err)
}
```

### 使用插件

```go
// 获取插件
p, err := manager.GetPlugin("my_plugin")
if err != nil {
    log.Fatalf("获取插件失败: %v", err)
}

// 初始化插件
if err := p.Init(map[string]interface{}{
    "option1": "value1",
}); err != nil {
    log.Fatalf("初始化插件失败: %v", err)
}

// 验证参数
params := map[string]interface{}{
    "name": "John",
}
if err := p.Validate(params); err != nil {
    log.Fatalf("验证参数失败: %v", err)
}

// 执行插件
ctx := plugin.PluginContext{
    Context: context.Background(),
    Timeout: 30 * time.Second,
    Params:  params,
}
result, err := p.Execute(ctx)
if err != nil {
    log.Fatalf("执行插件失败: %v", err)
}

// 处理结果
if result.Success {
    log.Printf("插件执行成功: %v", result.Data)
} else {
    log.Printf("插件执行失败: %s", result.Error)
}
```

## 安全性考虑

插件系统使用沙箱环境运行插件，以确保安全性：

- 限制插件的文件系统访问
- 限制插件的网络访问
- 限制插件的资源使用（CPU、内存等）
- 设置插件执行超时
- 验证插件的完整性和来源

## 插件开发最佳实践

1. **提供详细的文档**：描述插件的功能、参数和使用方法
2. **处理错误**：妥善处理错误并提供有用的错误信息
3. **资源管理**：确保插件正确释放资源
4. **参数验证**：验证输入参数，确保安全性和正确性
5. **版本兼容性**：明确声明插件的版本和兼容性要求
6. **测试**：全面测试插件的功能和性能
7. **安全性**：遵循安全最佳实践，避免常见的安全漏洞 
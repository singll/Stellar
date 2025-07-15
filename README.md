# 星络 (Stellar) - 分布式安全资产管理平台

<div align="center">

![Stellar Logo](https://via.placeholder.com/400x200/2563eb/ffffff?text=Stellar+%E6%98%9F%E7%BB%9C)

[![Go Version](https://img.shields.io/badge/Go-1.24.3-00ADD8?logo=go)](https://golang.org/)
[![Svelte Version](https://img.shields.io/badge/Svelte-5.7.0-FF3E00?logo=svelte)](https://svelte.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen)](https://github.com/StellarServer/StellarServer)
[![Docker](https://img.shields.io/badge/Docker-Supported-2496ED?logo=docker)](https://docker.com/)

**现代化的分布式安全资产管理和漏洞扫描平台**

[快速开始](#快速开始) • [功能特性](#功能特性) • [技术架构](#技术架构) • [部署指南](#部署指南) • [API 文档](#api-文档) • [贡献指南](#贡献指南)

</div>

---

## 📋 项目概述

**星络 (Stellar)** 是基于 **ScopeSentry** 项目的 Go 语言重构版本，是一个现代化的分布式安全资产管理和漏洞扫描平台。项目采用前后端分离架构，提供高性能、高可用性和高扩展性的安全资产管理解决方案。

### 🎯 核心目标

- **资产发现**: 自动发现和映射网络资产
- **漏洞扫描**: 智能化的漏洞检测和评估
- **资产管理**: 全生命周期的资产管理
- **分布式架构**: 支持多节点分布式扫描
- **实时监控**: 实时资产状态监控和告警
- **插件系统**: 丰富的插件生态系统

### 🔄 项目重构

**原项目架构:**
- **后端**: Python + FastAPI + MongoDB
- **前端**: Vue 3 + Element Plus + Vite

**重构目标架构:**
- **后端**: Go + Gin + MongoDB + Redis
- **前端**: Svelte 5 + SvelteKit + TypeScript
- **API**: 保持兼容的 RESTful API (/api/v1/)

---

## ✨ 功能特性

### 🔍 资产发现与管理

#### 资产类型支持
- **🌐 域名资产**: 主域名、子域名管理
- **🖥️ 主机资产**: IP 地址、主机信息
- **🔌 端口资产**: 端口扫描、服务识别
- **🌍 URL 资产**: Web 应用资产管理
- **📱 应用资产**: Web 应用和小程序资产

#### 资产发现功能
- **子域名枚举**: 多种发现方式（DNS、证书、搜索引擎）
- **端口扫描**: 高效的端口扫描和服务识别
- **Web 应用发现**: 自动发现 Web 应用和技术栈
- **网络拓扑**: 自动构建网络拓扑图
- **资产关联**: 智能的资产关联分析

### 🛡️ 安全扫描功能

#### 漏洞扫描
- **漏洞检测**: 基于 CVE 数据库的漏洞扫描
- **Web 漏洞**: SQL 注入、XSS、CSRF 等检测
- **配置检查**: 安全配置审计
- **合规检查**: 安全合规性评估

#### 敏感信息检测
- **信息泄露**: 敏感文件和信息检测
- **API 密钥**: 各种 API 密钥泄露检测
- **配置泄露**: 配置文件和敏感信息检测
- **源码泄露**: 源代码泄露检测

### 📊 监控与告警

#### 实时监控
- **资产变化**: 实时监控资产状态变化
- **新资产发现**: 自动发现新资产并告警
- **漏洞状态**: 漏洞修复状态跟踪
- **扫描进度**: 实时扫描进度监控

#### 告警系统
- **多种通知方式**: 邮件、Webhook、企业微信
- **告警规则**: 灵活的告警规则配置
- **告警分级**: 不同级别的告警处理
- **告警历史**: 完整的告警历史记录

### 🔧 系统功能

#### 分布式架构
- **主从节点**: 支持主从节点架构
- **任务分发**: 智能任务分发和负载均衡
- **结果聚合**: 分布式扫描结果聚合
- **节点管理**: 节点状态监控和管理

#### 插件系统
- **YAML 插件**: 声明式插件定义
- **Go 插件**: 高性能编译型插件
- **Python 插件**: 灵活的脚本型插件
- **插件市场**: 丰富的插件生态

#### 用户管理
- **用户认证**: JWT 认证和权限管理
- **角色权限**: 基于角色的访问控制
- **项目管理**: 多项目隔离和管理
- **审计日志**: 完整的操作审计

---

## 🏗️ 技术架构

### 📚 技术栈

#### 后端技术栈
- **编程语言**: Go 1.24.3
- **Web 框架**: Gin 1.10+
- **数据库**: MongoDB 6.0+ + Redis 7.0+
- **认证授权**: JWT + bcrypt
- **日志系统**: zerolog
- **任务调度**: robfig/cron
- **网络库**: gorilla/websocket, miekg/dns

#### 前端技术栈
- **框架**: Svelte 5.7.0 + SvelteKit 2.0+
- **构建工具**: Vite 6.0+
- **语言**: TypeScript 5.0+
- **UI 组件**: shadcn-svelte + Tailwind CSS
- **HTTP 客户端**: axios 1.6.0
- **状态管理**: Svelte runes + TanStack Store
- **表单处理**: felte + zod

### 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                    前端层 (Frontend)                        │
├─────────────────────────────────────────────────────────────┤
│  Svelte 5 + SvelteKit + TypeScript + Tailwind CSS         │
│  • 响应式UI界面     • 实时数据展示    • 交互式图表        │
│  • 资产管理界面     • 任务控制面板    • 报告生成          │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ HTTP/WebSocket
                              │
┌─────────────────────────────────────────────────────────────┐
│                    API 网关层 (API Gateway)                │
├─────────────────────────────────────────────────────────────┤
│  Gin + JWT 认证 + 中间件                                   │
│  • 请求路由       • 权限验证       • 限流控制            │
│  • 日志记录       • 错误处理       • 跨域处理            │
└─────────────────────────────────────────────────────────────┘
                              │
                              │
┌─────────────────────────────────────────────────────────────┐
│                    业务逻辑层 (Business Logic)             │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │  资产管理   │  │  任务管理   │  │  扫描引擎   │          │
│  │  服务模块   │  │  调度模块   │  │  执行模块   │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │  漏洞管理   │  │  通知告警   │  │  报告生成   │          │
│  │  服务模块   │  │  服务模块   │  │  服务模块   │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
                              │
                              │
┌─────────────────────────────────────────────────────────────┐
│                    数据存储层 (Data Storage)               │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────┐  ┌─────────────────────────┐    │
│  │       MongoDB          │  │        Redis           │    │
│  │   • 资产数据           │  │   • 缓存数据           │    │
│  │   • 扫描结果           │  │   • 会话数据           │    │
│  │   • 用户信息           │  │   • 实时数据           │    │
│  │   • 配置数据           │  │   • 任务队列           │    │
│  └─────────────────────────┘  └─────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### 🔀 数据流架构

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   用户界面   │───▶│  API 网关   │───▶│  业务逻辑   │
│  (Frontend) │    │ (API Layer) │    │ (Services)  │
└─────────────┘    └─────────────┘    └─────────────┘
                                             │
                                             ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   扫描引擎   │───▶│  任务队列   │───▶│  结果处理   │
│  (Scanner)  │    │ (Task Queue)│    │ (Processor) │
└─────────────┘    └─────────────┘    └─────────────┘
                                             │
                                             ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   数据存储   │◀───│  缓存层     │◀───│  数据聚合   │
│  (MongoDB)  │    │  (Redis)    │    │ (Aggregator)│
└─────────────┘    └─────────────┘    └─────────────┘
```

---

## 🚀 快速开始

### 📋 环境要求

#### 系统要求
- **操作系统**: Windows 10+ / macOS 10.15+ / Linux (Ubuntu 18.04+)
- **CPU**: 4+ 核心 (推荐 8+ 核心)
- **内存**: 8GB+ (推荐 16GB+)
- **存储**: 100GB+ SSD

#### 软件依赖
- **Go**: 1.21+ (推荐 1.24.3)
- **Node.js**: 20+ (推荐 20.10+)
- **pnpm**: 8+ (推荐 8.10+)
- **MongoDB**: 6.0+
- **Redis**: 7.0+
- **Git**: 2.0+

### 🛠️ 安装步骤

#### 1. 克隆项目
```bash
git clone https://github.com/StellarServer/StellarServer.git
cd StellarServer
```

#### 2. 配置环境
```bash
# 复制配置文件
cp config.dev.yaml config.yaml

# 编辑配置文件，配置数据库连接
vi config.yaml
```

#### 3. 安装依赖
```bash
# 安装后端依赖
go mod tidy
go mod download

# 安装前端依赖
cd web
pnpm install
cd ..
```

#### 4. 启动服务

##### 方式一: 使用 Makefile (推荐)
```bash
# 检查系统依赖
make check-deps

# 安装项目依赖
make install-deps

# 启动开发环境
make dev
```

##### 方式二: 手动启动
```bash
# 启动后端服务
go run cmd/main.go -config config.yaml

# 启动前端服务 (新终端)
cd web
pnpm dev
```

### 📱 访问应用

启动成功后，可以通过以下地址访问：

- **前端界面**: http://localhost:5173
- **后端API**: http://localhost:8090
- **API 文档**: http://localhost:8090/api/v1/docs

### 🔧 开发工具

#### 推荐 IDE
- **Visual Studio Code**: 推荐插件
  - Go (官方)
  - Svelte for VS Code
  - TypeScript Importer
  - Tailwind CSS IntelliSense
  - GitLens

#### 有用的命令
```bash
# 查看项目状态
make status

# 查看实时日志
make logs

# 运行测试
make test

# 构建项目
make build

# 清理项目
make clean
```

---

## 📖 使用指南

### 🎯 核心功能使用

#### 1. 项目管理
1. **创建项目**: 登录后在项目页面创建新项目
2. **配置项目**: 设置项目的基本信息和扫描配置
3. **管理成员**: 添加项目成员并分配权限

#### 2. 资产发现
1. **子域名枚举**: 
   - 进入项目 → 子域名模块
   - 添加主域名
   - 配置扫描参数
   - 启动扫描任务

2. **端口扫描**:
   - 进入项目 → 端口扫描模块
   - 添加目标 IP 或 IP 段
   - 配置扫描端口和参数
   - 启动扫描任务

3. **Web 应用发现**:
   - 进入项目 → Web 应用模块
   - 添加目标 URL
   - 配置爬虫参数
   - 启动发现任务

#### 3. 漏洞扫描
1. **漏洞检测**:
   - 进入项目 → 漏洞扫描模块
   - 选择扫描目标
   - 配置扫描插件
   - 启动扫描任务

2. **查看结果**:
   - 实时查看扫描进度
   - 查看详细扫描报告
   - 导出扫描结果

#### 4. 监控告警
1. **配置告警**:
   - 进入设置 → 告警配置
   - 设置告警规则
   - 配置通知方式

2. **查看告警**:
   - 查看告警历史
   - 处理告警事件
   - 分析告警趋势

---

## 📦 部署指南

### 🐳 Docker 部署 (推荐)

#### 1. 准备 Docker 环境
```bash
# 安装 Docker 和 Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### 2. 部署应用
```bash
# 克隆项目
git clone https://github.com/StellarServer/StellarServer.git
cd StellarServer

# 构建镜像
docker build -t stellarserver:latest .

# 启动服务
docker-compose up -d
```

#### 3. 访问应用
```bash
# 查看服务状态
docker-compose ps

# 访问应用
# 前端: http://your-domain:80
# 后端: http://your-domain:8090
```

### 🖥️ 传统部署

#### 1. 准备服务器环境
```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装依赖
sudo apt install -y git curl wget

# 安装 Go
wget https://go.dev/dl/go1.24.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 安装 Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs
sudo npm install -g pnpm

# 安装 MongoDB
wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | sudo apt-key add -
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/6.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-6.0.list
sudo apt-get update
sudo apt-get install -y mongodb-org

# 安装 Redis
sudo apt install -y redis-server

# 安装 Nginx
sudo apt install -y nginx
```

#### 2. 部署应用
```bash
# 克隆项目
git clone https://github.com/StellarServer/StellarServer.git
cd StellarServer

# 构建后端
go build -o stellar cmd/main.go

# 构建前端
cd web
pnpm install
pnpm build
cd ..

# 配置系统服务
sudo cp stellar.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable stellar
sudo systemctl start stellar

# 配置 Nginx
sudo cp nginx.conf /etc/nginx/sites-available/stellar
sudo ln -s /etc/nginx/sites-available/stellar /etc/nginx/sites-enabled/
sudo systemctl reload nginx
```

### ⚙️ 配置说明

#### 环境变量
```bash
# 数据库配置
export MONGODB_URI="mongodb://localhost:27017"
export MONGODB_DATABASE="stellarserver"
export REDIS_ADDR="localhost:6379"

# 系统配置
export JWT_SECRET="your-secret-key"
export SERVER_HOST="0.0.0.0"
export SERVER_PORT="8090"
export LOG_LEVEL="info"

# 安全配置
export ENABLE_TLS="true"
export TLS_CERT_PATH="/path/to/cert.pem"
export TLS_KEY_PATH="/path/to/key.pem"
```

#### 配置文件
```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 8090
  mode: "release"

mongodb:
  uri: "mongodb://localhost:27017"
  database: "stellarserver"
  user: "admin"
  password: "password"

redis:
  addr: "localhost:6379"
  password: "password"
  db: 0

auth:
  jwtSecret: "your-secret-key"
  tokenExpiry: 24

# 扫描配置
subdomain:
  timeout: 10
  maxConcurrency: 100
  retryTimes: 3

portscan:
  timeout: 5
  rateLimit: 1000
  maxConcurrency: 100
```

---

## 🔌 API 文档

### 📋 API 概览

星络提供完整的 RESTful API，支持所有核心功能的程序化访问。

#### 基础信息
- **API 版本**: v1
- **基础路径**: `/api/v1`
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON

#### 统一响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 响应数据
  }
}
```

### 🔐 认证接口

#### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

#### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "newuser",
  "password": "password",
  "email": "user@example.com"
}
```

#### 刷新令牌
```http
POST /api/v1/auth/refresh
Authorization: Bearer <token>
```

### 📊 项目管理

#### 获取项目列表
```http
GET /api/v1/projects
Authorization: Bearer <token>
```

#### 创建项目
```http
POST /api/v1/projects
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "项目名称",
  "description": "项目描述",
  "domain": "example.com"
}
```

#### 更新项目
```http
PUT /api/v1/projects/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "更新的项目名称",
  "description": "更新的项目描述"
}
```

### 🎯 资产管理

#### 获取资产列表
```http
GET /api/v1/assets/assets?projectId={projectId}&type={type}&page=1&pageSize=20
Authorization: Bearer <token>
```

#### 创建资产
```http
POST /api/v1/assets/assets
Authorization: Bearer <token>
Content-Type: application/json

{
  "type": "subdomain",
  "projectId": "project-id",
  "data": {
    "host": "sub.example.com",
    "ips": ["1.2.3.4"],
    "cname": "example.com"
  }
}
```

#### 批量创建资产
```http
POST /api/v1/assets/batch
Authorization: Bearer <token>
Content-Type: application/json

{
  "type": "subdomain",
  "projectId": "project-id",
  "assets": [
    {
      "host": "sub1.example.com",
      "ips": ["1.2.3.4"]
    },
    {
      "host": "sub2.example.com",
      "ips": ["1.2.3.5"]
    }
  ]
}
```

### 🔍 扫描任务

#### 创建子域名扫描任务
```http
POST /api/v1/subdomain/tasks
Authorization: Bearer <token>
Content-Type: application/json

{
  "projectId": "project-id",
  "rootDomain": "example.com",
  "config": {
    "timeout": 30,
    "maxConcurrency": 50,
    "enableDnsResolution": true,
    "enableWildcardDetection": true
  }
}
```

#### 创建端口扫描任务
```http
POST /api/v1/portscan/tasks
Authorization: Bearer <token>
Content-Type: application/json

{
  "projectId": "project-id",
  "targets": ["192.168.1.0/24"],
  "config": {
    "ports": [80, 443, 8080, 8443],
    "timeout": 5,
    "maxConcurrency": 100
  }
}
```

#### 获取任务状态
```http
GET /api/v1/tasks/{taskId}
Authorization: Bearer <token>
```

### 📈 统计数据

#### 获取项目统计
```http
GET /api/v1/statistics/project/{projectId}
Authorization: Bearer <token>
```

#### 获取资产统计
```http
GET /api/v1/statistics/assets?projectId={projectId}
Authorization: Bearer <token>
```

---

## 🔧 插件开发

### 📖 插件系统概述

星络提供了强大的插件系统，支持多种插件类型和开发方式。

#### 插件类型
- **扫描类插件**: 用于各种扫描任务
- **信息收集插件**: 用于信息收集和分析
- **漏洞检测插件**: 用于漏洞检测和验证
- **工具类插件**: 用于辅助功能

#### 插件格式
- **YAML 插件**: 声明式插件定义
- **Go 插件**: 高性能编译型插件
- **Python 插件**: 灵活的脚本型插件

### 📝 YAML 插件开发

#### 基础结构
```yaml
# plugin.yaml
id: example_plugin
name: 示例插件
version: "1.0.0"
author: Your Name
description: 这是一个示例插件
type: scanner
category: subdomain

config:
  timeout: 30
  max_threads: 10
  enable_feature: true

script:
  language: python
  entry: main
  content: |
    import sys
    import json
    
    def main():
        # 插件逻辑
        result = {
            "status": "success",
            "data": []
        }
        print(json.dumps(result))
    
    if __name__ == "__main__":
        main()
```

#### 插件配置
```yaml
# 详细配置示例
dependencies:
  - requests
  - beautifulsoup4

tags:
  - subdomain
  - discovery
  - dns

input:
  - type: string
    name: domain
    description: 目标域名
    required: true

output:
  - type: array
    name: subdomains
    description: 发现的子域名列表
```

### 🐍 Python 插件开发

#### 插件模板
```python
# plugin.py
import sys
import json
import requests
from typing import List, Dict, Any

class SubdomainScanner:
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.timeout = config.get('timeout', 30)
        self.max_threads = config.get('max_threads', 10)
    
    def scan(self, domain: str) -> List[str]:
        """扫描子域名"""
        subdomains = []
        
        # 实现扫描逻辑
        # ...
        
        return subdomains
    
    def run(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """插件入口函数"""
        try:
            domain = params.get('domain')
            if not domain:
                return {
                    "status": "error",
                    "message": "域名参数不能为空"
                }
            
            subdomains = self.scan(domain)
            
            return {
                "status": "success",
                "data": {
                    "subdomains": subdomains,
                    "count": len(subdomains)
                }
            }
        except Exception as e:
            return {
                "status": "error",
                "message": str(e)
            }

def main():
    # 读取输入参数
    if len(sys.argv) < 2:
        print(json.dumps({"status": "error", "message": "缺少参数"}))
        sys.exit(1)
    
    params = json.loads(sys.argv[1])
    config = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}
    
    # 创建插件实例
    scanner = SubdomainScanner(config)
    
    # 执行扫描
    result = scanner.run(params)
    
    # 输出结果
    print(json.dumps(result))

if __name__ == "__main__":
    main()
```

### 🔧 Go 插件开发

#### 插件接口
```go
// plugin.go
package main

import (
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/StellarServer/internal/plugin/sdk"
)

type SubdomainPlugin struct {
    config sdk.PluginConfig
}

func (p *SubdomainPlugin) Init(config sdk.PluginConfig) error {
    p.config = config
    return nil
}

func (p *SubdomainPlugin) GetInfo() sdk.PluginInfo {
    return sdk.PluginInfo{
        ID:          "go_subdomain_scanner",
        Name:        "Go 子域名扫描器",
        Version:     "1.0.0",
        Author:      "Your Name",
        Description: "高性能的子域名扫描插件",
        Type:        "scanner",
        Category:    "subdomain",
    }
}

func (p *SubdomainPlugin) Execute(params map[string]interface{}) (map[string]interface{}, error) {
    domain, ok := params["domain"].(string)
    if !ok {
        return nil, fmt.Errorf("域名参数无效")
    }
    
    // 实现扫描逻辑
    subdomains, err := p.scanSubdomains(domain)
    if err != nil {
        return nil, err
    }
    
    return map[string]interface{}{
        "status": "success",
        "data": map[string]interface{}{
            "subdomains": subdomains,
            "count":      len(subdomains),
        },
    }, nil
}

func (p *SubdomainPlugin) scanSubdomains(domain string) ([]string, error) {
    // 实现具体的扫描逻辑
    var subdomains []string
    
    // 示例：DNS 查询
    // ...
    
    return subdomains, nil
}

func (p *SubdomainPlugin) Stop() error {
    // 清理资源
    return nil
}

// 插件导出函数
func NewPlugin() sdk.Plugin {
    return &SubdomainPlugin{}
}
```

---

## 🛠️ 开发指南

### 📋 开发环境设置

#### 1. 开发工具配置
```bash
# 安装开发依赖
go install -a github.com/cosmtrek/air@latest
go install github.com/swaggo/swag/cmd/swag@latest

# 配置开发环境
git config --global core.autocrlf true
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

#### 2. VS Code 配置
```json
// .vscode/settings.json
{
  "go.useLanguageServer": true,
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.testFlags": ["-v", "-race"],
  "svelte.enable-ts-plugin": true,
  "typescript.preferences.importModuleSpecifier": "relative",
  "tailwindCSS.includeLanguages": {
    "svelte": "html"
  }
}
```

#### 3. 推荐插件
- **Go**: 官方 Go 语言支持
- **Svelte for VS Code**: Svelte 语言支持
- **TypeScript Importer**: 自动导入 TypeScript 模块
- **Tailwind CSS IntelliSense**: Tailwind CSS 智能提示
- **GitLens**: Git 增强工具

### 🏗️ 项目结构

#### 后端项目结构
```
Stellar/
├── cmd/                    # 应用入口
│   ├── main.go            # 主程序入口
│   ├── web_dev.go         # 开发环境配置
│   └── web_prod.go        # 生产环境配置
├── internal/              # 内部包
│   ├── api/               # HTTP API 处理器
│   ├── config/            # 配置管理
│   ├── database/          # 数据库连接
│   ├── models/            # 数据模型
│   ├── services/          # 业务逻辑服务
│   └── utils/             # 工具函数
├── plugins/               # 插件目录
├── scripts/               # 脚本文件
├── config.yaml            # 配置文件
└── go.mod                 # Go 模块文件
```

#### 前端项目结构
```
web/
├── src/
│   ├── lib/               # 组件库和工具
│   │   ├── components/    # 可复用组件
│   │   ├── stores/        # 状态管理
│   │   ├── utils/         # 工具函数
│   │   └── api/           # API 客户端
│   ├── routes/            # SvelteKit 路由
│   │   ├── (app)/         # 应用页面组
│   │   ├── (auth)/        # 认证页面组
│   │   └── +layout.svelte # 根布局
│   └── app.html           # HTML 模板
├── static/                # 静态资源
├── package.json           # 依赖配置
└── svelte.config.js       # Svelte 配置
```

### 📝 开发规范

#### 代码风格
```go
// 后端代码风格示例
package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/StellarServer/internal/models"
)

// AssetHandler 资产处理器
type AssetHandler struct {
    db *mongo.Database
}

// CreateAsset 创建资产
func (h *AssetHandler) CreateAsset(c *gin.Context) {
    var req models.CreateAssetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    400,
            "message": "参数错误",
            "details": err.Error(),
        })
        return
    }
    
    // 处理逻辑
    asset, err := h.createAsset(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    500,
            "message": "创建失败",
            "details": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "code":    200,
        "message": "创建成功",
        "data":    asset,
    })
}
```

```typescript
// 前端代码风格示例
// src/lib/api/asset.ts
import api from './axios-config';
import type { Asset, CreateAssetRequest } from '$lib/types/asset';

export const assetApi = {
  async createAsset(data: CreateAssetRequest): Promise<Asset> {
    const response = await api.post<{ code: number; message: string; data: Asset }>(
      '/assets/assets',
      data
    );
    
    if (response.data.code !== 200) {
      throw new Error(response.data.message);
    }
    
    return response.data.data;
  },

  async getAssets(params?: AssetQueryParams): Promise<AssetListResult> {
    const response = await api.get<{ code: number; message: string; data: AssetListResult }>(
      '/assets/assets',
      { params }
    );
    
    if (response.data.code !== 200) {
      throw new Error(response.data.message);
    }
    
    return response.data.data;
  }
};
```

#### 提交规范
```bash
# 提交消息格式
<type>(<scope>): <subject>

# 类型说明
feat:     新功能
fix:      修复问题
docs:     文档更新
style:    代码格式（不影响功能的更改）
refactor: 重构代码
test:     测试相关
chore:    构建过程或辅助工具的变动

# 示例
feat(asset): 添加批量创建资产功能
fix(auth): 修复JWT令牌刷新问题
docs(readme): 更新安装说明
```

### 🧪 测试指南

#### 后端测试
```go
// internal/api/asset_test.go
package api

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestCreateAsset(t *testing.T) {
    // 准备测试数据
    req := models.CreateAssetRequest{
        Type:      "subdomain",
        ProjectID: "test-project-id",
        Data: map[string]interface{}{
            "host": "test.example.com",
            "ips":  []string{"1.2.3.4"},
        },
    }
    
    // 执行测试
    asset, err := handler.createAsset(&req)
    
    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, asset)
    assert.Equal(t, "test.example.com", asset.Host)
}
```

#### 前端测试
```typescript
// src/lib/api/__tests__/asset.test.ts
import { describe, it, expect, vi } from 'vitest';
import { assetApi } from '../asset';

describe('Asset API', () => {
  it('should create asset successfully', async () => {
    // Mock API 响应
    const mockAsset = {
      id: 'test-id',
      type: 'subdomain',
      host: 'test.example.com'
    };
    
    vi.spyOn(api, 'post').mockResolvedValue({
      data: {
        code: 200,
        message: 'success',
        data: mockAsset
      }
    });
    
    // 执行测试
    const result = await assetApi.createAsset({
      type: 'subdomain',
      projectId: 'test-project',
      data: { host: 'test.example.com' }
    });
    
    // 验证结果
    expect(result).toEqual(mockAsset);
  });
});
```

#### 运行测试
```bash
# 后端测试
go test -v ./internal/...

# 前端测试
cd web
pnpm test

# E2E 测试
cd web
pnpm test:e2e
```

---

## 🤝 贡献指南

### 🎯 贡献方式

我们欢迎以下类型的贡献：

- **🐛 Bug 修复**: 发现并修复项目中的问题
- **✨ 新功能**: 添加新的功能或改进现有功能
- **📚 文档完善**: 改进文档、教程和示例
- **🔧 插件开发**: 开发新的扫描插件或工具
- **🧪 测试用例**: 添加测试用例和改进测试覆盖率
- **💡 建议反馈**: 提出改进建议或功能请求

### 📋 贡献流程

#### 1. 准备工作
```bash
# Fork 项目到您的 GitHub 账户
# 克隆您的 Fork
git clone https://github.com/YOUR_USERNAME/StellarServer.git
cd StellarServer

# 添加上游仓库
git remote add upstream https://github.com/StellarServer/StellarServer.git

# 创建新分支
git checkout -b feature/your-feature-name
```

#### 2. 开发工作
```bash
# 保持代码更新
git fetch upstream
git rebase upstream/main

# 进行开发
# ... 编写代码 ...

# 运行测试
make test

# 代码格式化
make format
```

#### 3. 提交更改
```bash
# 添加文件
git add .

# 提交更改
git commit -m "feat(scope): add new feature"

# 推送到您的 Fork
git push origin feature/your-feature-name
```

#### 4. 创建 Pull Request
1. 访问 GitHub 上的项目页面
2. 点击 "Compare & pull request"
3. 填写 PR 描述，包括：
   - 更改的内容
   - 相关的 Issue
   - 测试说明
   - 截图（如果适用）

### 📝 开发规范

#### 代码质量要求
- **测试覆盖率**: 新功能需要包含单元测试
- **文档完善**: 重要功能需要更新文档
- **性能考虑**: 避免引入性能问题
- **安全检查**: 确保没有安全漏洞

#### 提交规范
```bash
# 提交消息格式
<type>(<scope>): <subject>

<body>

<footer>
```

### 🏆 贡献者

感谢所有为项目做出贡献的开发者！

<a href="https://github.com/StellarServer/StellarServer/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=StellarServer/StellarServer" />
</a>

---

## 📄 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

```
MIT License

Copyright (c) 2024 StellarServer

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

## 🔗 相关链接

- **项目主页**: https://github.com/StellarServer/StellarServer
- **在线文档**: https://docs.stellarserver.com
- **问题反馈**: https://github.com/StellarServer/StellarServer/issues
- **讨论区**: https://github.com/StellarServer/StellarServer/discussions
- **更新日志**: https://github.com/StellarServer/StellarServer/releases

---

## 📞 联系我们

如果您有任何问题或建议，请通过以下方式联系我们：

- **GitHub Issues**: [提交问题](https://github.com/StellarServer/StellarServer/issues)
- **GitHub Discussions**: [参与讨论](https://github.com/StellarServer/StellarServer/discussions)
- **邮件**: stellar-dev@example.com
- **QQ群**: 123456789

---

<div align="center">

**⭐ 如果这个项目对您有帮助，请给我们一个星标！**

**Star History**

[![Star History Chart](https://api.star-history.com/svg?repos=StellarServer/StellarServer&type=Date)](https://star-history.com/#StellarServer/StellarServer&Date)

</div>

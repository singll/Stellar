# YAML插件格式规范

Stellar 平台支持使用 YAML 格式定义插件，这种方式允许开发者快速创建功能丰富的插件而无需编译。

## 插件结构

### 基本信息

```yaml
id: plugin_unique_id           # 插件唯一标识符
name: 插件显示名称              # 插件显示名称
version: "1.0.0"              # 插件版本号
author: 作者名称               # 插件作者
description: 插件功能描述       # 详细的功能描述
type: scanner                 # 插件类型
category: subdomain           # 插件分类
```

### 插件类型 (type)

- `scanner` - 扫描类插件
- `info_gatherer` - 信息收集插件
- `vulnerability` - 漏洞检测插件
- `utility` - 工具类插件
- `custom` - 自定义插件

### 插件分类 (category)

- `subdomain` - 子域名发现
- `port` - 端口扫描
- `web` - Web应用
- `network` - 网络扫描
- `osint` - 开源情报
- `misc` - 其他

### 标签和依赖

```yaml
tags:                         # 标签列表
  - subdomain
  - discovery
  - dns

dependencies:                 # 依赖列表
  - nmap                      # 系统工具依赖
  - python3                   # 运行时依赖
```

### 配置参数

```yaml
config:                       # 插件配置
  timeout: 30                 # 超时时间
  max_threads: 10             # 最大线程数
  enable_feature: true        # 功能开关
  wordlist_path: "/path/to/wordlist.txt"  # 文件路径
```

### 脚本配置

```yaml
script:
  language: python            # 脚本语言
  entry: main                 # 入口函数名
  args: []                    # 参数列表
  content: |                  # 脚本内容
    # 这里是具体的脚本代码
```

## 支持的脚本语言

### Python

```yaml
script:
  language: python
  entry: main
  content: |
    import json
    import sys
    
    def main(params):
        # 插件逻辑
        domain = params.get('domain', '')
        return {
            'success': True,
            'result': f'处理域名: {domain}'
        }
    
    if __name__ == '__main__':
        input_data = sys.stdin.read()
        params = json.loads(input_data) if input_data else {}
        result = main(params)
        print(json.dumps(result, ensure_ascii=False))
```

### JavaScript (Node.js)

```yaml
script:
  language: javascript
  entry: main
  content: |
    function main(params) {
        return new Promise((resolve) => {
            const url = params.url || '';
            resolve({
                success: true,
                result: `处理URL: ${url}`
            });
        });
    }
    
    if (require.main === module) {
        let inputData = '';
        process.stdin.on('data', (chunk) => { inputData += chunk; });
        process.stdin.on('end', async () => {
            const params = inputData ? JSON.parse(inputData) : {};
            const result = await main(params);
            console.log(JSON.stringify(result, null, 2));
        });
    }
```

### Shell脚本

```yaml
script:
  language: shell
  entry: main
  content: |
    #!/bin/bash
    read -r input_data
    
    # 解析JSON参数
    target=$(echo "$input_data" | jq -r '.target // "localhost"')
    
    # 执行操作
    result=$(nmap -p 80,443 "$target" 2>/dev/null)
    
    # 输出JSON结果
    cat << EOF
    {
        "success": true,
        "target": "$target",
        "result": "$(echo "$result" | sed 's/"/\\"/g')"
    }
    EOF
```

### Lua

```yaml
script:
  language: lua
  entry: main
  content: |
    local json = require("json")
    
    function main(params)
        local domain = params.domain or ""
        return {
            success = true,
            result = "处理域名: " .. domain
        }
    end
    
    -- 读取输入
    local input = io.read("*all")
    local params = input and json.decode(input) or {}
    
    -- 执行并输出结果
    local result = main(params)
    print(json.encode(result))
```

## 参数传递

插件通过标准输入接收JSON格式的参数：

```json
{
    "domain": "example.com",
    "config": {
        "timeout": 30,
        "threads": 5
    },
    "options": {
        "deep_scan": true
    }
}
```

## 结果返回

插件必须通过标准输出返回JSON格式的结果：

```json
{
    "success": true,
    "data": {
        "subdomains": ["www.example.com", "api.example.com"],
        "count": 2
    },
    "error": null
}
```

### 成功响应格式

```json
{
    "success": true,
    "data": {
        // 具体的结果数据
    },
    "metadata": {
        "execution_time": 1.5,
        "plugin_version": "1.0.0"
    }
}
```

### 错误响应格式

```json
{
    "success": false,
    "error": "错误描述信息",
    "error_code": "ERROR_CODE",
    "data": null
}
```

## 最佳实践

### 1. 错误处理

```python
def main(params):
    try:
        # 插件逻辑
        result = perform_operation(params)
        return {'success': True, 'data': result}
    except Exception as e:
        return {'success': False, 'error': str(e)}
```

### 2. 参数验证

```python
def main(params):
    # 验证必需参数
    domain = params.get('domain')
    if not domain:
        return {'success': False, 'error': '缺少domain参数'}
    
    # 验证参数格式
    if not isinstance(domain, str) or '.' not in domain:
        return {'success': False, 'error': 'domain格式无效'}
```

### 3. 配置使用

```python
def main(params):
    # 获取配置
    config = params.get('config', {})
    timeout = config.get('timeout', 30)
    max_threads = config.get('max_threads', 5)
    
    # 使用配置执行操作
```

### 4. 进度报告

对于长时间运行的插件，可以通过标准错误输出报告进度：

```python
import sys

def report_progress(message):
    print(f"PROGRESS: {message}", file=sys.stderr)

def main(params):
    report_progress("开始扫描...")
    # 执行操作
    report_progress("扫描完成")
```

## 调试和测试

### 本地测试

```bash
# 创建测试参数文件
echo '{"domain": "example.com"}' > test_params.json

# 测试Python插件
cat test_params.json | python3 plugin_script.py

# 测试Shell插件
cat test_params.json | bash plugin_script.sh
```

### 日志记录

```python
import logging

# 配置日志
logging.basicConfig(level=logging.INFO, 
                   format='%(asctime)s - %(levelname)s - %(message)s')

def main(params):
    logging.info("插件开始执行")
    # 插件逻辑
    logging.info("插件执行完成")
```

## 示例插件

参考 `/plugins/examples/` 目录中的示例插件：

- `subdomain_hunter.yaml` - Python子域名发现插件
- `simple_port_scanner.yaml` - Shell端口扫描插件  
- `web_title_fetcher.yaml` - JavaScript网站信息获取插件

这些示例展示了不同语言和用途的YAML插件实现方式。
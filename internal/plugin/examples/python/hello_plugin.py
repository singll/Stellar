#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import sys
import time
from datetime import datetime

# 插件信息
PLUGIN_INFO = {
    "id": "hello_plugin_python",
    "name": "Hello Plugin (Python)",
    "version": "1.0.0",
    "type": "utility",
    "author": "StellarServer Team",
    "description": "A simple example plugin written in Python",
    "category": "example",
    "tags": ["example", "hello", "utility", "python"],
    "created_at": datetime.now().isoformat(),
    "updated_at": datetime.now().isoformat(),
    "params": [
        {
            "name": "name",
            "type": "string",
            "description": "Your name",
            "required": True,
            "default": "World"
        },
        {
            "name": "greeting",
            "type": "string",
            "description": "Greeting message",
            "required": False,
            "default": "Hello",
            "options": ["Hello", "Hi", "Hey", "Greetings"]
        }
    ],
    "language": "Python"
}

# 全局配置
CONFIG = {}

def info():
    """返回插件信息"""
    return PLUGIN_INFO

def init(config_json):
    """初始化插件"""
    global CONFIG
    CONFIG = json.loads(config_json)
    return {"success": True}

def execute(params_json):
    """执行插件"""
    # 解析参数
    params = json.loads(params_json)
    
    # 获取参数值
    name = params.get("name", "World")
    greeting = params.get("greeting", "Hello")
    
    # 构建消息
    message = f"{greeting}, {name}!"
    
    # 返回结果
    result = {
        "success": True,
        "data": message,
        "message": "Plugin executed successfully",
        "metadata": {
            "timestamp": datetime.now().isoformat()
        }
    }
    
    return json.dumps(result)

def validate(params_json):
    """验证参数"""
    params = json.loads(params_json)
    
    # 检查必需参数
    if "name" not in params:
        return json.dumps({
            "success": False,
            "error": "Missing required parameter: name"
        })
    
    return json.dumps({"success": True})

def cleanup():
    """清理资源"""
    return json.dumps({"success": True})

if __name__ == "__main__":
    # 命令行接口
    if len(sys.argv) < 2:
        print("Usage: hello_plugin.py <command> [args]")
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
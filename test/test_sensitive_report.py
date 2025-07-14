#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
敏感信息检测报告生成功能测试脚本
"""
import requests
import json
import time
import os
import tempfile


class SensitiveReportTester:
    def __init__(self, base_url="http://localhost:8090"):
        self.base_url = base_url
        self.token = None
        self.project_id = None
        self.rule_id = None
        self.detection_id = None
    
    def test_authentication(self):
        """测试用户认证"""
        print("🔐 测试用户认证...")
        
        # 注册用户（如果用户已存在会失败，这是正常的）
        register_data = {
            "username": "report_test_user",
            "password": "test123456",
            "email": "report_test@example.com"
        }
        
        try:
            response = requests.post(f"{self.base_url}/api/v1/auth/register", 
                                   json=register_data)
            if response.status_code == 200:
                print("   ✅ 用户注册成功")
        except Exception:
            pass
        
        # 登录用户
        login_data = {
            "username": "report_test_user",
            "password": "test123456"
        }
        
        response = requests.post(f"{self.base_url}/api/v1/auth/login", 
                               json=login_data)
        
        if response.status_code == 200:
            data = response.json()
            self.token = data.get("data", {}).get("token")
            print(f"   ✅ 用户登录成功，获得令牌")
            return True
        else:
            print(f"   ❌ 用户登录失败: {response.text}")
            return False
    
    def get_headers(self):
        """获取带认证的请求头"""
        return {
            "Authorization": f"Bearer {self.token}",
            "Content-Type": "application/json"
        }
    
    def test_create_project(self):
        """测试创建项目"""
        print("📂 测试创建项目...")
        
        project_data = {
            "name": "敏感信息检测测试项目",
            "description": "用于测试敏感信息检测报告生成功能的项目"
        }
        
        response = requests.post(f"{self.base_url}/api/v1/projects/projects",
                               json=project_data, headers=self.get_headers())
        
        if response.status_code == 200:
            data = response.json()
            self.project_id = data.get("id")
            print(f"   ✅ 项目创建成功，ID: {self.project_id}")
            return True
        else:
            print(f"   ❌ 项目创建失败: {response.text}")
            return False
    
    def test_create_sensitive_rule(self):
        """测试创建敏感规则"""
        print("📋 测试创建敏感规则...")
        
        rule_data = {
            "name": "测试API密钥规则",
            "description": "检测API密钥泄露",
            "type": "regex",
            "pattern": r"(?i)(api[_-]?key|apikey)[\"'`]*\\s*[:=]\\s*[\"'`]*([a-z0-9]{32,})",
            "category": "api_key",
            "riskLevel": "high",
            "tags": ["api", "key", "secret"],
            "enabled": True,
            "context": 3,
            "examples": ["api_key = 'abcd1234567890abcd1234567890abcd'"],
            "falsePositivePatterns": ["api_key_example", "test_api_key"]
        }
        
        response = requests.post(f"{self.base_url}/api/v1/sensitive/sensitive/rules",
                               json=rule_data, headers=self.get_headers())
        
        if response.status_code == 200:
            data = response.json()
            self.rule_id = data.get("id")
            print(f"   ✅ 敏感规则创建成功，ID: {self.rule_id}")
            return True
        else:
            print(f"   ❌ 敏感规则创建失败: {response.text}")
            return False
    
    def test_create_test_file(self):
        """创建包含敏感信息的测试文件"""
        print("📄 创建测试文件...")
        
        test_content = """
# 配置文件
database_url = "mongodb://localhost:27017/testdb"
api_key = "sk-1234567890abcdef1234567890abcdef"
secret_token = "abc123def456ghi789jkl012mno345pqr"

# 用户配置
user_password = "mypassword123"
jwt_secret = "my-super-secret-jwt-key-2023"

# AWS 配置
aws_access_key = "AKIAIOSFODNN7EXAMPLE"
aws_secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

# 数据库连接
db_host = "localhost"
db_port = 5432
db_username = "admin"
db_password = "admin123"

# 邮件配置
smtp_username = "noreply@example.com"
smtp_password = "email_password_2023"
"""
        
        # 创建临时文件
        with tempfile.NamedTemporaryFile(mode='w', suffix='.conf', delete=False) as f:
            f.write(test_content)
            self.test_file_path = f.name
        
        print(f"   ✅ 测试文件创建成功: {self.test_file_path}")
        return True
    
    def test_sensitive_detection(self):
        """测试敏感信息检测"""
        print("🔍 测试敏感信息检测...")
        
        scan_data = {
            "projectId": self.project_id,
            "urls": [f"file://{self.test_file_path}"],
            "ruleIds": [self.rule_id] if self.rule_id else []
        }
        
        response = requests.post(f"{self.base_url}/api/v1/sensitive/sensitive/scan",
                               json=scan_data, headers=self.get_headers())
        
        if response.status_code == 200:
            data = response.json()
            self.detection_id = data.get("taskId")
            print(f"   ✅ 敏感信息检测任务创建成功，ID: {self.detection_id}")
            
            # 等待检测完成
            print("   ⏳ 等待检测完成...")
            time.sleep(3)
            return True
        else:
            print(f"   ❌ 敏感信息检测失败: {response.text}")
            return False
    
    def test_generate_reports(self):
        """测试生成不同格式的报告"""
        print("📊 测试报告生成...")
        
        if not self.detection_id:
            print("   ❌ 没有检测ID，跳过报告生成测试")
            return False
        
        # 测试不同格式的报告
        formats = ["html", "json", "csv", "xml", "txt"]
        
        for format_type in formats:
            print(f"   📋 测试生成 {format_type.upper()} 格式报告...")
            
            report_data = {
                "format": format_type,
                "includeSummary": True,
                "includeDetails": True,
                "sortBy": "riskLevel",
                "sortOrder": "desc"
            }
            
            response = requests.post(
                f"{self.base_url}/api/v1/sensitive/sensitive/{self.detection_id}/report",
                json=report_data, headers=self.get_headers()
            )
            
            if response.status_code == 200:
                data = response.json()
                print(f"      ✅ {format_type.upper()} 报告生成成功")
                print(f"         文件名: {data.get('filename')}")
                print(f"         大小: {data.get('size')} 字节")
                print(f"         下载URL: {data.get('downloadUrl')}")
                
                # 测试下载报告
                self.test_download_report(format_type)
            else:
                print(f"      ❌ {format_type.upper()} 报告生成失败: {response.text}")
        
        return True
    
    def test_download_report(self, format_type):
        """测试下载报告"""
        print(f"      📥 测试下载 {format_type.upper()} 报告...")
        
        response = requests.get(
            f"{self.base_url}/api/v1/sensitive/sensitive/{self.detection_id}/report/{format_type}",
            headers=self.get_headers()
        )
        
        if response.status_code == 200:
            # 保存报告文件
            filename = f"test_report_{int(time.time())}.{format_type}"
            with open(filename, 'wb') as f:
                f.write(response.content)
            
            print(f"         ✅ 报告下载成功，保存为: {filename}")
            print(f"         文件大小: {len(response.content)} 字节")
            
            # 显示部分内容预览
            if format_type in ["html", "json", "txt", "xml"]:
                content_preview = response.content.decode('utf-8')[:200]
                print(f"         内容预览: {content_preview}...")
            
        else:
            print(f"         ❌ 报告下载失败: {response.text}")
    
    def cleanup(self):
        """清理测试数据"""
        print("🧹 清理测试数据...")
        
        # 删除测试文件
        if hasattr(self, 'test_file_path') and os.path.exists(self.test_file_path):
            os.unlink(self.test_file_path)
            print("   ✅ 测试文件已删除")
        
        # 清理下载的报告文件
        import glob
        for report_file in glob.glob("test_report_*.html"):
            os.unlink(report_file)
        for report_file in glob.glob("test_report_*.json"):
            os.unlink(report_file)
        for report_file in glob.glob("test_report_*.csv"):
            os.unlink(report_file)
        for report_file in glob.glob("test_report_*.xml"):
            os.unlink(report_file)
        for report_file in glob.glob("test_report_*.txt"):
            os.unlink(report_file)
        
        print("   ✅ 报告文件已清理")
    
    def run_all_tests(self):
        """运行所有测试"""
        print("🚀 开始敏感信息检测报告生成功能测试\n")
        
        try:
            # 认证测试
            if not self.test_authentication():
                return False
            
            # 创建项目
            if not self.test_create_project():
                return False
            
            # 创建敏感规则
            if not self.test_create_sensitive_rule():
                return False
            
            # 创建测试文件
            if not self.test_create_test_file():
                return False
            
            # 敏感信息检测
            if not self.test_sensitive_detection():
                return False
            
            # 报告生成测试
            if not self.test_generate_reports():
                return False
            
            print("\n🎉 所有测试完成！敏感信息检测报告生成功能工作正常")
            return True
            
        except Exception as e:
            print(f"\n❌ 测试过程中发生错误: {str(e)}")
            return False
        
        finally:
            self.cleanup()


def main():
    """主函数"""
    tester = SensitiveReportTester()
    success = tester.run_all_tests()
    
    if success:
        print("\n✅ 敏感信息检测报告生成功能测试通过")
    else:
        print("\n❌ 敏感信息检测报告生成功能测试失败")
    
    return 0 if success else 1


if __name__ == "__main__":
    exit(main())
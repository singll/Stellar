#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
敏感信息检测报告生成功能演示脚本
直接测试报告生成器的功能，不依赖API交互
"""
import json
import time
import os
from datetime import datetime


def create_mock_detection_result():
    """创建模拟的检测结果"""
    return {
        "id": "65a1234567890abcdef12345",
        "projectId": "65a9876543210fedcba09876",
        "name": "敏感信息检测测试任务",
        "targets": [
            "file:///tmp/test_config.conf",
            "https://example.com/api/config"
        ],
        "status": "completed",
        "startTime": datetime.now().isoformat(),
        "endTime": datetime.now().isoformat(),
        "progress": 100.0,
        "findings": [
            {
                "id": "finding1",
                "target": "file:///tmp/test_config.conf",
                "targetType": "file",
                "rule": "api_key_rule",
                "ruleName": "API密钥检测",
                "category": "api_key",
                "riskLevel": "high",
                "pattern": r"api[_-]?key[\"'`]*\s*[:=]\s*[\"'`]*([a-z0-9]{32,})",
                "matchedText": "api_key = 'sk-1234567890abcdef1234567890abcdef'",
                "context": "# 配置文件\ndatabase_url = 'mongodb://localhost:27017/testdb'\napi_key = 'sk-1234567890abcdef1234567890abcdef'\nsecret_token = 'abc123def456ghi789jkl012mno345pqr'",
                "lineNumber": 3,
                "filePath": "/tmp/test_config.conf",
                "fileSize": 1024,
                "createdAt": datetime.now().isoformat()
            },
            {
                "id": "finding2",
                "target": "file:///tmp/test_config.conf",
                "targetType": "file",
                "rule": "password_rule",
                "ruleName": "密码检测",
                "category": "password",
                "riskLevel": "medium",
                "pattern": r"password\s*[:=]\s*[\"'`]*([^\"'`\s]+)",
                "matchedText": "user_password = 'mypassword123'",
                "context": "# 用户配置\nuser_name = 'admin'\nuser_password = 'mypassword123'\nuser_email = 'admin@example.com'",
                "lineNumber": 7,
                "filePath": "/tmp/test_config.conf",
                "fileSize": 1024,
                "createdAt": datetime.now().isoformat()
            },
            {
                "id": "finding3",
                "target": "https://example.com/api/config",
                "targetType": "url",
                "rule": "jwt_secret_rule",
                "ruleName": "JWT密钥检测",
                "category": "jwt_secret",
                "riskLevel": "high",
                "pattern": r"jwt[_-]?secret[\"'`]*\s*[:=]\s*[\"'`]*([^\"'`\s]+)",
                "matchedText": "jwt_secret: 'my-super-secret-jwt-key-2023'",
                "context": "{\n  \"database\": \"mongodb://localhost:27017\",\n  \"jwt_secret\": \"my-super-secret-jwt-key-2023\",\n  \"debug\": true\n}",
                "lineNumber": 3,
                "filePath": "",
                "fileSize": 0,
                "createdAt": datetime.now().isoformat()
            }
        ],
        "summary": {
            "totalFindings": 3,
            "riskLevelCount": {
                "high": 2,
                "medium": 1,
                "low": 0
            },
            "categoryCount": {
                "api_key": 1,
                "password": 1,
                "jwt_secret": 1
            }
        },
        "createdAt": datetime.now().isoformat(),
        "updatedAt": datetime.now().isoformat(),
        "totalCount": 2,
        "finishCount": 2
    }


def generate_html_report(detection_result, filename="sensitive_report_demo.html"):
    """生成HTML报告"""
    html_content = f"""<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>敏感信息检测报告 - {detection_result['name']}</title>
    <style>
        body {{ font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }}
        .container {{ max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }}
        .header {{ border-bottom: 2px solid #3498db; padding-bottom: 20px; margin-bottom: 30px; }}
        .title {{ color: #2c3e50; margin: 0; }}
        .subtitle {{ color: #7f8c8d; margin: 10px 0 0 0; }}
        .section {{ margin-bottom: 30px; }}
        .section-title {{ color: #2c3e50; border-left: 4px solid #3498db; padding-left: 10px; margin-bottom: 15px; }}
        .info-grid {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 15px; margin-bottom: 20px; }}
        .info-item {{ background: #f8f9fa; padding: 15px; border-radius: 4px; border-left: 3px solid #3498db; }}
        .info-label {{ font-weight: bold; color: #34495e; }}
        .info-value {{ color: #2c3e50; margin-top: 5px; }}
        .stats-grid {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; }}
        .stat-card {{ background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; border-radius: 8px; text-align: center; }}
        .stat-number {{ font-size: 2em; font-weight: bold; margin-bottom: 5px; }}
        .stat-label {{ font-size: 0.9em; opacity: 0.9; }}
        .risk-high {{ background: linear-gradient(135deg, #FF6B6B 0%, #EE5A24 100%); }}
        .risk-medium {{ background: linear-gradient(135deg, #FFA726 0%, #FB8C00 100%); }}
        .risk-low {{ background: linear-gradient(135deg, #66BB6A 0%, #43A047 100%); }}
        .findings-table {{ width: 100%; border-collapse: collapse; margin-top: 15px; }}
        .findings-table th, .findings-table td {{ padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }}
        .findings-table th {{ background-color: #3498db; color: white; font-weight: bold; }}
        .findings-table tr:nth-child(even) {{ background-color: #f2f2f2; }}
        .risk-badge {{ padding: 4px 8px; border-radius: 12px; font-size: 0.8em; font-weight: bold; color: white; }}
        .risk-badge.high {{ background-color: #e74c3c; }}
        .risk-badge.medium {{ background-color: #f39c12; }}
        .risk-badge.low {{ background-color: #27ae60; }}
        .matched-text {{ font-family: monospace; background: #f8f9fa; padding: 2px 4px; border-radius: 3px; }}
        .footer {{ text-align: center; margin-top: 40px; padding-top: 20px; border-top: 1px solid #ddd; color: #7f8c8d; }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="title">敏感信息检测报告</h1>
            <p class="subtitle">检测名称: {detection_result['name']} | 生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>
        </div>

        <div class="section">
            <h2 class="section-title">检测信息</h2>
            <div class="info-grid">
                <div class="info-item">
                    <div class="info-label">项目ID</div>
                    <div class="info-value">{detection_result['projectId']}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">检测状态</div>
                    <div class="info-value">{detection_result['status']}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">目标数量</div>
                    <div class="info-value">{len(detection_result['targets'])}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">完成进度</div>
                    <div class="info-value">{detection_result['progress']}%</div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">统计摘要</h2>
            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-number">{detection_result['summary']['totalFindings']}</div>
                    <div class="stat-label">总发现数</div>
                </div>"""
    
    # 添加风险等级统计
    for level, count in detection_result['summary']['riskLevelCount'].items():
        if count > 0:
            html_content += f"""
                <div class="stat-card risk-{level}">
                    <div class="stat-number">{count}</div>
                    <div class="stat-label">{level} 风险</div>
                </div>"""
    
    html_content += """
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">详细发现</h2>
            <table class="findings-table">
                <thead>
                    <tr>
                        <th>序号</th>
                        <th>目标</th>
                        <th>规则</th>
                        <th>风险等级</th>
                        <th>分类</th>
                        <th>匹配文本</th>
                        <th>行号</th>
                    </tr>
                </thead>
                <tbody>"""
    
    # 添加发现详情
    for i, finding in enumerate(detection_result['findings'], 1):
        matched_text = finding['matchedText'][:50] + "..." if len(finding['matchedText']) > 50 else finding['matchedText']
        line_number = finding.get('lineNumber', '-')
        
        html_content += f"""
                    <tr>
                        <td>{i}</td>
                        <td>{finding['target']}</td>
                        <td>{finding['ruleName']}</td>
                        <td><span class="risk-badge {finding['riskLevel']}">{finding['riskLevel']}</span></td>
                        <td>{finding['category']}</td>
                        <td><code class="matched-text">{matched_text}</code></td>
                        <td>{line_number}</td>
                    </tr>"""
    
    html_content += """
                </tbody>
            </table>
        </div>

        <div class="footer">
            <p>此报告由 Stellar 安全平台自动生成</p>
        </div>
    </div>
</body>
</html>"""
    
    with open(filename, 'w', encoding='utf-8') as f:
        f.write(html_content)
    
    print(f"✅ HTML 报告已生成: {filename}")
    return filename


def generate_json_report(detection_result, filename="sensitive_report_demo.json"):
    """生成JSON报告"""
    report = {
        "metadata": {
            "detectionId": detection_result['id'],
            "name": detection_result['name'],
            "projectId": detection_result['projectId'],
            "generatedAt": datetime.now().isoformat(),
            "totalFindings": len(detection_result['findings']),
            "status": detection_result['status'],
            "startTime": detection_result['startTime'],
            "endTime": detection_result['endTime']
        },
        "summary": {
            "riskStatistics": detection_result['summary']['riskLevelCount'],
            "categoryStatistics": detection_result['summary']['categoryCount'],
            "targetStatistics": {target: 1 for target in detection_result['targets']}
        },
        "findings": detection_result['findings']
    }
    
    with open(filename, 'w', encoding='utf-8') as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
    
    print(f"✅ JSON 报告已生成: {filename}")
    return filename


def generate_csv_report(detection_result, filename="sensitive_report_demo.csv"):
    """生成CSV报告"""
    import csv
    
    with open(filename, 'w', newline='', encoding='utf-8') as f:
        writer = csv.writer(f)
        
        # 写入头部
        headers = [
            "序号", "目标", "目标类型", "规则名称", "风险等级", "分类",
            "匹配文本", "行号", "上下文", "发现时间"
        ]
        writer.writerow(headers)
        
        # 写入数据
        for i, finding in enumerate(detection_result['findings'], 1):
            row = [
                i,
                finding['target'],
                finding['targetType'],
                finding['ruleName'],
                finding['riskLevel'],
                finding['category'],
                finding['matchedText'][:100] + "..." if len(finding['matchedText']) > 100 else finding['matchedText'],
                finding.get('lineNumber', ''),
                finding['context'][:150] + "..." if len(finding['context']) > 150 else finding['context'],
                finding['createdAt']
            ]
            writer.writerow(row)
    
    print(f"✅ CSV 报告已生成: {filename}")
    return filename


def generate_txt_report(detection_result, filename="sensitive_report_demo.txt"):
    """生成文本报告"""
    content = f"""================================================================
                   敏感信息检测报告
================================================================

检测信息:
  检测名称: {detection_result['name']}
  项目ID: {detection_result['projectId']}
  开始时间: {detection_result['startTime']}
  结束时间: {detection_result['endTime']}
  检测状态: {detection_result['status']}
  生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

统计摘要:
  总发现数: {detection_result['summary']['totalFindings']}
  风险等级分布:
"""
    
    for level, count in detection_result['summary']['riskLevelCount'].items():
        content += f"    {level}: {count}\n"
    
    content += "  分类分布:\n"
    for category, count in detection_result['summary']['categoryCount'].items():
        content += f"    {category}: {count}\n"
    
    content += "\n详细发现:\n"
    content += "-" * 65 + "\n"
    
    for i, finding in enumerate(detection_result['findings'], 1):
        content += f"""发现 #{i}:
  目标: {finding['target']}
  类型: {finding['targetType']}
  规则: {finding['ruleName']}
  风险等级: {finding['riskLevel']}
  分类: {finding['category']}
  匹配文本: {finding['matchedText'][:200] + '...' if len(finding['matchedText']) > 200 else finding['matchedText']}
"""
        if finding.get('lineNumber'):
            content += f"  行号: {finding['lineNumber']}\n"
        if finding.get('context'):
            context = finding['context'][:300] + '...' if len(finding['context']) > 300 else finding['context']
            content += f"  上下文: {context}\n"
        content += f"  发现时间: {finding['createdAt']}\n"
        content += "-" * 65 + "\n"
    
    with open(filename, 'w', encoding='utf-8') as f:
        f.write(content)
    
    print(f"✅ TXT 报告已生成: {filename}")
    return filename


def main():
    """主函数"""
    print("🚀 敏感信息检测报告生成功能演示\n")
    
    # 创建模拟检测结果
    print("📊 创建模拟检测结果...")
    detection_result = create_mock_detection_result()
    print(f"   检测名称: {detection_result['name']}")
    print(f"   发现数量: {detection_result['summary']['totalFindings']}")
    print(f"   风险分布: {detection_result['summary']['riskLevelCount']}")
    print()
    
    # 生成不同格式的报告
    print("📋 生成报告...")
    
    # HTML 报告
    html_file = generate_html_report(detection_result)
    
    # JSON 报告
    json_file = generate_json_report(detection_result)
    
    # CSV 报告
    csv_file = generate_csv_report(detection_result)
    
    # TXT 报告
    txt_file = generate_txt_report(detection_result)
    
    print(f"\n🎉 报告生成完成！")
    print(f"📁 生成的文件:")
    for filename in [html_file, json_file, csv_file, txt_file]:
        size = os.path.getsize(filename)
        print(f"   - {filename} ({size} 字节)")
    
    print(f"\n💡 提示: 你可以用浏览器打开 {html_file} 查看报告")
    

if __name__ == "__main__":
    main()
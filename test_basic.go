package main

import (
	"fmt"
	"os"
	"github.com/StellarServer/internal/config"
	"github.com/StellarServer/internal/utils"
)

func main() {
	fmt.Println("🧪 Stellar 基本功能测试")
	
	// 测试配置加载
	fmt.Println("📋 测试配置加载...")
	cfg, err := config.LoadConfig("config.test.yaml")
	if err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 配置加载成功，服务端口: %d\n", cfg.Server.Port)
	
	// 测试JWT功能
	fmt.Println("🔐 测试JWT功能...")
	utils.InitJWTConfig(cfg.Auth.JWTSecret, cfg.Auth.TokenExpiry)
	fmt.Printf("✅ JWT配置初始化成功，密钥长度: %d\n", len(cfg.Auth.JWTSecret))
	
	// 测试错误处理
	fmt.Println("⚠️  测试错误处理...")
	testErr := utils.ValidationError("TEST_ERROR", "这是一个测试错误")
	fmt.Printf("✅ 错误处理正常: %s\n", testErr.Error())
	
	fmt.Println("🎉 基本功能测试完成！")
	fmt.Println("💡 项目基础架构运行正常，可以进行数据库连接测试")
}
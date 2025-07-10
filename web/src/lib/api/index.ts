/**
 * API 统一入口文件
 * 导出所有API模块和通用配置
 */

// 导出 axios 配置的默认实例
export { default as api } from './axios-config';

// 导出所有API模块
export * from './auth';
export * from './assets';
export * from './asset';
export * from './projects';
export * from './tasks';
export * from './nodes';

// 重新导出主要的API实例，供其他模块使用
export { authApi } from './auth';

// 导出通用类型和工具
export type { APIError } from './auth';

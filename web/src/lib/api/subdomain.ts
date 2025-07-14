/**
 * 子域名枚举API服务
 * 提供子域名枚举任务的创建、查询、管理等功能
 */

import api from './axios-config';
import type {
	SubdomainTask,
	SubdomainResult,
	SubdomainTaskCreateRequest,
	TaskListResponse
} from '$lib/types/subdomain';
import type { ApiResponse, PaginationParams } from '$lib/types/api';

/**
 * 子域名枚举API服务类
 */
export class SubdomainAPI {
	/**
	 * 创建子域名枚举任务
	 * @param taskData 任务创建请求数据
	 * @returns 创建的任务信息
	 */
	async createTask(
		taskData: SubdomainTaskCreateRequest
	): Promise<{ taskId: string; message: string }> {
		const response = await api.post('/api/v1/subdomains/tasks', {
			projectId: taskData.projectId,
			rootDomain: taskData.target,
			taskName: taskData.name,
			config: {
				dictionaryPath: taskData.wordlistPath || 'dicts/subdomain_dict.txt',
				methods: taskData.enumMethods || ['dns_brute'],
				concurrency: taskData.maxWorkers || 50,
				timeout: taskData.timeout || 5,
				retryCount: taskData.maxRetries || 3,
				rateLimit: taskData.rateLimit || 100,
				resolverServers: taskData.dnsServers || ['8.8.8.8', '1.1.1.1'],
				verifySubdomains: taskData.verifySubdomains !== false,
				recursiveSearch: taskData.enableRecursive || false,
				saveToDB: true
			}
		});

		return response.data;
	}

	/**
	 * 获取子域名枚举任务列表
	 * @param params 分页和过滤参数
	 * @returns 任务列表
	 */
	async getTasks(
		params?: PaginationParams & {
			projectId?: string;
			status?: string;
			target?: string;
		}
	): Promise<TaskListResponse> {
		const queryParams: any = {};

		if (params?.limit) queryParams.limit = params.limit;
		if (params?.page) queryParams.skip = (params.page - 1) * (params.limit || 10);
		if (params?.projectId) queryParams.projectId = params.projectId;
		if (params?.status) queryParams.status = params.status;
		if (params?.target) queryParams.target = params.target;

		const response = await api.get('/api/v1/subdomains/tasks', { params: queryParams });

		return {
			tasks: response.data.tasks,
			total: response.data.total,
			page: params?.page || 1,
			limit: params?.limit || 10,
			totalPages: Math.ceil(response.data.total / (params?.limit || 10))
		};
	}

	/**
	 * 获取单个子域名枚举任务详情
	 * @param taskId 任务ID
	 * @returns 任务详情
	 */
	async getTask(taskId: string): Promise<SubdomainTask> {
		const response = await api.get(`/api/v1/subdomains/tasks/${taskId}`);
		return response.data;
	}

	/**
	 * 获取子域名枚举任务结果
	 * @param taskId 任务ID
	 * @param params 分页参数
	 * @returns 枚举结果
	 */
	async getTaskResults(
		taskId: string,
		params?: PaginationParams
	): Promise<{
		results: SubdomainResult[];
		total: number;
		summary: any;
	}> {
		const queryParams: any = {};
		if (params?.limit) queryParams.limit = params.limit;
		if (params?.page) queryParams.skip = (params.page - 1) * (params.limit || 10);

		const response = await api.get(`/api/v1/subdomains/tasks/${taskId}/results`, {
			params: queryParams
		});

		return response.data;
	}

	/**
	 * 删除子域名枚举任务
	 * @param taskId 任务ID
	 */
	async deleteTask(taskId: string): Promise<void> {
		await api.delete(`/api/v1/subdomains/tasks/${taskId}`);
	}

	/**
	 * 取消正在运行的子域名枚举任务
	 * @param taskId 任务ID
	 */
	async cancelTask(taskId: string): Promise<void> {
		await api.post(`/api/v1/subdomains/tasks/${taskId}/cancel`);
	}

	/**
	 * 重新运行子域名枚举任务
	 * @param taskId 任务ID
	 */
	async retryTask(taskId: string): Promise<SubdomainTask> {
		// 获取原任务信息，重新创建
		const originalTask = await this.getTask(taskId);
		const newTaskResponse = await this.createTask({
			name: originalTask.name + ' (重试)',
			target: originalTask.rootDomain,
			projectId: originalTask.projectId,
			wordlistPath: originalTask.config?.dictionaryPath,
			enumMethods: originalTask.config?.methods,
			maxWorkers: originalTask.config?.concurrency,
			timeout: originalTask.config?.timeout,
			maxRetries: originalTask.config?.retryCount,
			rateLimit: originalTask.config?.rateLimit,
			dnsServers: originalTask.config?.resolverServers,
			verifySubdomains: originalTask.config?.verifySubdomains,
			enableRecursive: originalTask.config?.recursiveSearch
		});

		return this.getTask(newTaskResponse.taskId);
	}

	/**
	 * 导出子域名枚举结果
	 * @param taskId 任务ID
	 * @param format 导出格式 ('csv' | 'json' | 'xlsx')
	 * @returns 文件下载URL或数据
	 */
	async exportResults(
		taskId: string,
		format: 'csv' | 'json' | 'xlsx' = 'csv'
	): Promise<{
		downloadUrl?: string;
		data?: any;
		filename: string;
	}> {
		const response = await api.get(`/api/v1/subdomains/tasks/${taskId}/export?format=${format}`);
		return response.data;
	}

	/**
	 * 获取子域名枚举统计信息
	 * @param params 统计参数
	 * @returns 统计数据
	 */
	async getStatistics(params?: {
		projectId?: string;
		dateRange?: {
			startDate: string;
			endDate: string;
		};
	}): Promise<{
		totalTasks: number;
		completedTasks: number;
		failedTasks: number;
		totalSubdomains: number;
		uniqueSubdomains: number;
		sourceStats: Record<string, number>;
		recentTasks: SubdomainTask[];
	}> {
		// 获取任务列表并统计
		const tasks = await this.getTasks({ projectId: params?.projectId, limit: 1000 });
		const totalTasks = tasks.total;
		const completedTasks = tasks.tasks.filter((t) => t.status === 'completed').length;
		const failedTasks = tasks.tasks.filter((t) => t.status === 'failed').length;

		// 简化的统计信息
		return {
			totalTasks,
			completedTasks,
			failedTasks,
			totalSubdomains: 0, // 需要从结果中统计
			uniqueSubdomains: 0, // 需要从结果中统计
			sourceStats: {},
			recentTasks: tasks.tasks.slice(0, 5)
		};
	}

	/**
	 * 获取预设枚举方法配置
	 * @returns 预设枚举方法配置列表
	 */
	getPresetMethods(): Record<string, { name: string; methods: string[]; description: string }> {
		return {
			basic: {
				name: '基础枚举',
				methods: ['dns_brute'],
				description: '仅使用DNS暴力破解，速度快，结果较少'
			},
			standard: {
				name: '标准枚举',
				methods: ['dns_brute', 'cert_transparency'],
				description: '结合DNS暴力破解和证书透明度日志，平衡速度和结果'
			},
			comprehensive: {
				name: '全面枚举',
				methods: ['dns_brute', 'cert_transparency', 'search_engine'],
				description: '使用所有可用方法，结果最全面但耗时较长'
			},
			passive: {
				name: '被动枚举',
				methods: ['cert_transparency', 'search_engine'],
				description: '仅使用被动信息收集，不主动发送DNS请求'
			}
		};
	}

	/**
	 * 获取预设DNS服务器配置
	 * @returns 预设DNS服务器配置列表
	 */
	getPresetDNSServers(): Record<string, { name: string; servers: string[]; description: string }> {
		return {
			public: {
				name: '公共DNS',
				servers: ['8.8.8.8', '1.1.1.1', '208.67.222.222', '9.9.9.9'],
				description: '使用知名的公共DNS服务器'
			},
			china: {
				name: '国内DNS',
				servers: ['114.114.114.114', '223.5.5.5', '119.29.29.29', '180.76.76.76'],
				description: '使用国内的DNS服务器，适合国内目标'
			},
			custom: {
				name: '自定义',
				servers: [],
				description: '使用自定义的DNS服务器列表'
			}
		};
	}

	/**
	 * 获取预设字典配置
	 * @returns 预设字典配置列表
	 */
	getPresetWordlists(): Record<string, { name: string; size: string; description: string }> {
		return {
			common: {
				name: '常用字典',
				size: '~1000条',
				description: '包含最常见的子域名前缀，适合快速扫描'
			},
			comprehensive: {
				name: '全面字典',
				size: '~10000条',
				description: '包含大量子域名前缀，覆盖率高但耗时较长'
			},
			security: {
				name: '安全字典',
				size: '~5000条',
				description: '专注于安全相关的子域名前缀'
			},
			custom: {
				name: '自定义字典',
				size: '可变',
				description: '使用用户提供的自定义字典'
			}
		};
	}

	/**
	 * 验证域名格式
	 * @param domain 域名字符串
	 * @returns 验证结果
	 */
	validateDomain(domain: string): { valid: boolean; message?: string } {
		if (!domain.trim()) {
			return { valid: false, message: '域名不能为空' };
		}

		// 简单的域名格式验证
		const domainRegex = /^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
		if (!domainRegex.test(domain.trim())) {
			return { valid: false, message: '域名格式不正确' };
		}

		// 检查域名长度
		if (domain.length > 253) {
			return { valid: false, message: '域名长度不能超过253个字符' };
		}

		// 检查各级域名长度
		const parts = domain.split('.');
		for (const part of parts) {
			if (part.length === 0 || part.length > 63) {
				return { valid: false, message: '域名各级长度必须在1-63个字符之间' };
			}
		}

		return { valid: true };
	}

	/**
	 * 验证自定义字典
	 * @param wordlist 字典数组
	 * @returns 验证结果
	 */
	validateWordlist(wordlist: string[]): { valid: boolean; message?: string; count?: number } {
		if (!wordlist || wordlist.length === 0) {
			return { valid: false, message: '字典不能为空' };
		}

		// 检查字典大小
		if (wordlist.length > 50000) {
			return { valid: false, message: '字典条目不能超过50000条' };
		}

		// 检查字典项格式
		const invalidItems = wordlist.filter((item) => {
			if (typeof item !== 'string' || item.length === 0) return true;
			if (item.length > 63) return true;
			if (!/^[a-zA-Z0-9-]+$/.test(item)) return true;
			return false;
		});

		if (invalidItems.length > 0) {
			return { valid: false, message: `发现${invalidItems.length}个无效的字典项` };
		}

		return { valid: true, count: wordlist.length };
	}

	/**
	 * 估算扫描时间
	 * @param config 扫描配置
	 * @returns 估算的扫描时间（秒）
	 */
	estimateScanTime(config: {
		wordlistSize: number;
		enumMethods: string[];
		maxWorkers: number;
		rateLimit: number;
		enableRecursive: boolean;
	}): number {
		let baseTime = 0;

		// DNS暴力破解时间估算
		if (config.enumMethods.includes('dns_brute')) {
			const dnsTime = config.wordlistSize / Math.min(config.maxWorkers, config.rateLimit);
			baseTime += dnsTime;
		}

		// 证书透明度查询时间（相对固定）
		if (config.enumMethods.includes('cert_transparency')) {
			baseTime += 30; // 约30秒
		}

		// 搜索引擎查询时间（相对固定）
		if (config.enumMethods.includes('search_engine')) {
			baseTime += 60; // 约60秒
		}

		// 递归枚举会增加时间
		if (config.enableRecursive) {
			baseTime *= 1.5; // 增加50%的时间
		}

		return Math.ceil(baseTime);
	}
}

// 导出单例实例
export const subdomainApi = new SubdomainAPI();
export default subdomainApi;

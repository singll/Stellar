/**
 * 端口扫描API服务
 * 提供端口扫描任务的创建、查询、管理等功能
 */

import api from './axios-config';
import type {
	PortScanTask,
	PortScanResult,
	TaskCreateRequest,
	TaskListResponse
} from '$lib/types/portscan';
import type { ApiResponse, PaginationParams } from '$lib/types/api';

/**
 * 端口扫描API服务类
 */
export class PortScanAPI {
	/**
	 * 创建端口扫描任务
	 * @param taskData 任务创建请求数据
	 * @returns 创建的任务信息
	 */
	async createTask(taskData: TaskCreateRequest): Promise<PortScanTask> {
		const response = await api.post<ApiResponse<PortScanTask>>('/api/v1/tasks', {
			name: taskData.name,
			description: taskData.description,
			type: 'port_scan',
			config: {
				target: taskData.target,
				ports: taskData.ports,
				scan_method: taskData.scanMethod || 'tcp',
				max_workers: taskData.maxWorkers || 100,
				timeout: taskData.timeout || 30,
				enable_banner: taskData.enableBanner || false,
				enable_ssl: taskData.enableSSL || false,
				enable_service: taskData.enableService || true,
				rate_limit: taskData.rateLimit || 100
			},
			projectId: taskData.projectId
		});

		if (!response.data.success) {
			throw new Error(response.data.message || 'Failed to create port scan task');
		}

		return response.data.data;
	}

	/**
	 * 获取端口扫描任务列表
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
		const searchParams = new URLSearchParams();

		if (params?.page) searchParams.set('page', params.page.toString());
		if (params?.limit) searchParams.set('limit', params.limit.toString());
		if (params?.projectId) searchParams.set('project_id', params.projectId);
		if (params?.status) searchParams.set('status', params.status);
		if (params?.target) searchParams.set('target', params.target);

		// 只获取端口扫描任务
		searchParams.set('type', 'port_scan');

		const response = await api.get<ApiResponse<TaskListResponse>>(`/api/v1/tasks?${searchParams}`);

		if (!response.data.success) {
			throw new Error(response.data.message || 'Failed to fetch port scan tasks');
		}

		return response.data.data;
	}

	/**
	 * 获取单个端口扫描任务详情
	 * @param taskId 任务ID
	 * @returns 任务详情
	 */
	async getTask(taskId: string): Promise<PortScanTask> {
		const response = await api.get<ApiResponse<PortScanTask>>(`/api/v1/tasks/${taskId}`);

		if (!response.data.success) {
			throw new Error(response.data.message || 'Failed to fetch port scan task');
		}

		return response.data.data;
	}

	/**
	 * 获取端口扫描任务结果
	 * @param taskId 任务ID
	 * @param params 分页参数
	 * @returns 扫描结果
	 */
	async getTaskResults(
		taskId: string,
		params?: PaginationParams
	): Promise<{
		results: PortScanResult[];
		total: number;
		summary: any;
	}> {
		const searchParams = new URLSearchParams();

		if (params?.page) searchParams.set('page', params.page.toString());
		if (params?.limit) searchParams.set('limit', params.limit.toString());

		const response = await api.get<
			ApiResponse<{
				results: PortScanResult[];
				total: number;
				summary: any;
			}>
		>(`/api/v1/tasks/${taskId}/results?${searchParams}`);

		if (!response.data.success) {
			throw new Error(response.data.message || 'Failed to fetch task results');
		}

		return response.data.data;
	}

	/**
	 * 删除端口扫描任务
	 * @param taskId 任务ID
	 */
	async deleteTask(taskId: string): Promise<void> {
		const response = await api.delete<ApiResponse<void>>(`/api/v1/tasks/${taskId}`);

		if (!response.data.success) {
			throw new Error(response.data.message || 'Failed to delete port scan task');
		}
	}

	/**
	 * 取消正在运行的端口扫描任务
	 * @param taskId 任务ID
	 */
	async cancelTask(taskId: string): Promise<void> {
		const response = await api.post<ApiResponse<void>>(`/api/v1/tasks/${taskId}/cancel`);

		if (!response.data.success) {
			throw new Error(response.data.message || 'Failed to cancel port scan task');
		}
	}

	/**
	 * 重新运行端口扫描任务
	 * @param taskId 任务ID
	 */
	async retryTask(taskId: string): Promise<PortScanTask> {
		const response = await api.post<ApiResponse<PortScanTask>>(`/api/v1/tasks/${taskId}/retry`);

		if (!response.data.success) {
			throw new Error(response.data.message || 'Failed to retry port scan task');
		}

		return response.data.data;
	}

	/**
	 * 导出端口扫描结果
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
		const response = await api.get<
			ApiResponse<{
				downloadUrl?: string;
				data?: any;
				filename: string;
			}>
		>(`/api/v1/tasks/${taskId}/export?format=${format}`);

		if (!response.data.success) {
			throw new Error(response.data.message || 'Failed to export results');
		}

		return response.data.data;
	}

	/**
	 * 获取端口扫描统计信息
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
		totalPorts: number;
		openPorts: number;
		serviceStats: Record<string, number>;
		recentTasks: PortScanTask[];
	}> {
		const searchParams = new URLSearchParams();

		if (params?.projectId) searchParams.set('project_id', params.projectId);
		if (params?.dateRange) {
			searchParams.set('start_date', params.dateRange.startDate);
			searchParams.set('end_date', params.dateRange.endDate);
		}

		searchParams.set('type', 'port_scan');

		const response = await api.get<
			ApiResponse<{
				totalTasks: number;
				completedTasks: number;
				failedTasks: number;
				totalPorts: number;
				openPorts: number;
				serviceStats: Record<string, number>;
				recentTasks: PortScanTask[];
			}>
		>(`/api/v1/tasks/statistics?${searchParams}`);

		if (!response.data.success) {
			throw new Error(response.data.message || 'Failed to fetch statistics');
		}

		return response.data.data;
	}

	/**
	 * 获取预设端口配置
	 * @returns 预设端口配置列表
	 */
	getPresetPorts(): Record<string, { name: string; ports: string; description: string }> {
		return {
			top_100: {
				name: '常用100端口',
				ports:
					'21,22,23,25,53,80,110,111,135,139,143,443,993,995,1723,3306,3389,5432,5900,6379,8080,8443,9200,27017',
				description: '最常用的100个端口，适合快速扫描'
			},
			top_1000: {
				name: '常用1000端口',
				ports: '1-1000',
				description: '前1000个端口，平衡速度和覆盖率'
			},
			web_ports: {
				name: 'Web服务端口',
				ports: '80,443,8000,8001,8008,8080,8443,8888,9000,9001,9999',
				description: '常见Web服务端口'
			},
			database_ports: {
				name: '数据库端口',
				ports: '1433,1521,3306,5432,6379,9042,27017,28017',
				description: '常见数据库服务端口'
			},
			all_ports: {
				name: '全端口扫描',
				ports: '1-65535',
				description: '扫描所有端口（耗时较长）'
			}
		};
	}

	/**
	 * 验证端口配置
	 * @param ports 端口配置字符串
	 * @returns 验证结果
	 */
	validatePorts(ports: string): { valid: boolean; message?: string; count?: number } {
		if (!ports.trim()) {
			return { valid: false, message: '端口配置不能为空' };
		}

		try {
			const portList = this.parsePorts(ports);
			if (portList.length === 0) {
				return { valid: false, message: '未解析到有效端口' };
			}

			if (portList.length > 10000) {
				return { valid: false, message: '端口数量过多，建议控制在10000个以内' };
			}

			return { valid: true, count: portList.length };
		} catch (error) {
			return { valid: false, message: `端口配置格式错误: ${error}` };
		}
	}

	/**
	 * 解析端口配置字符串
	 * @param ports 端口配置
	 * @returns 端口数组
	 */
	private parsePorts(ports: string): number[] {
		const portList: number[] = [];
		const parts = ports.split(',').map((p) => p.trim());

		for (const part of parts) {
			if (part.includes('-')) {
				// 端口范围
				const [start, end] = part.split('-').map((p) => parseInt(p.trim()));
				if (isNaN(start) || isNaN(end) || start > end || start < 1 || end > 65535) {
					throw new Error(`无效的端口范围: ${part}`);
				}
				for (let i = start; i <= end; i++) {
					portList.push(i);
				}
			} else {
				// 单个端口
				const port = parseInt(part);
				if (isNaN(port) || port < 1 || port > 65535) {
					throw new Error(`无效的端口号: ${part}`);
				}
				portList.push(port);
			}
		}

		// 去重并排序
		return [...new Set(portList)].sort((a, b) => a - b);
	}
}

// 导出单例实例
export const portScanApi = new PortScanAPI();
export default portScanApi;

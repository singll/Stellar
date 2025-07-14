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
	async createTask(taskData: TaskCreateRequest): Promise<{ taskId: string; message: string }> {
		const response = await api.post(
			'/api/v1/portscan/tasks',
			{
				name: taskData.name,
				description: taskData.description,
				targets: Array.isArray(taskData.target) ? taskData.target : [taskData.target],
				config: {
					ports: taskData.ports,
					scan_type: taskData.scanMethod || 'tcp',
					scan_method: 'connect',
					concurrency: taskData.maxWorkers || 100,
					timeout: taskData.timeout || 3,
					retry_count: 2,
					rate_limit: taskData.rateLimit || 100,
					service_detection: taskData.enableService || true,
					save_to_db: true
				}
			},
			{
				params: {
					projectId: taskData.projectId
				}
			}
		);

		return response.data;
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
		const queryParams: any = {};

		if (params?.limit) queryParams.limit = params.limit;
		if (params?.page) queryParams.skip = (params.page - 1) * (params.limit || 10);
		if (params?.projectId) queryParams.projectId = params.projectId;
		if (params?.status) queryParams.status = params.status;

		const response = await api.get('/api/v1/portscan/tasks', { params: queryParams });

		return {
			tasks: response.data.tasks,
			total: response.data.total,
			page: params?.page || 1,
			limit: params?.limit || 10,
			totalPages: Math.ceil(response.data.total / (params?.limit || 10))
		};
	}

	/**
	 * 获取单个端口扫描任务详情
	 * @param taskId 任务ID
	 * @returns 任务详情
	 */
	async getTask(taskId: string): Promise<PortScanTask> {
		const response = await api.get(`/api/v1/portscan/tasks/${taskId}`);
		return response.data;
	}

	/**
	 * 启动端口扫描任务
	 * @param taskId 任务ID
	 */
	async startTask(taskId: string): Promise<{ message: string; taskId: string }> {
		const response = await api.post(`/api/v1/portscan/tasks/${taskId}/start`);
		return response.data;
	}

	/**
	 * 停止端口扫描任务
	 * @param taskId 任务ID
	 */
	async stopTask(taskId: string): Promise<{ message: string; taskId: string }> {
		const response = await api.post(`/api/v1/portscan/tasks/${taskId}/stop`);
		return response.data;
	}

	/**
	 * 获取任务状态
	 * @param taskId 任务ID
	 */
	async getTaskStatus(taskId: string): Promise<{ status: string; taskId: string }> {
		const response = await api.get(`/api/v1/portscan/tasks/${taskId}/status`);
		return response.data;
	}

	/**
	 * 获取任务进度
	 * @param taskId 任务ID
	 */
	async getTaskProgress(taskId: string): Promise<{ progress: number; taskId: string }> {
		const response = await api.get(`/api/v1/portscan/tasks/${taskId}/progress`);
		return response.data;
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
		taskId: string;
	}> {
		const queryParams: any = {};
		if (params?.limit) queryParams.limit = params.limit;
		if (params?.page) queryParams.skip = (params.page - 1) * (params.limit || 10);

		const response = await api.get(`/api/v1/portscan/tasks/${taskId}/results`, {
			params: queryParams
		});
		return response.data;
	}

	/**
	 * 删除端口扫描任务
	 * @param taskId 任务ID
	 */
	async deleteTask(taskId: string): Promise<void> {
		await api.delete(`/api/v1/portscan/tasks/${taskId}`);
	}

	/**
	 * 取消正在运行的端口扫描任务
	 * @param taskId 任务ID
	 */
	async cancelTask(taskId: string): Promise<void> {
		await this.stopTask(taskId);
	}

	/**
	 * 重新运行端口扫描任务
	 * @param taskId 任务ID
	 */
	async retryTask(taskId: string): Promise<PortScanTask> {
		// 获取原任务信息，重新创建
		const originalTask = await this.getTask(taskId);
		const newTaskResponse = await this.createTask({
			name: originalTask.name + ' (重试)',
			description: originalTask.description,
			target: originalTask.config?.target || 'unknown',
			ports: originalTask.config?.ports || '1-1000',
			projectId: originalTask.projectId
		});

		// 启动新任务
		await this.startTask(newTaskResponse.taskId);
		return this.getTask(newTaskResponse.taskId);
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
		const response = await api.get(`/api/v1/portscan/tasks/${taskId}/export?format=${format}`);
		return response.data;
	}

	/**
	 * 轮询任务状态直到完成
	 */
	async pollTaskStatus(
		taskId: string,
		options?: {
			interval?: number;
			timeout?: number;
			onProgress?: (progress: number) => void;
			onStatusChange?: (status: string) => void;
		}
	): Promise<PortScanTask> {
		const { interval = 2000, timeout = 300000 } = options || {};
		const startTime = Date.now();

		return new Promise((resolve, reject) => {
			const poll = async () => {
				try {
					if (Date.now() - startTime > timeout) {
						reject(new Error('任务轮询超时'));
						return;
					}

					const [statusResp, progressResp] = await Promise.all([
						this.getTaskStatus(taskId),
						this.getTaskProgress(taskId)
					]);

					// 调用进度回调
					if (options?.onProgress) {
						options.onProgress(progressResp.progress);
					}

					// 调用状态变化回调
					if (options?.onStatusChange) {
						options.onStatusChange(statusResp.status);
					}

					// 检查任务是否完成
					if (
						statusResp.status === 'completed' ||
						statusResp.status === 'failed' ||
						statusResp.status === 'stopped'
					) {
						const task = await this.getTask(taskId);
						resolve(task);
						return;
					}

					// 继续轮询
					setTimeout(poll, interval);
				} catch (error) {
					reject(error);
				}
			};

			poll();
		});
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
			totalPorts: 0, // 需要从结果中统计
			openPorts: 0, // 需要从结果中统计
			serviceStats: {},
			recentTasks: tasks.tasks.slice(0, 5)
		};
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

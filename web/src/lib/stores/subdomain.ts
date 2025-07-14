/**
 * 子域名枚举状态管理Store
 * 使用Svelte 5 runes语法管理子域名枚举相关状态
 */

import type {
	SubdomainTask,
	SubdomainResult,
	SubdomainFilter,
	SubdomainStatistics,
	ProgressUpdateEvent
} from '$lib/types/subdomain';
import { subdomainApi } from '$lib/api/subdomain';
import { toastStore } from '$lib/stores/toast';

// 使用Svelte 5 runes创建响应式状态
let tasks = $state<SubdomainTask[]>([]);
let currentTask = $state<SubdomainTask | null>(null);
let taskResults = $state<SubdomainResult[]>([]);

// 加载状态
let loading = $state({
	tasks: false,
	taskDetail: false,
	results: false,
	creating: false,
	statistics: false
});

// 分页状态
let pagination = $state({
	page: 1,
	limit: 20,
	total: 0,
	totalPages: 0
});

// 过滤器状态
let filters = $state<SubdomainFilter>({});

// 统计数据
let statistics = $state<SubdomainStatistics | null>(null);

// 实时进度更新
let progressUpdates = $state<Record<string, ProgressUpdateEvent>>({});

// 派生状态：运行中的任务
let runningTasks = $derived(
	tasks.filter((task) => task.status === 'running' || task.status === 'queued')
);

// 派生状态：最近完成的任务
let recentCompletedTasks = $derived(
	tasks
		.filter((task) => task.status === 'completed')
		.sort(
			(a, b) =>
				new Date(b.endTime || b.updatedAt).getTime() - new Date(a.endTime || a.updatedAt).getTime()
		)
		.slice(0, 5)
);

// 派生状态：任务统计摘要
let taskSummary = $derived(() => {
	const total = tasks.length;
	const completed = tasks.filter((t) => t.status === 'completed').length;
	const running = tasks.filter((t) => t.status === 'running').length;
	const failed = tasks.filter((t) => t.status === 'failed').length;
	const pending = tasks.filter((t) => t.status === 'pending' || t.status === 'queued').length;

	return {
		total,
		completed,
		running,
		failed,
		pending,
		successRate: total > 0 ? Math.round((completed / total) * 100) : 0
	};
});

// 派生状态：子域名结果统计
let resultSummary = $derived(() => {
	const total = taskResults.length;
	const active = taskResults.filter((r) => r.httpStatus && r.httpStatus < 400).length;
	const withCNAME = taskResults.filter((r) => r.cname).length;
	const vulnerable = taskResults.filter((r) => r.takeover?.vulnerable).length;

	// 按来源统计
	const sourceStats: Record<string, number> = {};
	taskResults.forEach((result) => {
		sourceStats[result.source] = (sourceStats[result.source] || 0) + 1;
	});

	// 按IP统计
	const ipStats: Record<string, number> = {};
	taskResults.forEach((result) => {
		result.ips.forEach((ip) => {
			ipStats[ip] = (ipStats[ip] || 0) + 1;
		});
	});

	return {
		total,
		active,
		withCNAME,
		vulnerable,
		sourceStats,
		ipStats,
		uniqueIPs: Object.keys(ipStats).length
	};
});

/**
 * 子域名枚举操作函数
 */
const subdomainActions = {
	/**
	 * 加载任务列表
	 */
	async loadTasks(params?: { page?: number; limit?: number; filters?: SubdomainFilter }) {
		loading.tasks = true;

		try {
			const currentFilters = params?.filters || {};
			const response = await subdomainApi.getTasks({
				page: params?.page || 1,
				limit: params?.limit || 20,
				...currentFilters
			});

			tasks = response.tasks;
			pagination.page = response.page;
			pagination.limit = response.limit;
			pagination.total = response.total;
			pagination.totalPages = Math.ceil(response.total / response.limit);

			if (params?.filters) {
				filters = currentFilters;
			}
		} catch (error) {
			console.error('Failed to load subdomain tasks:', error);
			toastStore.error('加载子域名枚举任务失败: ' + (error as Error).message);
		} finally {
			loading.tasks = false;
		}
	},

	/**
	 * 加载任务详情
	 */
	async loadTask(taskId: string) {
		loading.taskDetail = true;

		try {
			const task = await subdomainApi.getTask(taskId);
			currentTask = task;
			return task;
		} catch (error) {
			console.error('Failed to load task:', error);
			toastStore.error('加载任务详情失败: ' + (error as Error).message);
			throw error;
		} finally {
			loading.taskDetail = false;
		}
	},

	/**
	 * 加载任务结果
	 */
	async loadTaskResults(taskId: string, params?: { page?: number; limit?: number }) {
		loading.results = true;

		try {
			const response = await subdomainApi.getTaskResults(taskId, {
				page: params?.page || 1,
				limit: params?.limit || 100
			});

			taskResults = response.results;
			return response;
		} catch (error) {
			console.error('Failed to load task results:', error);
			toastStore.error('加载任务结果失败: ' + (error as Error).message);
			throw error;
		} finally {
			loading.results = false;
		}
	},

	/**
	 * 创建新任务
	 */
	async createTask(taskData: any) {
		loading.creating = true;

		try {
			const response = await subdomainApi.createTask(taskData);

			// 创建成功后重新加载任务列表
			await this.loadTasks();

			toastStore.success(`子域名枚举任务创建成功`);
			return response;
		} catch (error) {
			console.error('Failed to create task:', error);
			toastStore.error('创建任务失败: ' + (error as Error).message);
			throw error;
		} finally {
			loading.creating = false;
		}
	},

	/**
	 * 删除任务
	 */
	async deleteTask(taskId: string) {
		try {
			await subdomainApi.deleteTask(taskId);

			// 从任务列表中移除
			tasks = tasks.filter((t) => t.id !== taskId);

			// 如果是当前任务，清除
			if (currentTask?.id === taskId) {
				currentTask = null;
			}

			toastStore.success('任务删除成功');
		} catch (error) {
			console.error('Failed to delete task:', error);
			toastStore.error('删除任务失败: ' + (error as Error).message);
			throw error;
		}
	},

	/**
	 * 取消任务
	 */
	async cancelTask(taskId: string) {
		try {
			await subdomainApi.cancelTask(taskId);

			// 更新任务状态
			tasks = tasks.map((t) => (t.id === taskId ? { ...t, status: 'canceled' as const } : t));

			if (currentTask?.id === taskId) {
				currentTask = { ...currentTask, status: 'canceled' as const };
			}

			toastStore.success('任务已取消');
		} catch (error) {
			console.error('Failed to cancel task:', error);
			toastStore.error('取消任务失败: ' + (error as Error).message);
			throw error;
		}
	},

	/**
	 * 重试任务
	 */
	async retryTask(taskId: string) {
		try {
			const task = await subdomainApi.retryTask(taskId);

			// 更新任务列表
			tasks = tasks.map((t) => (t.id === taskId ? task : t));

			if (currentTask?.id === taskId) {
				currentTask = task;
			}

			toastStore.success('任务重试成功');
			return task;
		} catch (error) {
			console.error('Failed to retry task:', error);
			toastStore.error('重试任务失败: ' + (error as Error).message);
			throw error;
		}
	},

	/**
	 * 导出结果
	 */
	async exportResults(taskId: string, format: 'csv' | 'json' | 'xlsx' = 'csv') {
		try {
			const result = await subdomainApi.exportResults(taskId, format);

			if (result.downloadUrl) {
				// 下载文件
				const link = document.createElement('a');
				link.href = result.downloadUrl;
				link.download = result.filename;
				document.body.appendChild(link);
				link.click();
				document.body.removeChild(link);
			} else if (result.data) {
				// 直接下载数据
				const blob = new Blob([JSON.stringify(result.data, null, 2)], {
					type: 'application/json'
				});
				const url = URL.createObjectURL(blob);
				const link = document.createElement('a');
				link.href = url;
				link.download = result.filename;
				document.body.appendChild(link);
				link.click();
				document.body.removeChild(link);
				URL.revokeObjectURL(url);
			}

			toastStore.success(`结果导出成功: ${result.filename}`);
		} catch (error) {
			console.error('Failed to export results:', error);
			toastStore.error('导出结果失败: ' + (error as Error).message);
			throw error;
		}
	},

	/**
	 * 加载统计数据
	 */
	async loadStatistics(params?: { projectId?: string; dateRange?: any }) {
		loading.statistics = true;

		try {
			const stats = await subdomainApi.getStatistics(params);

			// 补充缺失的属性，确保类型匹配
			const fullStats: SubdomainStatistics = {
				successRate: stats.totalTasks > 0 ? (stats.completedTasks / stats.totalTasks) * 100 : 0,
				averageDuration: 0, // API 暂未提供，设置默认值
				methodStats: {}, // API 暂未提供，设置默认值
				topTargets: [], // API 暂未提供，设置默认值
				...stats
			};

			statistics = fullStats;
			return fullStats;
		} catch (error) {
			console.error('Failed to load statistics:', error);
			toastStore.error('加载统计数据失败: ' + (error as Error).message);
		} finally {
			loading.statistics = false;
		}
	},

	/**
	 * 更新过滤器
	 */
	updateFilters(newFilters: SubdomainFilter) {
		filters = newFilters;
	},

	/**
	 * 清除过滤器
	 */
	clearFilters() {
		filters = {};
	},

	/**
	 * 刷新当前页面
	 */
	async refresh() {
		await this.loadTasks({
			page: pagination.page,
			limit: pagination.limit,
			filters: filters
		});
	},

	/**
	 * 处理实时进度更新
	 */
	handleProgressUpdate(update: ProgressUpdateEvent) {
		// 更新进度缓存
		progressUpdates[update.taskId] = update;

		// 更新任务列表中的任务状态
		tasks = tasks.map((task) => {
			if (task.id === update.taskId) {
				return {
					...task,
					progress: update.progress,
					status: update.status as any
				};
			}
			return task;
		});

		// 更新当前任务状态
		if (currentTask?.id === update.taskId) {
			currentTask = {
				...currentTask,
				progress: update.progress,
				status: update.status as any
			};
		}

		// 如果有最新结果，更新结果列表
		if (update.latestResults && update.latestResults.length > 0) {
			// 合并新结果，避免重复
			const existingIds = new Set(taskResults.map((r) => r.subdomain));
			const newResults = update.latestResults!.filter((r) => !existingIds.has(r.subdomain));
			taskResults = [...taskResults, ...newResults];
		}
	},

	/**
	 * 清除进度更新缓存
	 */
	clearProgressUpdates() {
		progressUpdates = {};
	},

	/**
	 * 重置所有状态
	 */
	reset() {
		tasks = [];
		currentTask = null;
		taskResults = [];
		statistics = null;
		filters = {};
		progressUpdates = {};
		pagination = {
			page: 1,
			limit: 20,
			total: 0,
			totalPages: 0
		};
		loading = {
			tasks: false,
			taskDetail: false,
			results: false,
			creating: false,
			statistics: false
		};
	}
};

// 导出Store对象
export const subdomainStore = {
	// 状态getters
	get tasks() {
		return tasks;
	},
	get currentTask() {
		return currentTask;
	},
	get taskResults() {
		return taskResults;
	},
	get loading() {
		return loading;
	},
	get pagination() {
		return pagination;
	},
	get filters() {
		return filters;
	},
	get statistics() {
		return statistics;
	},
	get progressUpdates() {
		return progressUpdates;
	},
	get runningTasks() {
		return runningTasks;
	},
	get recentCompletedTasks() {
		return recentCompletedTasks;
	},
	get taskSummary() {
		return taskSummary;
	},
	get resultSummary() {
		return resultSummary;
	},

	// 操作函数
	actions: subdomainActions
};

// 导出类型
export type {
	SubdomainTask,
	SubdomainResult,
	SubdomainFilter,
	SubdomainStatistics,
	ProgressUpdateEvent
};

// 默认导出
export default subdomainStore;

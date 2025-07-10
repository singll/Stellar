/**
 * 子域名枚举状态管理Store
 * 使用Svelte 5 runes语法管理子域名枚举相关状态
 */

import { writable, derived } from 'svelte/store';
import type {
	SubdomainTask,
	SubdomainResult,
	TaskFilter,
	SubdomainStatistics,
	ProgressUpdateEvent
} from '$lib/types/subdomain';
import { subdomainApi } from '$lib/api/subdomain';
import { toastStore } from '$lib/stores/toast';

// 任务列表状态
export const subdomainTasks = writable<SubdomainTask[]>([]);
export const currentTask = writable<SubdomainTask | null>(null);
export const taskResults = writable<SubdomainResult[]>([]);

// 加载状态
export const loading = writable({
	tasks: false,
	taskDetail: false,
	results: false,
	creating: false,
	statistics: false
});

// 分页状态
export const pagination = writable({
	page: 1,
	limit: 20,
	total: 0,
	totalPages: 0
});

// 过滤器状态
export const filters = writable<TaskFilter>({});

// 统计数据
export const statistics = writable<SubdomainStatistics | null>(null);

// 实时进度更新
export const progressUpdates = writable<Record<string, ProgressUpdateEvent>>({});

// 派生状态：运行中的任务
export const runningTasks = derived(subdomainTasks, ($tasks) =>
	$tasks.filter((task) => task.status === 'running' || task.status === 'queued')
);

// 派生状态：最近完成的任务
export const recentCompletedTasks = derived(subdomainTasks, ($tasks) =>
	$tasks
		.filter((task) => task.status === 'completed')
		.sort(
			(a, b) =>
				new Date(b.endTime || b.updatedAt).getTime() - new Date(a.endTime || a.updatedAt).getTime()
		)
		.slice(0, 5)
);

// 派生状态：任务统计摘要
export const taskSummary = derived(subdomainTasks, ($tasks) => {
	const total = $tasks.length;
	const completed = $tasks.filter((t) => t.status === 'completed').length;
	const running = $tasks.filter((t) => t.status === 'running').length;
	const failed = $tasks.filter((t) => t.status === 'failed').length;
	const pending = $tasks.filter((t) => t.status === 'pending' || t.status === 'queued').length;

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
export const resultSummary = derived(taskResults, ($results) => {
	const total = $results.length;
	const active = $results.filter((r) => r.httpStatus && r.httpStatus < 400).length;
	const withCNAME = $results.filter((r) => r.cname).length;
	const vulnerable = $results.filter((r) => r.takeover?.vulnerable).length;

	// 按来源统计
	const sourceStats: Record<string, number> = {};
	$results.forEach((result) => {
		sourceStats[result.source] = (sourceStats[result.source] || 0) + 1;
	});

	// 按IP统计
	const ipStats: Record<string, number> = {};
	$results.forEach((result) => {
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
export const subdomainActions = {
	/**
	 * 加载任务列表
	 */
	async loadTasks(params?: { page?: number; limit?: number; filters?: TaskFilter }) {
		loading.update((state) => ({ ...state, tasks: true }));

		try {
			const currentFilters = params?.filters || {};
			const response = await subdomainApi.getTasks({
				page: params?.page || 1,
				limit: params?.limit || 20,
				...currentFilters
			});

			subdomainTasks.set(response.tasks);
			pagination.update((state) => ({
				...state,
				page: response.page,
				limit: response.limit,
				total: response.total,
				totalPages: response.totalPages
			}));

			if (params?.filters) {
				filters.set(currentFilters);
			}
		} catch (error) {
			console.error('Failed to load subdomain tasks:', error);
			toastStore.error('加载子域名枚举任务失败: ' + (error as Error).message);
		} finally {
			loading.update((state) => ({ ...state, tasks: false }));
		}
	},

	/**
	 * 加载任务详情
	 */
	async loadTask(taskId: string) {
		loading.update((state) => ({ ...state, taskDetail: true }));

		try {
			const task = await subdomainApi.getTask(taskId);
			currentTask.set(task);
			return task;
		} catch (error) {
			console.error('Failed to load task:', error);
			toastStore.error('加载任务详情失败: ' + (error as Error).message);
			throw error;
		} finally {
			loading.update((state) => ({ ...state, taskDetail: false }));
		}
	},

	/**
	 * 加载任务结果
	 */
	async loadTaskResults(taskId: string, params?: { page?: number; limit?: number }) {
		loading.update((state) => ({ ...state, results: true }));

		try {
			const response = await subdomainApi.getTaskResults(taskId, {
				page: params?.page || 1,
				limit: params?.limit || 100
			});

			taskResults.set(response.results);
			return response;
		} catch (error) {
			console.error('Failed to load task results:', error);
			toastStore.error('加载任务结果失败: ' + (error as Error).message);
			throw error;
		} finally {
			loading.update((state) => ({ ...state, results: false }));
		}
	},

	/**
	 * 创建新任务
	 */
	async createTask(taskData: any) {
		loading.update((state) => ({ ...state, creating: true }));

		try {
			const task = await subdomainApi.createTask(taskData);

			// 添加到任务列表
			subdomainTasks.update((tasks) => [task, ...tasks]);

			toastStore.success(`子域名枚举任务 "${task.name}" 创建成功`);
			return task;
		} catch (error) {
			console.error('Failed to create task:', error);
			toastStore.error('创建任务失败: ' + (error as Error).message);
			throw error;
		} finally {
			loading.update((state) => ({ ...state, creating: false }));
		}
	},

	/**
	 * 删除任务
	 */
	async deleteTask(taskId: string) {
		try {
			await subdomainApi.deleteTask(taskId);

			// 从任务列表中移除
			subdomainTasks.update((tasks) => tasks.filter((t) => t.id !== taskId));

			// 如果是当前任务，清除
			currentTask.update((current) => (current?.id === taskId ? null : current));

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
			subdomainTasks.update((tasks) =>
				tasks.map((t) => (t.id === taskId ? { ...t, status: 'canceled' as const } : t))
			);

			currentTask.update((current) =>
				current?.id === taskId ? { ...current, status: 'canceled' as const } : current
			);

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
			subdomainTasks.update((tasks) => tasks.map((t) => (t.id === taskId ? task : t)));

			currentTask.update((current) => (current?.id === taskId ? task : current));

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
		loading.update((state) => ({ ...state, statistics: true }));

		try {
			const stats = await subdomainApi.getStatistics(params);
			statistics.set(stats);
			return stats;
		} catch (error) {
			console.error('Failed to load statistics:', error);
			toastStore.error('加载统计数据失败: ' + (error as Error).message);
		} finally {
			loading.update((state) => ({ ...state, statistics: false }));
		}
	},

	/**
	 * 更新过滤器
	 */
	updateFilters(newFilters: TaskFilter) {
		filters.set(newFilters);
	},

	/**
	 * 清除过滤器
	 */
	clearFilters() {
		filters.set({});
	},

	/**
	 * 刷新当前页面
	 */
	async refresh() {
		const currentPagination = await new Promise((resolve) => {
			const unsubscribe = pagination.subscribe(resolve);
			unsubscribe();
		});

		const currentFilters = await new Promise((resolve) => {
			const unsubscribe = filters.subscribe(resolve);
			unsubscribe();
		});

		await this.loadTasks({
			page: currentPagination.page,
			limit: currentPagination.limit,
			filters: currentFilters
		});
	},

	/**
	 * 处理实时进度更新
	 */
	handleProgressUpdate(update: ProgressUpdateEvent) {
		// 更新进度缓存
		progressUpdates.update((updates) => ({
			...updates,
			[update.taskId]: update
		}));

		// 更新任务列表中的任务状态
		subdomainTasks.update((tasks) =>
			tasks.map((task) => {
				if (task.id === update.taskId) {
					return {
						...task,
						progress: update.progress,
						status: update.status as any
					};
				}
				return task;
			})
		);

		// 更新当前任务状态
		currentTask.update((current) => {
			if (current?.id === update.taskId) {
				return {
					...current,
					progress: update.progress,
					status: update.status as any
				};
			}
			return current;
		});

		// 如果有最新结果，更新结果列表
		if (update.latestResults && update.latestResults.length > 0) {
			taskResults.update((results) => {
				// 合并新结果，避免重复
				const existingIds = new Set(results.map((r) => r.subdomain));
				const newResults = update.latestResults!.filter((r) => !existingIds.has(r.subdomain));
				return [...results, ...newResults];
			});
		}
	},

	/**
	 * 清除进度更新缓存
	 */
	clearProgressUpdates() {
		progressUpdates.set({});
	},

	/**
	 * 重置所有状态
	 */
	reset() {
		subdomainTasks.set([]);
		currentTask.set(null);
		taskResults.set([]);
		statistics.set(null);
		filters.set({});
		progressUpdates.set({});
		pagination.set({
			page: 1,
			limit: 20,
			total: 0,
			totalPages: 0
		});
		loading.set({
			tasks: false,
			taskDetail: false,
			results: false,
			creating: false,
			statistics: false
		});
	},

	/**
	 * 批量操作：删除多个任务
	 */
	async batchDeleteTasks(taskIds: string[]) {
		const results = {
			success: 0,
			failed: 0,
			errors: [] as string[]
		};

		for (const taskId of taskIds) {
			try {
				await subdomainApi.deleteTask(taskId);
				results.success++;
			} catch (error) {
				results.failed++;
				results.errors.push(`删除任务 ${taskId} 失败: ${(error as Error).message}`);
			}
		}

		// 刷新任务列表
		await this.refresh();

		if (results.success > 0) {
			toastStore.success(`成功删除 ${results.success} 个任务`);
		}
		if (results.failed > 0) {
			toastStore.error(`删除失败 ${results.failed} 个任务`);
		}

		return results;
	},

	/**
	 * 获取任务结果的唯一子域名列表
	 */
	getUniqueSubdomains(results: SubdomainResult[]): SubdomainResult[] {
		const seen = new Set<string>();
		return results.filter((result) => {
			if (seen.has(result.subdomain)) {
				return false;
			}
			seen.add(result.subdomain);
			return true;
		});
	},

	/**
	 * 按来源分组结果
	 */
	groupResultsBySource(results: SubdomainResult[]): Record<string, SubdomainResult[]> {
		return results.reduce(
			(groups, result) => {
				const source = result.source;
				if (!groups[source]) {
					groups[source] = [];
				}
				groups[source].push(result);
				return groups;
			},
			{} as Record<string, SubdomainResult[]>
		);
	},

	/**
	 * 过滤活跃的子域名
	 */
	filterActiveSubdomains(results: SubdomainResult[]): SubdomainResult[] {
		return results.filter(
			(result) => result.httpStatus && result.httpStatus >= 200 && result.httpStatus < 500
		);
	},

	/**
	 * 查找可能的子域名接管
	 */
	findPotentialTakeovers(results: SubdomainResult[]): SubdomainResult[] {
		return results.filter(
			(result) => result.takeover?.vulnerable || (result.cname && !result.ips.length)
		);
	}
};

// 导出常用功能
export { subdomainActions as actions };
export default {
	tasks: subdomainTasks,
	currentTask,
	results: taskResults,
	loading,
	pagination,
	filters,
	statistics,
	progressUpdates,
	runningTasks,
	recentCompletedTasks,
	taskSummary,
	resultSummary,
	actions: subdomainActions
};

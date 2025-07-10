/**
 * 端口扫描状态管理Store
 * 使用Svelte 5 runes语法管理端口扫描相关状态
 */

import { writable, derived } from 'svelte/store';
import type {
	PortScanTask,
	PortScanResult,
	TaskFilter,
	PortScanStatistics,
	ProgressUpdateEvent
} from '$lib/types/portscan';
import { portScanApi } from '$lib/api/portscan';
import { toastStore } from '$lib/stores/toast';

// 任务列表状态
export const portScanTasks = writable<PortScanTask[]>([]);
export const currentTask = writable<PortScanTask | null>(null);
export const taskResults = writable<PortScanResult[]>([]);

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
export const statistics = writable<PortScanStatistics | null>(null);

// 实时进度更新
export const progressUpdates = writable<Record<string, ProgressUpdateEvent>>({});

// 派生状态：运行中的任务
export const runningTasks = derived(portScanTasks, ($tasks) =>
	$tasks.filter((task) => task.status === 'running' || task.status === 'queued')
);

// 派生状态：最近完成的任务
export const recentCompletedTasks = derived(portScanTasks, ($tasks) =>
	$tasks
		.filter((task) => task.status === 'completed')
		.sort(
			(a, b) =>
				new Date(b.endTime || b.updatedAt).getTime() - new Date(a.endTime || a.updatedAt).getTime()
		)
		.slice(0, 5)
);

// 派生状态：任务统计摘要
export const taskSummary = derived(portScanTasks, ($tasks) => {
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

/**
 * 端口扫描操作函数
 */
export const portScanActions = {
	/**
	 * 加载任务列表
	 */
	async loadTasks(params?: { page?: number; limit?: number; filters?: TaskFilter }) {
		loading.update((state) => ({ ...state, tasks: true }));

		try {
			const currentFilters = params?.filters || {};
			const response = await portScanApi.getTasks({
				page: params?.page || 1,
				limit: params?.limit || 20,
				...currentFilters
			});

			portScanTasks.set(response.tasks);
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
			console.error('Failed to load port scan tasks:', error);
			toastStore.error('加载端口扫描任务失败: ' + (error as Error).message);
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
			const task = await portScanApi.getTask(taskId);
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
			const response = await portScanApi.getTaskResults(taskId, {
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
			const task = await portScanApi.createTask(taskData);

			// 添加到任务列表
			portScanTasks.update((tasks) => [task, ...tasks]);

			toastStore.success(`端口扫描任务 "${task.name}" 创建成功`);
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
			await portScanApi.deleteTask(taskId);

			// 从任务列表中移除
			portScanTasks.update((tasks) => tasks.filter((t) => t.id !== taskId));

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
			await portScanApi.cancelTask(taskId);

			// 更新任务状态
			portScanTasks.update((tasks) =>
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
			const task = await portScanApi.retryTask(taskId);

			// 更新任务列表
			portScanTasks.update((tasks) => tasks.map((t) => (t.id === taskId ? task : t)));

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
			const result = await portScanApi.exportResults(taskId, format);

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
			const stats = await portScanApi.getStatistics(params);
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
		portScanTasks.update((tasks) =>
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
				const existingIds = new Set(results.map((r) => `${r.host}:${r.port}`));
				const newResults = update.latestResults!.filter(
					(r) => !existingIds.has(`${r.host}:${r.port}`)
				);
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
		portScanTasks.set([]);
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
	}
};

// 导出常用功能
export { portScanActions as actions };
export default {
	tasks: portScanTasks,
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
	actions: portScanActions
};

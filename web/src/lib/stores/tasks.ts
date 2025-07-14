import { taskApi } from '$lib/api/tasks';
import { notifications } from '$lib/stores/notifications';
import { writable, derived, get } from 'svelte/store';
import type {
	Task,
	TaskResult,
	TaskEvent,
	TaskLog,
	TaskStats,
	TaskTemplate,
	TaskScheduleRule,
	TaskQueryParams,
	CreateTaskRequest,
	UpdateTaskRequest,
	TaskStatus,
	TaskType,
	TaskPriority,
	ExecutorInfo,
	TaskListResult
} from '$lib/types/task';

// 基础状态stores
export const tasks = writable<Task[]>([]);
export const selectedTask = writable<Task | null>(null);
export const taskResult = writable<TaskResult | null>(null);
export const taskEvents = writable<TaskEvent[]>([]);
export const taskStats = writable<TaskStats | null>(null);
export const templates = writable<TaskTemplate[]>([]);
export const scheduleRules = writable<TaskScheduleRule[]>([]);
export const executors = writable<Record<string, ExecutorInfo>>({});
export const runningTasks = writable<string[]>([]);
export const loading = writable<boolean>(false);
export const error = writable<string | null>(null);
export const pagination = writable({
	page: 1,
	pageSize: 20,
	total: 0,
	totalPages: 0
});
export const filters = writable<TaskQueryParams>({
	page: 1,
	pageSize: 20
});

// Derived stores
export const isTaskLoading = derived(loading, ($loading) => $loading);
export const taskError = derived(error, ($error) => $error);
export const taskCount = derived(tasks, ($tasks) => $tasks.length);
export const hasSelectedTask = derived(selectedTask, ($selectedTask) => $selectedTask !== null);
export const runningTaskCount = derived(runningTasks, ($runningTasks) => $runningTasks.length);

export const tasksByStatus = derived(tasks, ($tasks) => {
	const statusGroups: Record<TaskStatus, Task[]> = {
		pending: [],
		queued: [],
		running: [],
		completed: [],
		failed: [],
		cancelled: [], // 修正拼写
		timeout: []
	};

	$tasks.forEach((task) => {
		if (statusGroups[task.status as TaskStatus]) {
			statusGroups[task.status as TaskStatus].push(task);
		}
	});

	return statusGroups;
});

export const tasksByType = derived(tasks, ($tasks) => {
	const typeGroups: Record<TaskType, Task[]> = {
		subdomain_enum: [],
		port_scan: [],
		vuln_scan: [],
		asset_discovery: [],
		dir_scan: [],
		web_crawl: [],
		sensitive_scan: [],
		page_monitor: []
	};

	$tasks.forEach((task) => {
		if (typeGroups[task.type as TaskType]) {
			typeGroups[task.type as TaskType].push(task);
		}
	});

	return typeGroups;
});

export const tasksByPriority = derived(tasks, ($tasks) => {
	const priorityGroups: Record<string, Task[]> = {
		'1': [], // low
		'2': [], // normal
		'3': [], // high
		'4': [] // critical
	};

	$tasks.forEach((task) => {
		const priority = task.priority?.toString() || '2';
		if (priorityGroups[priority]) {
			priorityGroups[priority].push(task);
		}
	});

	return priorityGroups;
});

// 创建任务store的函数
function createTaskStore() {
	return {
		// 基础状态管理
		setLoading: (isLoading: boolean) => loading.set(isLoading),
		setError: (errorMessage: string | null) => {
			error.set(errorMessage);
			if (errorMessage) {
				notifications.add({
					type: 'error',
					message: errorMessage
				});
			}
		},
		reset: () => {
			tasks.set([]);
			selectedTask.set(null);
			taskResult.set(null);
			taskEvents.set([]);
			taskStats.set(null);
			templates.set([]);
			scheduleRules.set([]);
			executors.set({});
			runningTasks.set([]);
			loading.set(false);
			error.set(null);
			pagination.set({ page: 1, pageSize: 20, total: 0, totalPages: 0 });
			filters.set({ page: 1, pageSize: 20 });
		},

		// Store 状态访问器
		get tasks() {
			return get(tasks);
		},
		get selectedTask() {
			return get(selectedTask);
		},
		get taskResult() {
			return get(taskResult);
		},
		get taskEvents() {
			return get(taskEvents);
		},
		get taskStats() {
			return get(taskStats);
		},
		get templates() {
			return get(templates);
		},
		get scheduleRules() {
			return get(scheduleRules);
		},
		get executors() {
			return get(executors);
		},
		get runningTasks() {
			return get(runningTasks);
		},
		get loading() {
			return get(loading);
		},
		get error() {
			return get(error);
		},
		get pagination() {
			return get(pagination);
		},
		get filters() {
			return get(filters);
		},

		// 任务管理
		loadTasks: async (params?: TaskQueryParams) => {
			loading.set(true);
			error.set(null);

			try {
				const currentFilters = get(filters);
				const queryParams = { ...currentFilters, ...params };
				const response = await taskApi.getTasks(queryParams);

				if (response.code === 200) {
					tasks.set(response.data.items);
					pagination.set({
						page: response.data.page,
						pageSize: response.data.pageSize,
						total: response.data.total,
						totalPages: response.data.totalPages
					});
					filters.set(queryParams);
				} else {
					error.set(response.message || '加载任务列表失败');
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '加载任务列表失败');
			} finally {
				loading.set(false);
			}
		},

		// 加载任务统计
		loadTaskStats: async (): Promise<void> => {
			try {
				const response = await taskApi.getTaskStats();
				if (response.code === 200) {
					taskStats.set(response.data);
				} else {
					error.set(response.message || '加载任务统计失败');
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '加载任务统计失败');
			}
		},

		createTask: async (data: CreateTaskRequest): Promise<Task | null> => {
			loading.set(true);
			error.set(null);

			try {
				const response = await taskApi.createTask(data);

				if (response.code === 200) {
					tasks.update((currentTasks) => [response.data, ...currentTasks]);
					notifications.add({
						type: 'success',
						message: '任务创建成功'
					});
					return response.data;
				} else {
					error.set(response.message || '创建任务失败');
					return null;
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '创建任务失败');
				return null;
			} finally {
				loading.set(false);
			}
		},

		updateTask: async (taskId: string, data: UpdateTaskRequest): Promise<Task | null> => {
			loading.set(true);
			error.set(null);

			try {
				const response = await taskApi.updateTask(taskId, data);

				if (response.code === 200) {
					tasks.update((currentTasks) =>
						currentTasks.map((task) => (task.id === taskId ? response.data : task))
					);
					notifications.add({
						type: 'success',
						message: '任务更新成功'
					});
					return response.data;
				} else {
					error.set(response.message || '更新任务失败');
					return null;
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '更新任务失败');
				return null;
			} finally {
				loading.set(false);
			}
		},

		deleteTask: async (taskId: string): Promise<boolean> => {
			loading.set(true);
			error.set(null);

			try {
				const response = await taskApi.deleteTask(taskId);

				if (response.code === 200) {
					tasks.update((currentTasks) => currentTasks.filter((task) => task.id !== taskId));
					notifications.add({
						type: 'success',
						message: '任务删除成功'
					});
					return true;
				} else {
					error.set(response.message || '删除任务失败');
					return false;
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '删除任务失败');
				return false;
			} finally {
				loading.set(false);
			}
		},

		// 克隆任务
		cloneTask: async (taskId: string): Promise<boolean> => {
			loading.set(true);
			error.set(null);

			try {
				const response = await taskApi.cloneTask(taskId);

				if (response.code === 200) {
					// 重新加载任务列表
					await taskStore.loadTasks();
					notifications.add({
						type: 'success',
						message: '任务克隆成功'
					});
					return true;
				} else {
					error.set(response.message || '克隆任务失败');
					return false;
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '克隆任务失败');
				return false;
			} finally {
				loading.set(false);
			}
		},

		// 下载任务结果
		downloadTaskResult: async (
			taskId: string,
			format: 'json' | 'csv' | 'xml' = 'json'
		): Promise<void> => {
			try {
				await taskApi.downloadTaskResult(taskId, format);
				notifications.add({
					type: 'success',
					message: `任务结果已下载 (${format.toUpperCase()})`
				});
			} catch (err) {
				error.set(err instanceof Error ? err.message : '下载任务结果失败');
			}
		},

		// 获取任务配置模板
		getTaskConfigTemplate: async (type: string): Promise<Record<string, any> | null> => {
			try {
				const response = await taskApi.getTaskConfigTemplate(type);

				if (response.code === 200) {
					return response.data;
				} else {
					error.set(response.message || '获取配置模板失败');
					return null;
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '获取配置模板失败');
				return null;
			}
		},

		// 验证任务配置
		validateTaskConfig: async (
			type: string,
			config: Record<string, any>
		): Promise<{ valid: boolean; errors?: string[] }> => {
			try {
				const response = await taskApi.validateTaskConfig(type, config);

				if (response.code === 200) {
					return {
						valid: response.data.valid,
						errors: response.data.errors?.map((err) => err.message) || []
					};
				} else {
					error.set(response.message || '验证配置失败');
					return { valid: false, errors: [response.message || '验证失败'] };
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '验证配置失败');
				return { valid: false, errors: [err instanceof Error ? err.message : '验证失败'] };
			}
		},

		// 取消任务
		cancelTask: async (taskId: string): Promise<boolean> => {
			loading.set(true);
			error.set(null);

			try {
				const response = await taskApi.cancelTask(taskId);

				if (response.code === 200) {
					// 更新任务状态
					tasks.update((currentTasks) =>
						currentTasks.map((task) =>
							task.id === taskId ? { ...task, status: 'cancelled' } : task
						)
					);
					notifications.add({
						type: 'success',
						message: '任务已取消'
					});
					return true;
				} else {
					error.set(response.message || '取消任务失败');
					return false;
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '取消任务失败');
				return false;
			} finally {
				loading.set(false);
			}
		},

		// 重启任务
		restartTask: async (taskId: string): Promise<boolean> => {
			loading.set(true);
			error.set(null);

			try {
				const response = await taskApi.restartTask(taskId);

				if (response.code === 200) {
					// 重新加载任务列表
					await taskStore.loadTasks();
					notifications.add({
						type: 'success',
						message: '任务已重启'
					});
					return true;
				} else {
					error.set(response.message || '重启任务失败');
					return false;
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '重启任务失败');
				return false;
			} finally {
				loading.set(false);
			}
		},

		// 获取任务日志流
		getTaskLogStream: (
			taskId: string,
			onMessage: (log: TaskLog) => void,
			onError?: (error: Error) => void
		): EventSource | null => {
			try {
				return taskApi.getTaskLogStream(taskId, onMessage, onError);
			} catch (err) {
				error.set(err instanceof Error ? err.message : '获取日志流失败');
				return null;
			}
		},

		// 批量操作
		batchDelete: async (taskIds: string[]): Promise<boolean> => {
			loading.set(true);
			error.set(null);

			try {
				const promises = taskIds.map((id) => taskApi.deleteTask(id));
				const results = await Promise.all(promises);

				const successCount = results.filter((r) => r.code === 200).length;

				if (successCount > 0) {
					// 重新加载任务列表
					await taskStore.loadTasks();
					notifications.add({
						type: 'success',
						message: `成功删除 ${successCount} 个任务`
					});
				}

				if (successCount < taskIds.length) {
					error.set(`${taskIds.length - successCount} 个任务删除失败`);
				}

				return successCount === taskIds.length;
			} catch (err) {
				error.set(err instanceof Error ? err.message : '批量删除失败');
				return false;
			} finally {
				loading.set(false);
			}
		},

		batchCancel: async (taskIds: string[]): Promise<boolean> => {
			loading.set(true);
			error.set(null);

			try {
				const promises = taskIds.map((id) => taskApi.cancelTask(id));
				const results = await Promise.all(promises);

				const successCount = results.filter((r) => r.code === 200).length;

				if (successCount > 0) {
					// 更新任务状态
					tasks.update((currentTasks) =>
						currentTasks.map((task) =>
							taskIds.includes(task.id) ? { ...task, status: 'cancelled' } : task
						)
					);
					notifications.add({
						type: 'success',
						message: `成功取消 ${successCount} 个任务`
					});
				}

				if (successCount < taskIds.length) {
					error.set(`${taskIds.length - successCount} 个任务取消失败`);
				}

				return successCount === taskIds.length;
			} catch (err) {
				error.set(err instanceof Error ? err.message : '批量取消失败');
				return false;
			} finally {
				loading.set(false);
			}
		},

		// 生成任务执行计划
		generateTaskExecutionPlan: async (
			taskData: CreateTaskRequest
		): Promise<Record<string, any> | null> => {
			try {
				const response = await taskApi.generateTaskExecutionPlan(taskData);

				if (response.code === 200) {
					return response.data;
				} else {
					error.set(response.message || '生成执行计划失败');
					return null;
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '生成执行计划失败');
				return null;
			}
		},

		// 任务选择
		selectTask: async (taskId: string) => {
			try {
				const response = await taskApi.getTask(taskId);
				if (response.code === 200) {
					selectedTask.set(response.data);
				} else {
					error.set(response.message || '获取任务详情失败');
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '获取任务详情失败');
			}
		},

		// 任务操作
		startTask: async (taskId: string): Promise<boolean> => {
			try {
				const response = await taskApi.startTask(taskId);
				if (response.code === 200) {
					tasks.update((currentTasks) =>
						currentTasks.map((task) => (task.id === taskId ? { ...task, status: 'running' } : task))
					);
					runningTasks.update((current) => [...current, taskId]);
					return true;
				}
				return false;
			} catch {
				return false;
			}
		},

		stopTask: async (taskId: string): Promise<boolean> => {
			try {
				const response = await taskApi.stopTask(taskId);
				if (response.code === 200) {
					tasks.update((currentTasks) =>
						currentTasks.map((task) =>
							task.id === taskId ? { ...task, status: 'cancelled' } : task
						)
					);
					runningTasks.update((current) => current.filter((id) => id !== taskId));
					return true;
				}
				return false;
			} catch {
				return false;
			}
		},

		// 加载任务结果
		loadTaskResult: async (taskId: string): Promise<void> => {
			try {
				const response = await taskApi.getTaskResult(taskId);
				if (response.code === 200) {
					taskResult.set(response.data);
				} else {
					error.set(response.message || '获取任务结果失败');
				}
			} catch (err) {
				error.set(err instanceof Error ? err.message : '获取任务结果失败');
			}
		},

		// 获取任务事件流
		getTaskEventStream: (
			taskId: string,
			onEvent: (event: any) => void,
			onError?: (error: Error) => void
		): EventSource => {
			return taskApi.getTaskEventStream(taskId, onEvent, onError);
		}
	};
}

export const taskStore = createTaskStore();

// 导出taskActions别名
export const taskActions = taskStore;

// 导出store状态供组件使用
export {
	tasks as taskList,
	selectedTask as currentTask,
	loading as taskLoading,
	error as taskErrorMessage,
	pagination as taskPagination
};

import { writable, derived, get } from 'svelte/store';
import { taskApi } from '$lib/api/tasks';
import api from '$lib/api/axios-config';
import { notifications } from '$lib/stores/notifications';
import type {
	Task,
	TaskResult,
	TaskEvent,
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

// 任务管理状态
interface TaskState {
	tasks: Task[];
	selectedTask: Task | null;
	taskResult: TaskResult | null;
	taskEvents: TaskEvent[];
	taskStats: TaskStats | null;
	templates: TaskTemplate[];
	scheduleRules: TaskScheduleRule[];
	executors: Record<string, ExecutorInfo>;
	runningTasks: string[];
	loading: boolean;
	error: string | null;
	pagination: {
		page: number;
		pageSize: number;
		total: number;
		totalPages: number;
	};
	filters: TaskQueryParams;
}

// 初始状态
const initialState: TaskState = {
	tasks: [],
	selectedTask: null,
	taskResult: null,
	taskEvents: [],
	taskStats: null,
	templates: [],
	scheduleRules: [],
	executors: {},
	runningTasks: [],
	loading: false,
	error: null,
	pagination: {
		page: 1,
		pageSize: 20,
		total: 0,
		totalPages: 0
	},
	filters: {
		page: 1,
		pageSize: 20
	}
};

// 创建任务store
function createTaskStore() {
	const { subscribe, set, update } = writable<TaskState>(initialState);

	return {
		subscribe,

		// ==================== 状态管理 ====================

		/**
		 * 设置加载状态
		 */
		setLoading: (loading: boolean) => {
			update((state) => ({ ...state, loading }));
		},

		/**
		 * 设置错误状态
		 */
		setError: (error: string | null) => {
			update((state) => ({ ...state, error }));
			if (error) {
				notifications.add({
					type: 'error',
					message: error
				});
			}
		},

		/**
		 * 重置状态
		 */
		reset: () => {
			set(initialState);
		},

		// ==================== 任务管理 ====================

		/**
		 * 加载任务列表
		 */
		loadTasks: async (params?: TaskQueryParams) => {
			update((state) => ({ ...state, loading: true, error: null }));

			try {
				const queryParams = { ...get({ subscribe }).filters, ...params };
				const response = await taskApi.getTasks(queryParams);

				if (response.code === 200) {
					update((state) => ({
						...state,
						tasks: response.data.items,
						pagination: {
							page: response.data.page,
							pageSize: response.data.pageSize,
							total: response.data.total,
							totalPages: response.data.totalPages
						},
						filters: queryParams,
						loading: false
					}));
				} else {
					update((state) => ({
						...state,
						error: response.message || '加载任务列表失败',
						loading: false
					}));
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '加载任务列表失败',
					loading: false
				}));
			}
		},

		/**
		 * 创建任务
		 */
		createTask: async (data: CreateTaskRequest): Promise<Task | null> => {
			update((state) => ({ ...state, loading: true, error: null }));

			try {
				const response = await taskApi.createTask(data);

				if (response.code === 200) {
					update((state) => ({
						...state,
						tasks: [response.data, ...state.tasks],
						loading: false
					}));

					notifications.add({
						type: 'success',
						message: '任务创建成功'
					});

					return response.data;
				} else {
					update((state) => ({
						...state,
						error: response.message || '创建任务失败',
						loading: false
					}));
					return null;
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '创建任务失败',
					loading: false
				}));
				return null;
			}
		},

		/**
		 * 更新任务
		 */
		updateTask: async (taskId: string, data: UpdateTaskRequest): Promise<boolean> => {
			update((state) => ({ ...state, loading: true, error: null }));

			try {
				const response = await taskApi.updateTask(taskId, data);

				if (response.code === 200) {
					update((state) => ({
						...state,
						tasks: state.tasks.map((task) => (task.id === taskId ? response.data : task)),
						selectedTask: state.selectedTask?.id === taskId ? response.data : state.selectedTask,
						loading: false
					}));

					notifications.add({
						type: 'success',
						message: '任务更新成功'
					});

					return true;
				} else {
					update((state) => ({
						...state,
						error: response.message || '更新任务失败',
						loading: false
					}));
					return false;
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '更新任务失败',
					loading: false
				}));
				return false;
			}
		},

		/**
		 * 删除任务
		 */
		deleteTask: async (taskId: string): Promise<boolean> => {
			update((state) => ({ ...state, loading: true, error: null }));

			try {
				const response = await taskApi.deleteTask(taskId);

				if (response.code === 200) {
					update((state) => ({
						...state,
						tasks: state.tasks.filter((task) => task.id !== taskId),
						selectedTask: state.selectedTask?.id === taskId ? null : state.selectedTask,
						loading: false
					}));

					notifications.add({
						type: 'success',
						message: '任务删除成功'
					});

					return true;
				} else {
					update((state) => ({
						...state,
						error: response.message || '删除任务失败',
						loading: false
					}));
					return false;
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '删除任务失败',
					loading: false
				}));
				return false;
			}
		},

		/**
		 * 选择任务
		 */
		selectTask: async (taskId: string) => {
			update((state) => ({ ...state, loading: true, error: null }));

			try {
				const response = await taskApi.getTask(taskId);

				if (response.code === 200) {
					update((state) => ({
						...state,
						selectedTask: response.data,
						loading: false
					}));
				} else {
					update((state) => ({
						...state,
						error: response.message || '获取任务详情失败',
						loading: false
					}));
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '获取任务详情失败',
					loading: false
				}));
			}
		},

		/**
		 * 清除选择的任务
		 */
		clearSelectedTask: () => {
			update((state) => ({ ...state, selectedTask: null }));
		},

		// ==================== 任务操作 ====================

		/**
		 * 启动任务
		 */
		startTask: async (taskId: string): Promise<boolean> => {
			try {
				const response = await taskApi.startTask(taskId);

				if (response.code === 200) {
					update((state) => ({
						...state,
						tasks: state.tasks.map((task) =>
							task.id === taskId ? { ...task, status: 'running' as TaskStatus } : task
						)
					}));

					notifications.add({
						type: 'success',
						message: '任务启动成功'
					});

					return true;
				} else {
					update((state) => ({
						...state,
						error: response.message || '启动任务失败'
					}));
					return false;
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '启动任务失败'
				}));
				return false;
			}
		},

		/**
		 * 取消任务
		 */
		cancelTask: async (taskId: string): Promise<boolean> => {
			try {
				const response = await taskApi.cancelTask(taskId);

				if (response.code === 200) {
					update((state) => ({
						...state,
						tasks: state.tasks.map((task) =>
							task.id === taskId ? { ...task, status: 'canceled' as TaskStatus } : task
						)
					}));

					notifications.add({
						type: 'success',
						message: '任务取消成功'
					});

					return true;
				} else {
					update((state) => ({
						...state,
						error: response.message || '取消任务失败'
					}));
					return false;
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '取消任务失败'
				}));
				return false;
			}
		},

		/**
		 * 重启任务
		 */
		restartTask: async (taskId: string): Promise<boolean> => {
			try {
				const response = await taskApi.restartTask(taskId);

				if (response.code === 200) {
					update((state) => ({
						...state,
						tasks: state.tasks.map((task) =>
							task.id === taskId ? { ...task, status: 'pending' as TaskStatus } : task
						)
					}));

					notifications.add({
						type: 'success',
						message: '任务重启成功'
					});

					return true;
				} else {
					update((state) => ({
						...state,
						error: response.message || '重启任务失败'
					}));
					return false;
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '重启任务失败'
				}));
				return false;
			}
		},

		/**
		 * 克隆任务
		 */
		cloneTask: async (taskId: string): Promise<boolean> => {
			try {
				const response = await taskApi.cloneTask(taskId);

				if (response.code === 200) {
					update((state) => ({
						...state,
						tasks: [response.data, ...state.tasks]
					}));

					notifications.add({
						type: 'success',
						message: '任务克隆成功'
					});

					return true;
				} else {
					update((state) => ({
						...state,
						error: response.message || '克隆任务失败'
					}));
					return false;
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '克隆任务失败'
				}));
				return false;
			}
		},

		// ==================== 任务结果 ====================

		/**
		 * 加载任务结果
		 */
		loadTaskResult: async (taskId: string) => {
			update((state) => ({ ...state, loading: true, error: null }));

			try {
				const response = await taskApi.getTaskResult(taskId);

				if (response.code === 200) {
					update((state) => ({
						...state,
						taskResult: response.data,
						loading: false
					}));
				} else {
					update((state) => ({
						...state,
						error: response.message || '获取任务结果失败',
						loading: false
					}));
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '获取任务结果失败',
					loading: false
				}));
			}
		},

		/**
		 * 下载任务结果
		 */
		downloadTaskResult: async (taskId: string, format: 'json' | 'csv' | 'xml' = 'json') => {
			try {
				await taskApi.downloadTaskResult(taskId, format);
				notifications.add({
					type: 'success',
					message: '任务结果下载成功'
				});
			} catch (error) {
				notifications.add({
					type: 'error',
					message: error instanceof Error ? error.message : '下载任务结果失败'
				});
			}
		},

		// ==================== 任务统计 ====================

		/**
		 * 加载任务统计
		 */
		loadTaskStats: async (projectId?: string) => {
			try {
				const response = await taskApi.getTaskStats(projectId);

				if (response.code === 200) {
					update((state) => ({
						...state,
						taskStats: response.data
					}));
				} else {
					update((state) => ({
						...state,
						error: response.message || '获取任务统计失败'
					}));
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '获取任务统计失败'
				}));
			}
		},

		// ==================== 执行器管理 ====================

		/**
		 * 加载执行器信息
		 */
		loadExecutors: async () => {
			try {
				const response = await taskApi.getExecutors();

				if (response.code === 200) {
					update((state) => ({
						...state,
						executors: response.data
					}));
				} else {
					update((state) => ({
						...state,
						error: response.message || '获取执行器信息失败'
					}));
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '获取执行器信息失败'
				}));
			}
		},

		/**
		 * 加载运行中的任务
		 */
		loadRunningTasks: async () => {
			try {
				const response = await taskApi.getRunningTasks();

				if (response.code === 200) {
					update((state) => ({
						...state,
						runningTasks: response.data
					}));
				} else {
					update((state) => ({
						...state,
						error: response.message || '获取运行中任务失败'
					}));
				}
			} catch (error) {
				update((state) => ({
					...state,
					error: error instanceof Error ? error.message : '获取运行中任务失败'
				}));
			}
		},

		// ==================== 过滤器 ====================

		/**
		 * 更新过滤器
		 */
		updateFilters: (filters: Partial<TaskQueryParams>) => {
			update((state) => ({
				...state,
				filters: { ...state.filters, ...filters }
			}));
		},

		/**
		 * 重置过滤器
		 */
		resetFilters: () => {
			update((state) => ({
				...state,
				filters: {
					page: 1,
					pageSize: 20
				}
			}));
		},

		// ==================== 分页 ====================

		/**
		 * 设置页码
		 */
		setPage: (page: number) => {
			update((state) => ({
				...state,
				pagination: { ...state.pagination, page },
				filters: { ...state.filters, page }
			}));
		},

		/**
		 * 设置每页大小
		 */
		setPageSize: (pageSize: number) => {
			update((state) => ({
				...state,
				pagination: { ...state.pagination, pageSize, page: 1 },
				filters: { ...state.filters, pageSize, page: 1 }
			}));
		}
	};
}

// 创建任务store实例
export const taskStore = createTaskStore();

// 派生状态
export const isTaskLoading = derived(taskStore, ($taskStore) => $taskStore.loading);
export const taskError = derived(taskStore, ($taskStore) => $taskStore.error);
export const tasks = derived(taskStore, ($taskStore) => $taskStore.tasks);
export const selectedTask = derived(taskStore, ($taskStore) => $taskStore.selectedTask);
export const taskResult = derived(taskStore, ($taskStore) => $taskStore.taskResult);
export const taskStats = derived(taskStore, ($taskStore) => $taskStore.taskStats);
export const executors = derived(taskStore, ($taskStore) => $taskStore.executors);
export const runningTasks = derived(taskStore, ($taskStore) => $taskStore.runningTasks);
export const taskPagination = derived(taskStore, ($taskStore) => $taskStore.pagination);
export const taskFilters = derived(taskStore, ($taskStore) => $taskStore.filters);

// 计算派生状态
export const tasksByStatus = derived(tasks, ($tasks) => {
	const grouped: Record<TaskStatus, Task[]> = {
		pending: [],
		queued: [],
		running: [],
		completed: [],
		failed: [],
		canceled: [],
		timeout: []
	};

	$tasks.forEach((task) => {
		grouped[task.status].push(task);
	});

	return grouped;
});

export const tasksByType = derived(tasks, ($tasks) => {
	const grouped: Record<TaskType, Task[]> = {
		subdomain_enum: [],
		port_scan: [],
		vuln_scan: [],
		asset_discovery: [],
		dir_scan: [],
		web_crawl: []
	};

	$tasks.forEach((task) => {
		grouped[task.type].push(task);
	});

	return grouped;
});

export const tasksByPriority = derived(tasks, ($tasks) => {
	const grouped: Record<TaskPriority, Task[]> = {
		low: [],
		normal: [],
		high: [],
		critical: []
	};

	$tasks.forEach((task) => {
		grouped[task.priority].push(task);
	});

	return grouped;
});

// 任务操作actions
export const taskActions = {
	...taskStore,

	// 批量操作
	batchStart: async (taskIds: string[]) => {
		const promises = taskIds.map((id) => taskStore.startTask(id));
		const results = await Promise.allSettled(promises);
		const successCount = results.filter((r) => r.status === 'fulfilled' && r.value).length;

		notifications.add({
			type: successCount === taskIds.length ? 'success' : 'warning',
			message: `批量启动任务：成功 ${successCount}/${taskIds.length}`
		});
	},

	batchCancel: async (taskIds: string[]) => {
		const promises = taskIds.map((id) => taskStore.cancelTask(id));
		const results = await Promise.allSettled(promises);
		const successCount = results.filter((r) => r.status === 'fulfilled' && r.value).length;

		notifications.add({
			type: successCount === taskIds.length ? 'success' : 'warning',
			message: `批量取消任务：成功 ${successCount}/${taskIds.length}`
		});
	},

	batchDelete: async (taskIds: string[]) => {
		const promises = taskIds.map((id) => taskStore.deleteTask(id));
		const results = await Promise.allSettled(promises);
		const successCount = results.filter((r) => r.status === 'fulfilled' && r.value).length;

		notifications.add({
			type: successCount === taskIds.length ? 'success' : 'warning',
			message: `批量删除任务：成功 ${successCount}/${taskIds.length}`
		});
	},

	// 获取任务配置模板
	getTaskConfigTemplate: async (taskType: string) => {
		try {
			const response = await api.get(`/api/v1/tasks/templates/${taskType}`);
			return response.data;
		} catch (error) {
			console.error('获取任务配置模板失败:', error);
			return null;
		}
	},

	// 验证任务配置
	validateTaskConfig: async (taskType: string, config: any) => {
		try {
			const response = await api.post(`/api/v1/tasks/validate/${taskType}`, config);
			return response.data;
		} catch (error) {
			console.error('验证任务配置失败:', error);
			return { valid: false, errors: [] };
		}
	},

	// 生成任务执行计划
	generateTaskExecutionPlan: async (taskData: any) => {
		try {
			const response = await api.post('/api/v1/tasks/plan', taskData);
			return response.data;
		} catch (error) {
			console.error('生成任务执行计划失败:', error);
			return null;
		}
	}
};

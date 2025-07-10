import api from './axios-config';

// 创建专门用于任务API的axios实例，因为任务API在 /api 而不是 /api/v1
const taskApiClient = api.create({
	baseURL: '/api',
	timeout: 30000,
	headers: {
		'Content-Type': 'application/json'
	}
});

// 使用相同的拦截器
taskApiClient.interceptors.request = api.interceptors.request;
taskApiClient.interceptors.response = api.interceptors.response;
import type {
	Task,
	TaskResult,
	TaskEvent,
	TaskQueue,
	TaskTemplate,
	TaskScheduleRule,
	TaskLog,
	TaskStats,
	ExecutorInfo,
	ExecutionContext,
	TaskPerformanceMetrics,
	TaskNotificationConfig,
	CreateTaskRequest,
	UpdateTaskRequest,
	UpdateTaskStatusRequest,
	CreateTaskTemplateRequest,
	CreateTaskScheduleRuleRequest,
	TaskQueryParams,
	TaskLogQueryParams,
	TaskConfigValidation,
	TaskExecutionPlan,
	TaskImportRequest,
	TaskImportResult,
	TaskExport,
	BatchTaskRequest,
	BatchTaskResult,
	TaskResponse,
	TaskListResponse,
	TaskResultResponse,
	TaskStatsResponse,
	ExecutorInfoResponse,
	TaskEventListResponse,
	TaskLogListResponse,
	TaskTemplateListResponse,
	TaskScheduleRuleListResponse
} from '$lib/types/task';
import type { APIResponse } from '$lib/types/api';

export const taskApi = {
	// ==================== 任务管理 ====================

	/**
	 * 创建任务
	 */
	createTask: async (data: CreateTaskRequest): Promise<TaskResponse> => {
		const response = await taskApiClient.post<TaskResponse>('/tasks', data);
		return response.data;
	},

	/**
	 * 获取任务列表
	 */
	getTasks: async (params?: TaskQueryParams): Promise<TaskListResponse> => {
		// 确保参数名称与后端一致：page和pageSize
		const queryParams = {
			...params,
			page: params?.page || 1,
			pageSize: params?.pageSize || 20
		};
		const response = await taskApiClient.get<TaskListResponse>('/tasks', { params: queryParams });
		return response.data;
	},

	/**
	 * 获取任务详情
	 */
	getTask: async (taskId: string): Promise<TaskResponse> => {
		const response = await taskApiClient.get<TaskResponse>(`/tasks/${taskId}`);
		return response.data;
	},

	/**
	 * 更新任务
	 */
	updateTask: async (taskId: string, data: UpdateTaskRequest): Promise<TaskResponse> => {
		const response = await taskApiClient.put<TaskResponse>(`/tasks/${taskId}`, data);
		return response.data;
	},

	/**
	 * 删除任务
	 */
	deleteTask: async (taskId: string): Promise<APIResponse<void>> => {
		const response = await taskApiClient.delete<APIResponse<void>>(`/tasks/${taskId}`);
		return response.data;
	},

	/**
	 * 更新任务状态
	 */
	updateTaskStatus: async (
		taskId: string,
		data: UpdateTaskStatusRequest
	): Promise<APIResponse<void>> => {
		const response = await api.put<APIResponse<void>>(`/tasks/${taskId}/status`, data);
		return response.data;
	},

	/**
	 * 启动任务
	 */
	startTask: async (taskId: string): Promise<APIResponse<void>> => {
		const response = await api.post<APIResponse<void>>(`/tasks/${taskId}/start`);
		return response.data;
	},

	/**
	 * 暂停任务
	 */
	pauseTask: async (taskId: string): Promise<APIResponse<void>> => {
		const response = await api.post<APIResponse<void>>(`/tasks/${taskId}/pause`);
		return response.data;
	},

	/**
	 * 恢复任务
	 */
	resumeTask: async (taskId: string): Promise<APIResponse<void>> => {
		const response = await api.post<APIResponse<void>>(`/tasks/${taskId}/resume`);
		return response.data;
	},

	/**
	 * 取消任务
	 */
	cancelTask: async (taskId: string): Promise<APIResponse<void>> => {
		const response = await api.post<APIResponse<void>>(`/tasks/${taskId}/cancel`);
		return response.data;
	},

	/**
	 * 重启任务
	 */
	restartTask: async (taskId: string): Promise<APIResponse<void>> => {
		const response = await api.post<APIResponse<void>>(`/tasks/${taskId}/restart`);
		return response.data;
	},

	/**
	 * 克隆任务
	 */
	cloneTask: async (taskId: string): Promise<TaskResponse> => {
		const response = await api.post<TaskResponse>(`/tasks/${taskId}/clone`);
		return response.data;
	},

	/**
	 * 批量操作任务
	 */
	batchOperation: async (data: BatchTaskRequest): Promise<APIResponse<BatchTaskResult>> => {
		const response = await api.post<APIResponse<BatchTaskResult>>('/tasks/batch', data);
		return response.data;
	},

	// ==================== 任务结果 ====================

	/**
	 * 获取任务结果
	 */
	getTaskResult: async (taskId: string): Promise<TaskResultResponse> => {
		const response = await api.get<TaskResultResponse>(`/tasks/${taskId}/result`);
		return response.data;
	},

	/**
	 * 获取任务结果列表
	 */
	getTaskResults: async (params?: { taskIds: string[] }): Promise<APIResponse<TaskResult[]>> => {
		const response = await api.get<APIResponse<TaskResult[]>>('/tasks/results', { params });
		return response.data;
	},

	/**
	 * 下载任务结果
	 */
	downloadTaskResult: async (
		taskId: string,
		format: 'json' | 'csv' | 'xml' = 'json'
	): Promise<void> => {
		const response = await api.get(`/tasks/${taskId}/result/download`, {
			params: { format },
			responseType: 'blob'
		});

		// 创建下载链接
		const url = window.URL.createObjectURL(new Blob([response.data]));
		const link = document.createElement('a');
		link.href = url;
		link.setAttribute('download', `task-result-${taskId}.${format}`);
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
		window.URL.revokeObjectURL(url);
	},

	// ==================== 任务事件 ====================

	/**
	 * 获取任务事件
	 */
	getTaskEvents: async (
		taskId: string,
		params?: { page?: number; pageSize?: number }
	): Promise<TaskEventListResponse> => {
		const response = await api.get<TaskEventListResponse>(`/tasks/${taskId}/events`, { params });
		return response.data;
	},

	/**
	 * 获取任务事件流 (SSE)
	 */
	getTaskEventStream: (
		taskId: string,
		onEvent: (event: TaskEvent) => void,
		onError?: (error: Error) => void
	): EventSource => {
		const eventSource = new EventSource(`/api/v1/tasks/${taskId}/events/stream`);

		eventSource.onmessage = (event) => {
			try {
				const taskEvent = JSON.parse(event.data) as TaskEvent;
				onEvent(taskEvent);
			} catch (error) {
				onError?.(error as Error);
			}
		};

		eventSource.onerror = (error) => {
			onError?.(error as unknown as Error);
		};

		return eventSource;
	},

	// ==================== 任务日志 ====================

	/**
	 * 获取任务日志
	 */
	getTaskLogs: async (
		taskId: string,
		params?: TaskLogQueryParams
	): Promise<TaskLogListResponse> => {
		const response = await api.get<TaskLogListResponse>(`/tasks/${taskId}/logs`, { params });
		return response.data;
	},

	/**
	 * 获取任务日志流 (SSE)
	 */
	getTaskLogStream: (
		taskId: string,
		onLog: (log: TaskLog) => void,
		onError?: (error: Error) => void
	): EventSource => {
		const eventSource = new EventSource(`/api/v1/tasks/${taskId}/logs/stream`);

		eventSource.onmessage = (event) => {
			try {
				const taskLog = JSON.parse(event.data) as TaskLog;
				onLog(taskLog);
			} catch (error) {
				onError?.(error as Error);
			}
		};

		eventSource.onerror = (error) => {
			onError?.(error as unknown as Error);
		};

		return eventSource;
	},

	// ==================== 任务统计 ====================

	/**
	 * 获取任务统计
	 */
	getTaskStats: async (projectId?: string): Promise<TaskStatsResponse> => {
		const params = projectId ? { projectId } : undefined;
		const response = await taskApiClient.get<TaskStatsResponse>('/tasks/stats', { params });
		return response.data;
	},

	/**
	 * 获取任务性能指标
	 */
	getTaskPerformanceMetrics: async (
		taskId: string
	): Promise<APIResponse<TaskPerformanceMetrics[]>> => {
		const response = await api.get<APIResponse<TaskPerformanceMetrics[]>>(
			`/tasks/${taskId}/metrics`
		);
		return response.data;
	},

	// ==================== 任务执行器 ====================

	/**
	 * 获取执行器信息
	 */
	getExecutors: async (): Promise<ExecutorInfoResponse> => {
		const response = await api.get<ExecutorInfoResponse>('/tasks/executors');
		return response.data;
	},

	/**
	 * 获取运行中的任务
	 */
	getRunningTasks: async (): Promise<APIResponse<string[]>> => {
		const response = await api.get<APIResponse<string[]>>('/tasks/running');
		return response.data;
	},

	/**
	 * 获取任务执行上下文
	 */
	getExecutionContext: async (taskId: string): Promise<APIResponse<ExecutionContext>> => {
		const response = await api.get<APIResponse<ExecutionContext>>(`/tasks/${taskId}/context`);
		return response.data;
	},

	// ==================== 任务模板 ====================

	/**
	 * 创建任务模板
	 */
	createTaskTemplate: async (
		data: CreateTaskTemplateRequest
	): Promise<APIResponse<TaskTemplate>> => {
		const response = await api.post<APIResponse<TaskTemplate>>('/tasks/templates', data);
		return response.data;
	},

	/**
	 * 获取任务模板列表
	 */
	getTaskTemplates: async (params?: {
		page?: number;
		pageSize?: number;
		search?: string;
		type?: string;
	}): Promise<TaskTemplateListResponse> => {
		const response = await api.get<TaskTemplateListResponse>('/tasks/templates', { params });
		return response.data;
	},

	/**
	 * 获取任务模板详情
	 */
	getTaskTemplate: async (templateId: string): Promise<APIResponse<TaskTemplate>> => {
		const response = await api.get<APIResponse<TaskTemplate>>(`/tasks/templates/${templateId}`);
		return response.data;
	},

	/**
	 * 更新任务模板
	 */
	updateTaskTemplate: async (
		templateId: string,
		data: Partial<CreateTaskTemplateRequest>
	): Promise<APIResponse<TaskTemplate>> => {
		const response = await api.put<APIResponse<TaskTemplate>>(
			`/tasks/templates/${templateId}`,
			data
		);
		return response.data;
	},

	/**
	 * 删除任务模板
	 */
	deleteTaskTemplate: async (templateId: string): Promise<APIResponse<void>> => {
		const response = await api.delete<APIResponse<void>>(`/tasks/templates/${templateId}`);
		return response.data;
	},

	/**
	 * 从模板创建任务
	 */
	createTaskFromTemplate: async (
		templateId: string,
		data: { projectId: string; config?: Record<string, any> }
	): Promise<TaskResponse> => {
		const response = await api.post<TaskResponse>(`/tasks/templates/${templateId}/create`, data);
		return response.data;
	},

	// ==================== 任务调度 ====================

	/**
	 * 创建任务调度规则
	 */
	createTaskScheduleRule: async (
		data: CreateTaskScheduleRuleRequest
	): Promise<APIResponse<TaskScheduleRule>> => {
		const response = await api.post<APIResponse<TaskScheduleRule>>('/tasks/schedule-rules', data);
		return response.data;
	},

	/**
	 * 获取任务调度规则列表
	 */
	getTaskScheduleRules: async (params?: {
		page?: number;
		pageSize?: number;
		projectId?: string;
	}): Promise<TaskScheduleRuleListResponse> => {
		const response = await api.get<TaskScheduleRuleListResponse>('/tasks/schedule-rules', {
			params
		});
		return response.data;
	},

	/**
	 * 获取任务调度规则详情
	 */
	getTaskScheduleRule: async (ruleId: string): Promise<APIResponse<TaskScheduleRule>> => {
		const response = await api.get<APIResponse<TaskScheduleRule>>(
			`/tasks/schedule-rules/${ruleId}`
		);
		return response.data;
	},

	/**
	 * 更新任务调度规则
	 */
	updateTaskScheduleRule: async (
		ruleId: string,
		data: Partial<CreateTaskScheduleRuleRequest>
	): Promise<APIResponse<TaskScheduleRule>> => {
		const response = await api.put<APIResponse<TaskScheduleRule>>(
			`/tasks/schedule-rules/${ruleId}`,
			data
		);
		return response.data;
	},

	/**
	 * 删除任务调度规则
	 */
	deleteTaskScheduleRule: async (ruleId: string): Promise<APIResponse<void>> => {
		const response = await api.delete<APIResponse<void>>(`/tasks/schedule-rules/${ruleId}`);
		return response.data;
	},

	/**
	 * 启用/禁用任务调度规则
	 */
	toggleTaskScheduleRule: async (ruleId: string, enabled: boolean): Promise<APIResponse<void>> => {
		const response = await api.put<APIResponse<void>>(`/tasks/schedule-rules/${ruleId}/toggle`, {
			enabled
		});
		return response.data;
	},

	/**
	 * 手动触发任务调度规则
	 */
	triggerTaskScheduleRule: async (ruleId: string): Promise<TaskResponse> => {
		const response = await api.post<TaskResponse>(`/tasks/schedule-rules/${ruleId}/trigger`);
		return response.data;
	},

	// ==================== 任务配置验证 ====================

	/**
	 * 验证任务配置
	 */
	validateTaskConfig: async (
		type: string,
		config: Record<string, any>
	): Promise<APIResponse<TaskConfigValidation>> => {
		const response = await api.post<APIResponse<TaskConfigValidation>>('/tasks/validate-config', {
			type,
			config
		});
		return response.data;
	},

	/**
	 * 获取任务配置模板
	 */
	getTaskConfigTemplate: async (type: string): Promise<APIResponse<Record<string, any>>> => {
		const response = await api.get<APIResponse<Record<string, any>>>(
			`/tasks/config-template/${type}`
		);
		return response.data;
	},

	/**
	 * 生成任务执行计划
	 */
	generateTaskExecutionPlan: async (
		data: CreateTaskRequest
	): Promise<APIResponse<TaskExecutionPlan>> => {
		const response = await api.post<APIResponse<TaskExecutionPlan>>('/tasks/execution-plan', data);
		return response.data;
	},

	// ==================== 任务导入导出 ====================

	/**
	 * 导出任务
	 */
	exportTasks: async (params: {
		projectId?: string;
		taskIds?: string[];
		includeResults?: boolean;
	}): Promise<void> => {
		const response = await api.get('/tasks/export', {
			params,
			responseType: 'blob'
		});

		// 创建下载链接
		const url = window.URL.createObjectURL(new Blob([response.data]));
		const link = document.createElement('a');
		link.href = url;
		link.setAttribute('download', `tasks-export-${new Date().toISOString().split('T')[0]}.json`);
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
		window.URL.revokeObjectURL(url);
	},

	/**
	 * 导入任务
	 */
	importTasks: async (data: TaskImportRequest): Promise<APIResponse<TaskImportResult>> => {
		const response = await api.post<APIResponse<TaskImportResult>>('/tasks/import', data);
		return response.data;
	},

	// ==================== 任务通知 ====================

	/**
	 * 获取任务通知配置
	 */
	getTaskNotificationConfig: async (
		taskId: string
	): Promise<APIResponse<TaskNotificationConfig>> => {
		const response = await api.get<APIResponse<TaskNotificationConfig>>(
			`/tasks/${taskId}/notification-config`
		);
		return response.data;
	},

	/**
	 * 更新任务通知配置
	 */
	updateTaskNotificationConfig: async (
		taskId: string,
		config: Partial<TaskNotificationConfig>
	): Promise<APIResponse<TaskNotificationConfig>> => {
		const response = await api.put<APIResponse<TaskNotificationConfig>>(
			`/tasks/${taskId}/notification-config`,
			config
		);
		return response.data;
	},

	// ==================== 任务队列 ====================

	/**
	 * 获取任务队列列表
	 */
	getTaskQueues: async (): Promise<APIResponse<TaskQueue[]>> => {
		const response = await api.get<APIResponse<TaskQueue[]>>('/tasks/queues');
		return response.data;
	},

	/**
	 * 获取任务队列详情
	 */
	getTaskQueue: async (queueId: string): Promise<APIResponse<TaskQueue>> => {
		const response = await api.get<APIResponse<TaskQueue>>(`/tasks/queues/${queueId}`);
		return response.data;
	},

	/**
	 * 暂停任务队列
	 */
	pauseTaskQueue: async (queueId: string): Promise<APIResponse<void>> => {
		const response = await api.post<APIResponse<void>>(`/tasks/queues/${queueId}/pause`);
		return response.data;
	},

	/**
	 * 恢复任务队列
	 */
	resumeTaskQueue: async (queueId: string): Promise<APIResponse<void>> => {
		const response = await api.post<APIResponse<void>>(`/tasks/queues/${queueId}/resume`);
		return response.data;
	},

	/**
	 * 清空任务队列
	 */
	clearTaskQueue: async (queueId: string): Promise<APIResponse<void>> => {
		const response = await api.post<APIResponse<void>>(`/tasks/queues/${queueId}/clear`);
		return response.data;
	}
};

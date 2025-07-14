// 任务管理相关类型定义，与后端models/task.go保持一致

// 任务状态枚举
export type TaskStatus =
	| 'pending'
	| 'queued'
	| 'running'
	| 'completed'
	| 'failed'
	| 'cancelled'
	| 'timeout';

// 任务优先级枚举
export type TaskPriority = 'low' | 'normal' | 'high' | 'critical';

// 任务类型枚举
export type TaskType =
	| 'subdomain_enum'
	| 'port_scan'
	| 'vuln_scan'
	| 'asset_discovery'
	| 'dir_scan'
	| 'web_crawl'
	| 'sensitive_scan'
	| 'page_monitor';

// 基础任务接口
export interface Task {
	id: string;
	name: string;
	description?: string;
	type: TaskType;
	status: TaskStatus;
	priority: TaskPriority;
	projectId: string;
	config: Record<string, any>;
	progress: number;
	startedAt?: string;
	completedAt?: string;
	createdAt: string;
	updatedAt: string;
	retryCount: number;
	maxRetries: number;
	timeout: number;
	tags: string[];
	parentTaskId?: string;
	childTaskIds: string[];
	dependencies: string[];
	nodeId?: string;
	assignedAt?: string;
	createdBy?: string;
	metadata?: Record<string, any>;
	result?: Record<string, any>; // 任务结果数据
	error?: string; // 任务错误信息
}

// 任务进度接口
export interface TaskProgress {
	taskId: string;
	progress: number;
	message?: string;
	details?: Record<string, any>;
	completed: number;
	total: number;
	startTime: string;
	estimatedEndTime?: string;
	lastUpdate: string;
}

// 任务结果接口
export interface TaskResult {
	id: string;
	taskId: string;
	status: TaskStatus;
	data: Record<string, any>;
	error?: string;
	summary?: string;
	startTime: string;
	endTime?: string;
	createdAt: string;
	updatedAt: string;
	metadata?: Record<string, any>;
}

// 任务事件接口
export interface TaskEvent {
	id: string;
	taskId: string;
	type: string;
	message: string;
	data?: Record<string, any>;
	timestamp: string;
	nodeId?: string;
	createdAt: string;
}

// 任务队列接口
export interface TaskQueue {
	id: string;
	name: string;
	type: TaskType;
	priority: TaskPriority;
	maxSize: number;
	currentSize: number;
	status: 'active' | 'paused' | 'stopped';
	createdAt: string;
	updatedAt: string;
}

// 任务执行器信息
export interface ExecutorInfo {
	name: string;
	version: string;
	description: string;
	author: string;
	supportedTypes: TaskType[];
}

// 任务执行上下文
export interface ExecutionContext {
	taskId: string;
	executorName: string;
	startTime: string;
	progress: number;
	status: TaskStatus;
	lastUpdate: string;
	retryCount: number;
}

// 任务创建请求
export interface CreateTaskRequest {
	name: string;
	description?: string;
	type: TaskType;
	priority?: TaskPriority;
	projectId: string;
	config: Record<string, any>;
	timeout?: number;
	maxRetries?: number;
	tags?: string[];
	parentTaskId?: string;
	dependencies?: string[];
	scheduledAt?: string;
	metadata?: Record<string, any>;
}

// 任务更新请求
export interface UpdateTaskRequest {
	name?: string;
	description?: string;
	priority?: TaskPriority;
	config?: Record<string, any>;
	timeout?: number;
	maxRetries?: number;
	tags?: string[];
	metadata?: Record<string, any>;
}

// 任务状态更新请求
export interface UpdateTaskStatusRequest {
	status: TaskStatus;
	progress?: number;
	message?: string;
	data?: Record<string, any>;
}

// 任务查询参数
export interface TaskQueryParams {
	projectId?: string;
	status?: TaskStatus;
	type?: TaskType;
	priority?: TaskPriority;
	search?: string;
	startTime?: string;
	endTime?: string;
	page?: number;
	pageSize?: number;
	sortBy?: string;
	sortDesc?: boolean;
	tags?: string[];
	nodeId?: string;
	createdBy?: string;
}

// 任务列表结果
export interface TaskListResult {
	items: Task[];
	total: number;
	page: number;
	pageSize: number;
	totalPages: number;
}

// 任务统计信息
export interface TaskStats {
	total: number;
	pending: number;
	queued: number;
	running: number;
	completed: number;
	failed: number;
	cancelled: number;
	timeout: number;
	byType: Record<TaskType, number>;
	byPriority: Record<TaskPriority, number>;
	byProject: Record<string, number>;
}

// 任务批量操作请求
export interface BatchTaskRequest {
	taskIds: string[];
	action: 'cancel' | 'retry' | 'delete' | 'pause' | 'resume';
	reason?: string;
}

// 任务批量操作结果
export interface BatchTaskResult {
	successCount: number;
	failedCount: number;
	results: Array<{
		taskId: string;
		success: boolean;
		error?: string;
	}>;
}

// 任务依赖关系
export interface TaskDependency {
	id: string;
	taskId: string;
	dependsOnTaskId: string;
	type: 'sequential' | 'parallel' | 'conditional';
	condition?: string;
	createdAt: string;
}

// 任务模板
export interface TaskTemplate {
	id: string;
	name: string;
	description?: string;
	type: TaskType;
	priority: TaskPriority;
	config: Record<string, any>;
	timeout: number;
	maxRetries: number;
	tags: string[];
	isPublic: boolean;
	createdBy: string;
	createdAt: string;
	updatedAt: string;
	usageCount: number;
	metadata?: Record<string, any>;
}

// 任务模板创建请求
export interface CreateTaskTemplateRequest {
	name: string;
	description?: string;
	type: TaskType;
	priority?: TaskPriority;
	config: Record<string, any>;
	timeout?: number;
	maxRetries?: number;
	tags?: string[];
	isPublic?: boolean;
	metadata?: Record<string, any>;
}

// 任务调度规则
export interface TaskScheduleRule {
	id: string;
	name: string;
	description?: string;
	templateId: string;
	cronExpression: string;
	timezone: string;
	enabled: boolean;
	projectId: string;
	nextRunTime?: string;
	lastRunTime?: string;
	runCount: number;
	maxRuns?: number;
	createdBy: string;
	createdAt: string;
	updatedAt: string;
	metadata?: Record<string, any>;
}

// 任务调度规则创建请求
export interface CreateTaskScheduleRuleRequest {
	name: string;
	description?: string;
	templateId: string;
	cronExpression: string;
	timezone?: string;
	enabled?: boolean;
	projectId: string;
	maxRuns?: number;
	metadata?: Record<string, any>;
}

// 任务性能指标
export interface TaskPerformanceMetrics {
	taskId: string;
	executionTime: number;
	cpuUsage: number;
	memoryUsage: number;
	networkIO: number;
	diskIO: number;
	errorRate: number;
	successRate: number;
	averageResponseTime: number;
	peakMemoryUsage: number;
	timestamp: string;
}

// 任务日志
export interface TaskLog {
	id: string;
	taskId: string;
	level: 'debug' | 'info' | 'warn' | 'error';
	message: string;
	data?: Record<string, any>;
	timestamp: string;
	nodeId?: string;
	source: string;
}

// 任务日志查询参数
export interface TaskLogQueryParams {
	taskId?: string;
	level?: string;
	startTime?: string;
	endTime?: string;
	search?: string;
	page?: number;
	pageSize?: number;
	source?: string;
}

// 任务配置验证结果
export interface TaskConfigValidation {
	valid: boolean;
	errors: Array<{
		field: string;
		message: string;
		code: string;
	}>;
	warnings: Array<{
		field: string;
		message: string;
		code: string;
	}>;
}

// 任务执行计划
export interface TaskExecutionPlan {
	taskId: string;
	estimatedDuration: number;
	requiredResources: {
		cpu: number;
		memory: number;
		network: number;
	};
	dependencies: string[];
	executionOrder: number;
	recommendedNode?: string;
	riskLevel: 'low' | 'medium' | 'high';
	notes?: string[];
}

// 任务导入导出
export interface TaskExport {
	version: string;
	exportedAt: string;
	tasks: Task[];
	templates: TaskTemplate[];
	scheduleRules: TaskScheduleRule[];
	metadata?: Record<string, any>;
}

export interface TaskImportRequest {
	data: TaskExport;
	projectId: string;
	overwriteExisting?: boolean;
	skipValidation?: boolean;
}

export interface TaskImportResult {
	importedTasks: number;
	importedTemplates: number;
	importedScheduleRules: number;
	skippedItems: number;
	errors: Array<{
		type: string;
		item: string;
		error: string;
	}>;
}

// 任务通知配置
export interface TaskNotificationConfig {
	id: string;
	taskId?: string;
	projectId?: string;
	events: Array<'created' | 'started' | 'completed' | 'failed' | 'cancelled'>;
	channels: Array<'email' | 'webhook' | 'slack' | 'dingtalk'>;
	recipients: string[];
	webhookUrl?: string;
	template?: string;
	enabled: boolean;
	createdAt: string;
	updatedAt: string;
}

// API响应类型
export interface TaskResponse {
	code: number;
	message: string;
	data: Task;
}

export interface TaskListResponse {
	code: number;
	message: string;
	data: TaskListResult;
}

export interface TaskResultResponse {
	code: number;
	message: string;
	data: TaskResult;
}

export interface TaskStatsResponse {
	code: number;
	message: string;
	data: TaskStats;
}

export interface ExecutorInfoResponse {
	code: number;
	message: string;
	data: Record<string, ExecutorInfo>;
}

export interface TaskEventListResponse {
	code: number;
	message: string;
	data: {
		items: TaskEvent[];
		total: number;
		page: number;
		pageSize: number;
	};
}

export interface TaskLogListResponse {
	code: number;
	message: string;
	data: {
		items: TaskLog[];
		total: number;
		page: number;
		pageSize: number;
	};
}

export interface TaskTemplateListResponse {
	code: number;
	message: string;
	data: {
		items: TaskTemplate[];
		total: number;
		page: number;
		pageSize: number;
	};
}

export interface TaskScheduleRuleListResponse {
	code: number;
	message: string;
	data: {
		items: TaskScheduleRule[];
		total: number;
		page: number;
		pageSize: number;
	};
}

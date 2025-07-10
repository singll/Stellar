/**
 * 节点管理相关类型定义
 * 与后端 internal/models/node.go 保持一致
 */

// 节点状态常量
export const NodeStatus = {
	ONLINE: 'online',
	OFFLINE: 'offline',
	DISABLED: 'disabled',
	MAINTAIN: 'maintain',
	REGISTING: 'registing'
} as const;

// 节点角色常量
export const NodeRole = {
	MASTER: 'master',
	SLAVE: 'slave',
	WORKER: 'worker'
} as const;

// 节点状态类型
export type NodeStatusType = (typeof NodeStatus)[keyof typeof NodeStatus];

// 节点角色类型
export type NodeRoleType = (typeof NodeRole)[keyof typeof NodeRole];

// 节点配置接口
export interface NodeConfig {
	maxConcurrentTasks: number; // 最大并发任务数
	maxMemoryUsage: number; // 最大内存使用量(MB)
	maxCpuUsage: number; // 最大CPU使用率(%)
	heartbeatInterval: number; // 心跳间隔(秒)
	taskTimeout: number; // 任务超时时间(秒)
	enabledTaskTypes: string[]; // 启用的任务类型
	logLevel: string; // 日志级别
	autoUpdate: boolean; // 是否自动更新
}

// 节点状态信息接口
export interface NodeStatusInfo {
	cpuUsage: number; // CPU使用率
	memoryUsage: number; // 内存使用量(MB)
	diskUsage: number; // 磁盘使用量(MB)
	loadAverage: number[]; // 负载平均值
	runningTasks: number; // 运行中的任务数
	queuedTasks: number; // 队列中的任务数
	networkIn: number; // 网络入流量(KB/s)
	networkOut: number; // 网络出流量(KB/s)
	uptimeSeconds: number; // 正常运行时间(秒)
	lastUpdateTime: string; // 最后更新时间
}

// 节点任务统计接口
export interface NodeTaskStats {
	totalTasks: number; // 总任务数
	successTasks: number; // 成功任务数
	failedTasks: number; // 失败任务数
	taskTypeStats: Record<string, number>; // 任务类型统计
	avgExecuteTime: number; // 平均执行时间(秒)
	lastTaskTime: string; // 最后任务时间
}

// 节点接口
export interface Node {
	id: string; // 节点ID
	name: string; // 节点名称
	role: NodeRoleType; // 节点角色
	status: NodeStatusType; // 节点状态
	ip: string; // 节点IP
	port: number; // 节点端口
	apiKey: string; // API密钥
	registerTime: string; // 注册时间
	lastHeartbeatTime: string; // 最后心跳时间
	tags: string[]; // 标签
	config: NodeConfig; // 节点配置
	nodeStatus: NodeStatusInfo; // 节点状态信息
	taskStats: NodeTaskStats; // 任务统计
}

// 节点心跳接口
export interface NodeHeartbeat {
	nodeId: string; // 节点ID
	timestamp: string; // 时间戳
	status: NodeStatusType; // 状态
	cpuUsage: number; // CPU使用率
	memoryUsage: number; // 内存使用量(MB)
	runningTasks: number; // 运行中的任务数
	queuedTasks: number; // 队列中的任务数
	version: string; // 版本
}

// 节点注册请求接口
export interface NodeRegistrationRequest {
	name: string; // 节点名称
	ip: string; // 节点IP
	port: number; // 节点端口
	role: NodeRoleType; // 节点角色
	tags: string[]; // 标签
	config: NodeConfig; // 节点配置
}

// 节点注册响应接口
export interface NodeRegistrationResponse {
	nodeId: string; // 节点ID
	apiKey: string; // API密钥
	status: NodeStatusType; // 状态
	message: string; // 消息
}

// 节点配置更新请求接口
export interface NodeConfigUpdateRequest {
	nodeId: string; // 节点ID
	config: NodeConfig; // 节点配置
}

// 节点状态响应接口
export interface NodeStatusResponse {
	nodeId: string; // 节点ID
	name: string; // 节点名称
	status: NodeStatusType; // 状态
	role: NodeRoleType; // 角色
	lastSeen: string; // 最后在线时间
	nodeStatus: NodeStatusInfo; // 节点状态信息
	taskStats: NodeTaskStats; // 任务统计
}

// 节点查询参数接口
export interface NodeQueryParams {
	status?: NodeStatusType; // 状态过滤
	role?: NodeRoleType; // 角色过滤
	tags?: string[]; // 标签过滤
	search?: string; // 搜索关键词
	page?: number; // 页码
	pageSize?: number; // 每页数量
	sortBy?: string; // 排序字段
	sortDesc?: boolean; // 是否降序
	onlineOnly?: boolean; // 仅在线节点
}

// 节点资源使用统计接口
export interface NodeResourceUsageStats {
	totalMemory: number; // 总内存 (MB)
	usedMemory: number; // 已用内存 (MB)
	totalDisk: number; // 总磁盘 (MB)
	usedDisk: number; // 已用磁盘 (MB)
	avgNetworkIn: number; // 平均网络入流量 (KB/s)
	avgNetworkOut: number; // 平均网络出流量 (KB/s)
	highCpuNodes: number; // 高CPU使用率节点数
	highMemNodes: number; // 高内存使用率节点数
	overloadedNodes: number; // 过载节点数
}

// 节点统计信息接口
export interface NodeStats {
	total: number; // 总节点数
	online: number; // 在线节点数
	offline: number; // 离线节点数
	disabled: number; // 禁用节点数
	maintaining: number; // 维护中节点数
	byRole: Record<string, number>; // 按角色统计
	byStatus: Record<string, number>; // 按状态统计
	totalTasks: number; // 总任务数
	runningTasks: number; // 运行中任务数
	queuedTasks: number; // 队列中任务数
	avgCpuUsage: number; // 平均CPU使用率
	avgMemoryUsage: number; // 平均内存使用率
	resourceUsage: NodeResourceUsageStats; // 资源使用统计
	lastUpdateTime: string; // 最后更新时间
}

// 节点健康状态接口
export interface NodeHealth {
	nodeId: string; // 节点ID
	name: string; // 节点名称
	status: NodeStatusType; // 状态
	healthy: boolean; // 是否健康
	lastSeen: string; // 最后在线时间
	uptime: number; // 正常运行时间(秒)
	score: number; // 健康评分
	issues: string[]; // 健康问题列表
}

// 节点更新请求接口
export interface NodeUpdateRequest {
	name?: string; // 节点名称
	role?: NodeRoleType; // 节点角色
	status?: NodeStatusType; // 节点状态
	tags?: string[]; // 标签
	config?: NodeConfig; // 节点配置
}

// 节点状态更新请求接口
export interface NodeStatusUpdateRequest {
	status: NodeStatusType; // 状态
}

// 节点维护模式请求接口
export interface NodeMaintenanceRequest {
	maintenance: boolean; // 是否维护模式
	reason?: string; // 维护原因
}

// 节点批量操作请求接口
export interface NodeBatchOperationRequest {
	action: 'delete' | 'updateStatus'; // 操作类型
	nodeIds: string[]; // 节点ID列表
	data?: any; // 操作数据
}

// 节点清理请求接口
export interface NodeCleanupRequest {
	timeoutHours: number; // 超时时间(小时)
}

// 节点清理响应接口
export interface NodeCleanupResponse {
	cleanedCount: number; // 清理的节点数
}

// 节点列表响应接口
export interface NodeListResponse {
	items: Node[]; // 节点列表
	total: number; // 总数
	page: number; // 当前页
	pageSize: number; // 每页数量
	totalPages: number; // 总页数
}

// 节点分页响应接口
export interface NodePaginatedResponse<T> {
	items: T[]; // 数据列表
	total: number; // 总数
	page: number; // 当前页
	pageSize: number; // 每页数量
	totalPages: number; // 总页数
}

// 节点状态选项
export interface NodeStatusOption {
	value: NodeStatusType;
	label: string;
	color: string;
}

// 节点角色选项
export interface NodeRoleOption {
	value: NodeRoleType;
	label: string;
	description: string;
}

// 节点标签选项
export interface NodeTagOption {
	value: string;
	label: string;
	count: number;
}

// 节点搜索过滤器
export interface NodeSearchFilter {
	status: NodeStatusType[];
	role: NodeRoleType[];
	tags: string[];
	search: string;
	onlineOnly: boolean;
}

// 节点排序选项
export interface NodeSortOption {
	field: string;
	label: string;
	desc: boolean;
}

// 节点操作权限
export interface NodePermission {
	canCreate: boolean;
	canUpdate: boolean;
	canDelete: boolean;
	canManage: boolean;
	canViewStats: boolean;
	canViewHealth: boolean;
}

// 节点操作历史
export interface NodeOperationHistory {
	id: string;
	nodeId: string;
	operation: string;
	operatorId: string;
	operatorName: string;
	timestamp: string;
	details: string;
	result: 'success' | 'failure';
}

// 节点事件类型
export type NodeEventType =
	| 'register'
	| 'unregister'
	| 'heartbeat'
	| 'status_change'
	| 'config_update'
	| 'maintenance'
	| 'error';

// 节点事件接口
export interface NodeEvent {
	id: string;
	nodeId: string;
	nodeName: string;
	eventType: NodeEventType;
	message: string;
	timestamp: string;
	details?: Record<string, any>;
}

// 节点监控数据接口
export interface NodeMonitorData {
	nodeId: string;
	timestamp: string;
	cpuUsage: number;
	memoryUsage: number;
	diskUsage: number;
	networkIn: number;
	networkOut: number;
	runningTasks: number;
	queuedTasks: number;
}

// 节点监控历史接口
export interface NodeMonitorHistory {
	nodeId: string;
	timeRange: string;
	data: NodeMonitorData[];
	summary: {
		avgCpuUsage: number;
		avgMemoryUsage: number;
		maxCpuUsage: number;
		maxMemoryUsage: number;
		totalTasks: number;
		uptime: number;
	};
}

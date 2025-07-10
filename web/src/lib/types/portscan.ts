/**
 * 端口扫描相关类型定义
 */

/**
 * 端口扫描任务配置
 */
export interface PortScanConfig {
	target: string; // 扫描目标（IP或域名）
	ports: string; // 端口配置（如 "80,443,1-1000"）
	scanMethod?: 'tcp' | 'udp' | 'both'; // 扫描方法
	maxWorkers?: number; // 最大并发数
	timeout?: number; // 超时时间（秒）
	enableBanner?: boolean; // 启用banner抓取
	enableSSL?: boolean; // 启用SSL检测
	enableService?: boolean; // 启用服务识别
	rateLimit?: number; // 速率限制（每秒请求数）
}

/**
 * 端口扫描任务
 */
export interface PortScanTask {
	id: string;
	name: string;
	description?: string;
	type: 'port_scan';
	status: 'pending' | 'queued' | 'running' | 'completed' | 'failed' | 'canceled';
	priority: number;
	config: PortScanConfig;
	projectId?: string;
	nodeId?: string;
	createdBy: string;
	createdAt: string;
	updatedAt: string;
	startTime?: string;
	endTime?: string;
	progress: TaskProgress;
	result?: TaskResult;
	error?: string;
}

/**
 * 任务进度
 */
export interface TaskProgress {
	total: number;
	current: number;
	completed: number;
	failed: number;
	percentage: number;
	speed?: number; // 每秒处理数量
	estimatedTime?: number; // 预计剩余时间（秒）
}

/**
 * 任务结果
 */
export interface TaskResult {
	openPorts: PortScanResult[];
	scannedPorts: PortScanResult[];
	openCount: number;
	totalScanned: number;
	target: string;
	scanMethod: string;
	serviceStats: Record<string, number>;
	scanDuration: number;
	summary: string;
}

/**
 * 端口扫描结果
 */
export interface PortScanResult {
	host: string;
	port: number;
	status: 'open' | 'closed' | 'filtered';
	protocol: 'tcp' | 'udp';
	service?: string;
	banner?: string;
	version?: string;
	responseTime: number;
	timestamp: string;
	error?: string;
	sslInfo?: SSLInfo;
	fingerprint?: Record<string, string>;
}

/**
 * SSL证书信息
 */
export interface SSLInfo {
	issuer: string;
	subject: string;
	notBefore: string;
	notAfter: string;
	fingerprint: string;
	version: string;
	cipher: string;
}

/**
 * 任务创建请求
 */
export interface TaskCreateRequest {
	name: string;
	description?: string;
	target: string;
	ports: string;
	scanMethod?: 'tcp' | 'udp' | 'both';
	maxWorkers?: number;
	timeout?: number;
	enableBanner?: boolean;
	enableSSL?: boolean;
	enableService?: boolean;
	rateLimit?: number;
	projectId?: string;
}

/**
 * 任务列表响应
 */
export interface TaskListResponse {
	tasks: PortScanTask[];
	total: number;
	page: number;
	limit: number;
	totalPages: number;
}

/**
 * 端口扫描表单数据
 */
export interface PortScanFormData {
	name: string;
	description: string;
	target: string;
	ports: string;
	preset: string;
	scanMethod: 'tcp' | 'udp' | 'both';
	advanced: {
		maxWorkers: number;
		timeout: number;
		enableBanner: boolean;
		enableSSL: boolean;
		enableService: boolean;
		rateLimit: number;
	};
	projectId?: string;
}

/**
 * 端口扫描统计数据
 */
export interface PortScanStatistics {
	totalTasks: number;
	completedTasks: number;
	failedTasks: number;
	totalPorts: number;
	openPorts: number;
	successRate: number;
	averageDuration: number;
	serviceStats: Record<string, number>;
	recentTasks: PortScanTask[];
	topTargets: Array<{
		target: string;
		count: number;
		lastScan: string;
	}>;
}

/**
 * 端口预设配置
 */
export interface PortPreset {
	id: string;
	name: string;
	ports: string;
	description: string;
	category: 'common' | 'web' | 'database' | 'service' | 'security';
}

/**
 * 扫描任务过滤器
 */
export interface TaskFilter {
	status?: string;
	target?: string;
	projectId?: string;
	dateRange?: {
		startDate: string;
		endDate: string;
	};
	createdBy?: string;
}

/**
 * 端口扫描配置验证结果
 */
export interface ValidationResult {
	valid: boolean;
	errors: string[];
	warnings: string[];
	estimatedDuration?: number;
	portCount?: number;
}

/**
 * 实时进度更新事件
 */
export interface ProgressUpdateEvent {
	taskId: string;
	progress: TaskProgress;
	latestResults?: PortScanResult[];
	status: string;
}

/**
 * 导出配置
 */
export interface ExportConfig {
	format: 'csv' | 'json' | 'xlsx' | 'xml';
	includeDetails: boolean;
	onlyOpenPorts: boolean;
	customFields?: string[];
}

/**
 * 服务指纹信息
 */
export interface ServiceFingerprint {
	service: string;
	version?: string;
	product?: string;
	extraInfo?: string;
	confidence: number;
	cpe?: string;
}

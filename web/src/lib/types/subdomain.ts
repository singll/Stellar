/**
 * 子域名枚举相关类型定义
 */

/**
 * 子域名枚举任务配置
 */
export interface SubdomainConfig {
	target: string; // 目标域名
	maxWorkers?: number; // 最大并发数
	timeout?: number; // 超时时间（秒）
	wordlistPath?: string; // 字典文件路径
	dnsServers?: string[]; // DNS服务器列表
	enableWildcard?: boolean; // 启用通配符检测
	maxRetries?: number; // 最大重试次数
	enumMethods?: string[]; // 枚举方法列表
	rateLimit?: number; // 速率限制（每秒请求数）
	enableDoH?: boolean; // 启用DNS over HTTPS
	enableRecursive?: boolean; // 启用递归枚举
	maxDepth?: number; // 最大递归深度
	verifySubdomains?: boolean; // 验证子域名活跃性
	enableCache?: boolean; // 启用DNS缓存
	cacheTimeout?: number; // 缓存超时时间（秒）
	searchEngineAPIs?: Record<string, string>; // 搜索引擎API配置
}

/**
 * 子域名枚举任务
 */
export interface SubdomainTask {
	id: string;
	name: string;
	description?: string;
	type: 'subdomain_enum';
	status: 'pending' | 'queued' | 'running' | 'completed' | 'failed' | 'canceled';
	priority: number;
	config: SubdomainConfig;
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
	found: number;
	failed: number;
	percentage: number;
	speed?: number; // 每秒处理数量
	estimatedTime?: number; // 预计剩余时间（秒）
}

/**
 * 任务结果
 */
export interface TaskResult {
	subdomains: SubdomainResult[];
	uniqueSubdomains: SubdomainResult[];
	subdomainCount: number;
	uniqueCount: number;
	target: string;
	enumMethods: string[];
	sourceStats: Record<string, number>;
	scanDuration: number;
	summary: string;
}

/**
 * 子域名枚举结果
 */
export interface SubdomainResult {
	subdomain: string;
	ips: string[];
	cname?: string;
	status: 'found' | 'not_found';
	source: string;
	responseTime: number;
	httpStatus?: number;
	httpTitle?: string;
	technologies?: string[];
	takeover?: TakeoverInfo;
	timestamp: string;
	metadata?: Record<string, string>;
}

/**
 * 子域名接管信息
 */
export interface TakeoverInfo {
	vulnerable: boolean;
	service?: string;
	pattern?: string;
	cname?: string;
}

/**
 * 任务创建请求
 */
export interface SubdomainTaskCreateRequest {
	name: string;
	description?: string;
	target: string;
	maxWorkers?: number;
	timeout?: number;
	wordlistPath?: string;
	customWordlist?: string[];
	dnsServers?: string[];
	dnsPreset?: string;
	enableWildcard?: boolean;
	maxRetries?: number;
	enumMethods?: string[];
	methodPreset?: string;
	rateLimit?: number;
	enableDoH?: boolean;
	enableRecursive?: boolean;
	maxDepth?: number;
	verifySubdomains?: boolean;
	enableCache?: boolean;
	cacheTimeout?: number;
	searchEngineAPIs?: Record<string, string>;
	projectId?: string;
}

/**
 * 任务列表响应
 */
export interface TaskListResponse {
	tasks: SubdomainTask[];
	total: number;
	page: number;
	limit: number;
	totalPages: number;
}

/**
 * 子域名枚举表单数据
 */
export interface SubdomainFormData {
	name: string;
	description: string;
	target: string;
	methodPreset: string;
	enumMethods: string[];
	wordlistType: 'preset' | 'custom';
	wordlistPreset: string;
	customWordlist: string[];
	dnsPreset: string;
	customDnsServers: string[];
	advanced: {
		maxWorkers: number;
		timeout: number;
		maxRetries: number;
		rateLimit: number;
		enableWildcard: boolean;
		enableDoH: boolean;
		enableRecursive: boolean;
		maxDepth: number;
		verifySubdomains: boolean;
		enableCache: boolean;
		cacheTimeout: number;
	};
	searchEngines: {
		enableGoogle: boolean;
		enableBing: boolean;
		censysApiKey: string;
		shodanApiKey: string;
	};
	projectId?: string;
}

/**
 * 子域名枚举统计数据
 */
export interface SubdomainStatistics {
	totalTasks: number;
	completedTasks: number;
	failedTasks: number;
	totalSubdomains: number;
	uniqueSubdomains: number;
	successRate: number;
	averageDuration: number;
	sourceStats: Record<string, number>;
	methodStats: Record<string, number>;
	recentTasks: SubdomainTask[];
	topTargets: Array<{
		target: string;
		count: number;
		lastScan: string;
	}>;
}

/**
 * 枚举方法预设
 */
export interface EnumMethodPreset {
	id: string;
	name: string;
	methods: string[];
	description: string;
	estimatedTime: string;
}

/**
 * DNS服务器预设
 */
export interface DNSPreset {
	id: string;
	name: string;
	servers: string[];
	description: string;
	location?: string;
}

/**
 * 字典预设
 */
export interface WordlistPreset {
	id: string;
	name: string;
	size: string;
	description: string;
	category: 'common' | 'comprehensive' | 'security' | 'custom';
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
	enumMethod?: string;
}

/**
 * 子域名枚举配置验证结果
 */
export interface ValidationResult {
	valid: boolean;
	errors: string[];
	warnings: string[];
	estimatedDuration?: number;
	wordlistSize?: number;
}

/**
 * 实时进度更新事件
 */
export interface ProgressUpdateEvent {
	taskId: string;
	progress: TaskProgress;
	latestResults?: SubdomainResult[];
	status: string;
}

/**
 * 导出配置
 */
export interface ExportConfig {
	format: 'csv' | 'json' | 'xlsx' | 'xml';
	includeDetails: boolean;
	includeMetadata: boolean;
	onlyActive: boolean;
	groupBySource: boolean;
	customFields?: string[];
}

/**
 * 子域名详细信息
 */
export interface SubdomainDetail {
	subdomain: string;
	ips: string[];
	cname?: string;
	status: string;
	source: string;
	httpInfo?: {
		status: number;
		title?: string;
		server?: string;
		technologies: string[];
		responseTime: number;
	};
	sslInfo?: {
		issuer: string;
		subject: string;
		notBefore: string;
		notAfter: string;
		fingerprint: string;
	};
	dnsInfo?: {
		aRecords: string[];
		cnameRecord?: string;
		mxRecords: string[];
		txtRecords: string[];
		nsRecords: string[];
	};
	geoInfo?: {
		country: string;
		region: string;
		city: string;
		org: string;
		asn: string;
	};
	takeoverInfo?: TakeoverInfo;
	firstSeen: string;
	lastSeen: string;
	scanCount: number;
}

/**
 * 子域名发现来源统计
 */
export interface SourceStatistics {
	source: string;
	count: number;
	percentage: number;
	averageResponseTime: number;
	successRate: number;
}

/**
 * 子域名验证结果
 */
export interface SubdomainVerification {
	subdomain: string;
	isActive: boolean;
	httpStatus?: number;
	httpsStatus?: number;
	responseTime: number;
	technologies: string[];
	lastChecked: string;
}

/**
 * 递归枚举配置
 */
export interface RecursiveConfig {
	enabled: boolean;
	maxDepth: number;
	prefixes: string[];
	skipWildcard: boolean;
	onlyActive: boolean;
}

/**
 * 搜索引擎配置
 */
export interface SearchEngineConfig {
	google: {
		enabled: boolean;
		apiKey?: string;
		searchEngineId?: string;
		maxResults: number;
	};
	bing: {
		enabled: boolean;
		apiKey?: string;
		maxResults: number;
	};
	censys: {
		enabled: boolean;
		apiId?: string;
		apiSecret?: string;
	};
	shodan: {
		enabled: boolean;
		apiKey?: string;
	};
}

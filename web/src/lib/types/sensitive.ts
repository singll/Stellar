// 敏感信息检测相关类型定义

export interface SensitiveRule {
	id: string;
	name: string;
	description: string;
	type: string;
	pattern: string;
	category: string;
	riskLevel: string;
	tags: string[];
	enabled: boolean;
	context: number;
	examples: string[];
	falsePositivePatterns: string[];
	createdAt: string;
	updatedAt: string;
}

export interface SensitiveRuleGroup {
	id: string;
	name: string;
	description: string;
	rules: string[];
	enabled: boolean;
	createdAt: string;
	updatedAt: string;
}

export interface SensitiveWhitelist {
	id: string;
	name: string;
	description: string;
	type: string;
	value: string;
	expiresAt: string;
	createdAt: string;
	updatedAt: string;
}

export interface SensitiveDetectionConfig {
	concurrency: number;
	timeout: number;
	maxDepth: number;
	contextLines: number;
	followLinks: boolean;
	userAgent: string;
	ignoreRobots: boolean;
	maxFileSize: number;
	fileTypes: string;
	excludeURLs: string;
	includeURLs: string;
	authentication: string;

	// 文件检测相关配置
	filePatterns?: string[];
	excludePatterns?: string[];
	recursiveSearch?: boolean;
	followSymlinks?: boolean;
	maxFileSizeBytes?: number;
	scanArchives?: boolean;
}

export interface SensitiveDetectionRequest {
	projectId: string;
	name: string;
	description: string;
	targets: string[];
	ruleGroups: { $oid: string }[];
	rules: { $oid: string }[];
	config: SensitiveDetectionConfig;
}

export interface SensitiveFinding {
	id: string;
	target: string;
	targetType: string; // url, file, directory
	rule: string;
	ruleName: string;
	category: string;
	riskLevel: string;
	pattern: string;
	matchedText: string;
	context: string;
	lineNumber?: number; // 行号（对于文件）
	filePath?: string; // 文件路径
	fileSize?: number; // 文件大小
	createdAt: string;
}

export interface SensitiveDetectionSummary {
	totalFindings: number;
	riskLevelCount: {
		high: number;
		medium: number;
		low: number;
	};
	categoryCount: Record<string, number>;
}

export interface SensitiveDetectionResult {
	id: string;
	projectId: string;
	name: string;
	targets: string[];
	status: SensitiveDetectionStatus;
	startTime: string;
	endTime: string;
	progress: number;
	config: SensitiveDetectionConfig;
	findings?: SensitiveFinding[];
	summary?: SensitiveDetectionSummary;
	createdAt: string;
	updatedAt: string;
	totalCount: number;
	finishCount: number;
}

export type SensitiveDetectionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

// 敏感信息规则创建请求
export interface SensitiveRuleCreateRequest {
	name: string;
	description: string;
	type: string;
	pattern: string;
	category: string;
	riskLevel: string;
	tags: string[];
	enabled: boolean;
	context: number;
	examples: string[];
	falsePositivePatterns: string[];
}

// 敏感信息规则更新请求
export interface SensitiveRuleUpdateRequest {
	name: string;
	description: string;
	pattern: string;
	category: string;
	riskLevel: string;
	tags: string[];
	enabled: boolean;
	context: number;
	examples: string[];
	falsePositivePatterns: string[];
}

// 敏感信息规则组创建请求
export interface SensitiveRuleGroupCreateRequest {
	name: string;
	description: string;
	rules: string[];
	enabled: boolean;
}

// 敏感信息规则组更新请求
export interface SensitiveRuleGroupUpdateRequest {
	name: string;
	description: string;
	rules: string[];
	enabled: boolean;
}

// 敏感信息白名单创建请求
export interface SensitiveWhitelistCreateRequest {
	name: string;
	description: string;
	type: string;
	value: string;
	expiresAt: Date;
}

// 敏感信息白名单更新请求
export interface SensitiveWhitelistUpdateRequest {
	name: string;
	description: string;
	type: string;
	value: string;
	expiresAt: Date;
}

// 规则类型枚举
export const SensitiveRuleType = {
	REGEX: 'regex',
	KEYWORD: 'keyword',
	PATTERN: 'pattern'
} as const;

export type SensitiveRuleTypeValue = (typeof SensitiveRuleType)[keyof typeof SensitiveRuleType];

// 风险等级枚举
export const SensitiveRiskLevel = {
	HIGH: 'high',
	MEDIUM: 'medium',
	LOW: 'low'
} as const;

export type SensitiveRiskLevelValue = (typeof SensitiveRiskLevel)[keyof typeof SensitiveRiskLevel];

// 敏感信息分类枚举
export const SensitiveCategory = {
	PASSWORD: 'password',
	API_KEY: 'api_key',
	TOKEN: 'token',
	EMAIL: 'email',
	PHONE: 'phone',
	CARD: 'card',
	ID_CARD: 'id_card',
	OTHER: 'other'
} as const;

export type SensitiveCategoryValue = (typeof SensitiveCategory)[keyof typeof SensitiveCategory];

// 白名单类型枚举
export const SensitiveWhitelistType = {
	TARGET: 'target',
	PATTERN: 'pattern'
} as const;

export type SensitiveWhitelistTypeValue =
	(typeof SensitiveWhitelistType)[keyof typeof SensitiveWhitelistType];

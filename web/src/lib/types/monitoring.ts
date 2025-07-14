// 页面监控相关类型定义

export interface PageMonitoring {
	id: string;
	url: string;
	name: string;
	status: string;
	projectId: string;
	interval: number;
	lastCheckAt: string;
	nextCheckAt: string;
	createdAt: string;
	updatedAt: string;
	tags: string[];
	config: MonitoringConfig;
	latestSnapshot?: PageSnapshot;
	previousSnapshot?: PageSnapshot;
	changeCount: number;
	hasChanged: boolean;
	similarity: number;
}

export interface MonitoringConfig {
	ignoreCSS: boolean;
	ignoreJS: boolean;
	ignoreImages: boolean;
	ignoreNumbers: boolean;
	ignorePatterns: string[];
	similarityThreshold: number;
	timeout: number;
	headers: Record<string, string>;
	authentication: AuthConfig;
	selector: string;
	compareMethod: string;
	notifyOnChange: boolean;
	notifyMethods: string[];
	notifyConfig: Record<string, string>;
}

export interface AuthConfig {
	type: string;
	username: string;
	password: string;
	cookie: string;
}

export interface PageSnapshot {
	id: string;
	monitoringId: string;
	url: string;
	statusCode: number;
	headers: Record<string, string>;
	html: string;
	text: string;
	contentHash: string;
	createdAt: string;
	size: number;
	loadTime: number;
}

export interface PageChange {
	id: string;
	monitoringId: string;
	url: string;
	status: string;
	oldSnapshotId: string;
	newSnapshotId: string;
	similarity: number;
	changedAt: string;
	diff: string;
	diffType: string;
}

export interface PageMonitoringCreateRequest {
	url: string;
	name: string;
	projectId: string;
	interval: number;
	tags: string[];
	config: MonitoringConfig;
}

export interface PageMonitoringUpdateRequest {
	name: string;
	status: string;
	interval: number;
	tags: string[];
	config: MonitoringConfig;
}

export interface PageMonitoringQueryRequest {
	projectId?: string;
	status?: string;
	tags?: string[];
	url?: string;
	name?: string;
	limit?: number;
	offset?: number;
	sortBy?: string;
	sortOrder?: string;
}

// 页面监控状态枚举
export const PageMonitoringStatus = {
	ACTIVE: 'active',
	INACTIVE: 'inactive',
	ERROR: 'error'
} as const;

export type PageMonitoringStatusValue =
	(typeof PageMonitoringStatus)[keyof typeof PageMonitoringStatus];

// 页面变更状态枚举
export const PageChangeStatus = {
	NEW: 'new',
	CHANGED: 'changed',
	REMOVED: 'removed',
	UNCHANGED: 'unchanged'
} as const;

export type PageChangeStatusValue = (typeof PageChangeStatus)[keyof typeof PageChangeStatus];

// 比较方法枚举
export const CompareMethod = {
	TEXT: 'text',
	HTML: 'html',
	VISUAL: 'visual',
	HASH: 'hash'
} as const;

export type CompareMethodValue = (typeof CompareMethod)[keyof typeof CompareMethod];

// 认证类型枚举
export const AuthType = {
	NONE: 'none',
	BASIC: 'basic',
	FORM: 'form',
	COOKIE: 'cookie'
} as const;

export type AuthTypeValue = (typeof AuthType)[keyof typeof AuthType];

// 通知方式枚举
export const NotifyMethod = {
	EMAIL: 'email',
	WEBHOOK: 'webhook',
	SMS: 'sms'
} as const;

export type NotifyMethodValue = (typeof NotifyMethod)[keyof typeof NotifyMethod];

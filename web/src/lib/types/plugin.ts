// 插件相关类型定义

export interface PluginMetadata {
	id: string;
	name: string;
	version: string;
	type: string;
	author: string;
	description: string;
	category: string;
	tags: string[];
	path: string;
	config: Record<string, any>;
	enabled: boolean;
	installTime: string;
	updateTime: string;
	lastRunTime: string;
	runCount: number;
	errorCount: number;
	avgRuntime: number;
	dependencies: PluginDependency[];
}

export interface PluginDependency {
	id: string;
	version: string;
}

export interface PluginRunRecord {
	id: string;
	pluginId: string;
	startTime: string;
	endTime: string;
	duration: number;
	success: boolean;
	error?: string;
	params: Record<string, any>;
	result?: any;
	taskId?: string;
	userId?: string;
}

export interface PluginConfig {
	id: string;
	pluginId: string;
	name: string;
	description: string;
	config: Record<string, any>;
	isDefault: boolean;
	createdAt: string;
	updatedAt: string;
	createdBy: string;
}

export interface PluginMarketItem {
	id: string;
	name: string;
	version: string;
	type: string;
	author: string;
	description: string;
	category: string;
	tags: string[];
	downloadUrl: string;
	homepage: string;
	license: string;
	stars: number;
	downloads: number;
	publishTime: string;
	updateTime: string;
	verified: boolean;
	screenshots: string[];
	dependencies: PluginDependency[];
}

// YAML插件配置
export interface YAMLPluginConfig {
	id: string;
	name: string;
	version: string;
	author: string;
	description: string;
	type: string;
	category: string;
	tags: string[];
	dependencies: string[];
	config: Record<string, any>;
	script: ScriptConfig;
}

export interface ScriptConfig {
	language: 'python' | 'javascript' | 'shell' | 'lua';
	content: string;
	entry: string;
	args: string[];
}

// 插件执行结果
export interface PluginExecutionResult {
	success: boolean;
	data?: any;
	error?: string;
	metadata?: {
		executionTime?: number;
		pluginVersion?: string;
	};
}

// 插件安装请求
export interface PluginInstallRequest {
	method: 'file' | 'url' | 'yaml';
	file?: File;
	url?: string;
	yaml?: string;
}

// 插件安装响应
export interface PluginInstallResponse {
	success: boolean;
	message: string;
	plugin?: PluginMetadata;
	error?: string;
}

// 插件类型枚举
export const PluginType = {
	SCANNER: 'scanner',
	INFO_GATHERER: 'info_gatherer',
	VULNERABILITY: 'vulnerability',
	UTILITY: 'utility',
	CUSTOM: 'custom'
} as const;

export type PluginTypeValue = (typeof PluginType)[keyof typeof PluginType];

// 插件分类枚举
export const PluginCategory = {
	SUBDOMAIN: 'subdomain',
	PORT: 'port',
	WEB: 'web',
	NETWORK: 'network',
	OSINT: 'osint',
	MISC: 'misc'
} as const;

export type PluginCategoryValue = (typeof PluginCategory)[keyof typeof PluginCategory];

// 脚本语言枚举
export const ScriptLanguage = {
	PYTHON: 'python',
	JAVASCRIPT: 'javascript',
	SHELL: 'shell',
	LUA: 'lua'
} as const;

export type ScriptLanguageValue = (typeof ScriptLanguage)[keyof typeof ScriptLanguage];

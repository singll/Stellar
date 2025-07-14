// 资产类型定义，与后端models/asset.go保持一致
export type AssetType =
	| 'domain'
	| 'subdomain'
	| 'ip'
	| 'port'
	| 'url'
	| 'http'
	| 'app'
	| 'miniapp'
	| 'other';

// 基础资产接口
export interface BaseAsset {
	id: string;
	createdAt: string;
	updatedAt: string;
	lastScanTime: string;
	type: AssetType;
	projectId: string;
	tags: string[];
	taskName?: string;
	rootDomain?: string;
	changeHistory?: AssetChange[];
	status?: 'active' | 'inactive' | 'deleted';
	riskLevel?: 'low' | 'medium' | 'high' | 'critical';
	// 公共属性，所有资产类型都应该有
	name?: string;
	description?: string;
	lastScan?: string;
	value?: string; // 资产的主要值（域名、IP、URL等）
}

// 资产变更记录
export interface AssetChange {
	time: string;
	fieldName: string;
	oldValue: any;
	newValue: any;
	changeType: 'add' | 'update' | 'delete';
	metadata?: Record<string, any>;
}

// ICP备案信息
export interface ICPInfo {
	icpNo: string;
	companyName: string;
	companyType: string;
	updateTime: string;
}

// 域名资产
export interface DomainAsset extends BaseAsset {
	type: 'domain';
	domain: string;
	ips: string[];
	whois?: string;
	icpInfo?: ICPInfo;
}

// 子域名资产
export interface SubdomainAsset extends BaseAsset {
	type: 'subdomain';
	host: string;
	ips: string[];
	cname?: string;
	dnsType?: string;
	dnsValue?: string[]; // 重命名为dnsValue避免与BaseAsset的value冲突
	takeOver?: boolean;
}

// IP地理位置信息
export interface IPLocation {
	country: string;
	countryCode: string;
	region: string;
	city: string;
	latitude: number;
	longitude: number;
}

// IP资产
export interface IPAsset extends BaseAsset {
	type: 'ip';
	ip: string;
	location?: IPLocation;
	asn?: string;
	isp?: string;
	fingerprint?: Record<string, any>;
}

// 端口资产
export interface PortAsset extends BaseAsset {
	type: 'port';
	ip: string;
	host?: string;
	port: number;
	service?: string;
	protocol?: string;
	version?: string;
	banner?: string;
	tls?: boolean;
	transport?: string;
	portStatus?: string; // 重命名为portStatus避免与BaseAsset的status冲突
}

// Favicon信息
export interface FaviconInfo {
	path: string;
	mmh3: string;
	content?: string;
}

// URL资产
export interface URLAsset extends BaseAsset {
	type: 'url';
	url: string;
	host: string;
	path?: string;
	query?: string;
	fragment?: string;
	statusCode?: number;
	title?: string;
	contentType?: string;
	contentLength?: number;
	hash?: string;
	screenshot?: string;
	technologies?: string[];
	headers?: Record<string, string>;
	favicon?: FaviconInfo;
	metadata?: Record<string, any>;
}

// HTTP服务资产
export interface HTTPAsset extends BaseAsset {
	type: 'http';
	host: string;
	ip: string;
	port: number;
	url: string;
	title?: string;
	statusCode?: number;
	contentType?: string;
	contentLength?: number;
	webServer?: string;
	tls?: boolean;
	hash?: string;
	cdnName?: string;
	cdn?: boolean;
	screenshot?: string;
	technologies?: string[];
	headers?: Record<string, string>;
	favicon?: FaviconInfo;
	jarm?: string;
	metadata?: Record<string, any>;
}

// 应用资产
export interface AppAsset extends BaseAsset {
	type: 'app';
	appName: string;
	packageName: string;
	platform: string;
	version?: string;
	developer?: string;
	downloadUrl?: string;
	description?: string;
	permissions?: string[];
	sha256?: string;
	iconUrl?: string;
	metadata?: Record<string, any>;
}

// 小程序资产
export interface MiniAppAsset extends BaseAsset {
	type: 'miniapp';
	appName: string;
	appId: string;
	platform: string;
	developer?: string;
	description?: string;
	iconUrl?: string;
	qrCodeUrl?: string;
	metadata?: Record<string, any>;
}

// 资产联合类型
export type Asset =
	| DomainAsset
	| SubdomainAsset
	| IPAsset
	| PortAsset
	| URLAsset
	| HTTPAsset
	| AppAsset
	| MiniAppAsset;

// 资产关系
export interface AssetRelation {
	id: string;
	sourceAssetId: string;
	targetAssetId: string;
	relationType: string;
	createdAt: string;
	updatedAt: string;
	projectId: string;
	metadata?: Record<string, any>;
}

// 创建资产请求
export interface CreateAssetRequest {
	type: AssetType;
	projectId: string;
	rootDomain?: string;
	taskName?: string;
	tags?: string[];
	data: Record<string, any>;
	// 特定类型的字段，用于表单数据绑定
	domain?: string;
	subdomain?: string;
	ip?: string;
	port?: number;
	url?: string;
	appName?: string;
	host?: string;
	path?: string;
	method?: string;
	title?: string;
	server?: string;
	[key: string]: any; // 允许任意其他属性
}

// 更新资产请求
export interface UpdateAssetRequest {
	type: AssetType;
	data: Record<string, any>;
}

// 资产查询参数
export interface AssetQueryParams {
	projectId?: string;
	type?: AssetType;
	rootDomain?: string;
	tags?: string[];
	taskName?: string;
	search?: string;
	page?: number;
	pageSize?: number;
	sortBy?: string;
	sortDesc?: boolean;
	startTime?: string;
	endTime?: string;
}

// 资产列表结果
export interface AssetListResult {
	items: Asset[];
	total: number;
	page: number;
	pageSize: number;
	totalPages: number;
}

// 资产筛选条件
export interface AssetFilter {
	projectId?: string;
	tags?: string[];
	type?: AssetType;
	rootDomain?: string;
	taskName?: string;
	search?: string;
	dateRange?: {
		startTime?: string;
		endTime?: string;
	};
}

// 资产响应
export interface AssetResponse {
	code: number;
	message: string;
	data: AssetListResult;
}

// 批量创建资产请求
export interface BatchCreateAssetsRequest {
	type: AssetType;
	projectId: string;
	rootDomain?: string;
	taskName?: string;
	tags?: string[];
	assets: Record<string, any>[];
}

// 批量删除资产请求
export interface BatchDeleteAssetsRequest {
	type: AssetType;
	ids: string[];
}

// 创建资产关系请求
export interface CreateAssetRelationRequest {
	sourceAssetId: string;
	targetAssetId: string;
	relationType: string;
	projectId: string;
}

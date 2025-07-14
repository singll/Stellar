// 项目类型定义
export interface Project {
	id: string;
	name: string;
	description?: string;
	target?: string;
	scan_status?: string;
	color?: string;
	is_private?: boolean;
	assets_count?: number;
	vulnerabilities_count?: number;
	tasks_count?: number;
	created_by?: string;
	created_at: string;
	updated_at: string;
}

// 创建项目请求类型
export interface CreateProjectRequest {
	name: string;
	description?: string;
	target?: string;
	color?: string;
	is_private?: boolean;
}

// 更新项目请求类型
export interface UpdateProjectRequest {
	name?: string;
	description?: string;
	target?: string;
	color?: string;
	is_private?: boolean;
}

// 项目查询参数类型
export interface ProjectQueryParams {
	page?: number;
	limit?: number;
	search?: string;
	sort_by?: string;
	sort_order?: 'asc' | 'desc';
}

// 项目筛选器类型
export interface ProjectFilters {
	search?: string;
	is_private?: boolean;
	scan_status?: string;
}

// 项目统计类型
export interface ProjectStats {
	total_projects: number;
	active_projects: number;
	total_assets: number;
	total_vulnerabilities: number;
	total_tasks: number;
}

// API响应类型
export interface ProjectListResponse {
	data: Project[];
	total: number;
	page: number;
	limit: number;
}

export interface ProjectResponse {
	id: string;
	name: string;
	description?: string;
	target?: string;
	scan_status?: string;
	color?: string;
	is_private?: boolean;
	assets_count?: number;
	vulnerabilities_count?: number;
	tasks_count?: number;
	created_by?: string;
	created_at: string;
	updated_at: string;
}

// 添加复数形式别名
export interface ProjectsResponse extends ProjectListResponse {}

// 项目颜色选项
export const PROJECT_COLORS = [
	'blue',
	'green',
	'red',
	'yellow',
	'purple',
	'pink',
	'indigo',
	'gray'
] as const;

export type ProjectColor = (typeof PROJECT_COLORS)[number];

// 扫描状态选项
export const SCAN_STATUS_OPTIONS = ['pending', 'running', 'completed', 'failed', 'paused'] as const;

export type ScanStatus = (typeof SCAN_STATUS_OPTIONS)[number];

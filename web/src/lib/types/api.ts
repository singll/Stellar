export interface APIResponse<T = any> {
	code: number;
	message: string;
	data: T;
}

export interface PaginatedResponse<T = any> {
	items: T[];
	total: number;
	page: number;
	pageSize: number;
	totalPages: number;
	limit?: number; // 兼容性
}

// 添加兼容别名
export type ApiResponse<T = any> = APIResponse<T>;

export interface PaginationParams {
	page?: number;
	limit?: number;
	sort?: string;
	order?: 'asc' | 'desc';
}

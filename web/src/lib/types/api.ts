export interface APIResponse<T = any> {
	code: number;
	message: string;
	data: T;
}

export interface PaginatedResponse<T = any> extends APIResponse<T> {
	total: number;
	page: number;
	limit: number;
}

// 添加兼容别名
export type ApiResponse<T = any> = APIResponse<T>;

export interface PaginationParams {
	page?: number;
	limit?: number;
	sort?: string;
	order?: 'asc' | 'desc';
}

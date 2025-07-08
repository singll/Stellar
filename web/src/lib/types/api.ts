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

export { default as Pagination } from './Pagination.svelte';

// 分页相关类型定义
export interface PaginationInfo {
	currentPage: number;
	totalPages: number;
	totalItems: number;
	pageSize: number;
	hasNextPage: boolean;
	hasPrevPage: boolean;
}

export interface PaginationOptions {
	page?: number;
	limit?: number;
	search?: string;
	sortBy?: string;
	sortOrder?: 'asc' | 'desc';
}

// 分页计算工具函数
export function calculatePagination(
	currentPage: number,
	pageSize: number,
	totalItems: number
): PaginationInfo {
	const totalPages = Math.ceil(totalItems / pageSize);
	
	return {
		currentPage: Math.max(1, Math.min(currentPage, totalPages)),
		totalPages: Math.max(1, totalPages),
		totalItems,
		pageSize,
		hasNextPage: currentPage < totalPages,
		hasPrevPage: currentPage > 1
	};
}

// 生成页码范围
export function generatePageRange(
	currentPage: number,
	totalPages: number,
	maxVisible: number = 7
): number[] {
	if (totalPages <= maxVisible) {
		return Array.from({ length: totalPages }, (_, i) => i + 1);
	}

	const half = Math.floor(maxVisible / 2);
	let start = Math.max(1, currentPage - half);
	let end = Math.min(totalPages, start + maxVisible - 1);

	if (end - start + 1 < maxVisible) {
		start = Math.max(1, end - maxVisible + 1);
	}

	return Array.from({ length: end - start + 1 }, (_, i) => start + i);
}
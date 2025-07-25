<!--
通用分页组件
支持页码导航、页数跳转、每页数量设置
与项目整体UI风格保持一致
-->
<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import Icon from '$lib/components/ui/Icon.svelte';

	// 分页配置
	interface PaginationProps {
		currentPage: number;
		totalPages: number;
		totalItems?: number;
		pageSize?: number;
		pageSizeOptions?: number[];
		showPageSizeSelector?: boolean;
		showPageInfo?: boolean;
		showFirstLast?: boolean;
		maxVisiblePages?: number;
		disabled?: boolean;
		class?: string;
	}

	let {
		currentPage = 1,
		totalPages = 1,
		totalItems = 0,
		pageSize = 20,
		pageSizeOptions = [10, 20, 50, 100],
		showPageSizeSelector = true,
		showPageInfo = true,
		showFirstLast = true,
		maxVisiblePages = 7,
		disabled = false,
		class: className = '',
		onPageChange,
		onPageSizeChange
	}: PaginationProps & {
		onPageChange?: (page: number) => void;
		onPageSizeChange?: (pageSize: number) => void;
	} = $props();

	// 计算可见的页码范围
	function getVisiblePages(current: number, total: number, maxVisible: number): number[] {
		// 如果总页数为0，返回空数组
		if (total <= 0) {
			return [];
		}
		
		// 如果总页数小于等于最大可见页数，显示所有页码
		if (total <= maxVisible) {
			return Array.from({ length: total }, (_, i) => i + 1);
		}

		const half = Math.floor(maxVisible / 2);
		let start = Math.max(1, current - half);
		let end = Math.min(total, start + maxVisible - 1);

		// 调整起始位置，确保显示足够的页码
		if (end - start + 1 < maxVisible && end < total) {
			start = Math.max(1, end - maxVisible + 1);
		}

		return Array.from({ length: end - start + 1 }, (_, i) => start + i);
	}

	// 使用 $derived 替代 $: 语法
	const visiblePages = $derived(getVisiblePages(currentPage, totalPages, maxVisiblePages));
	const startItem = $derived((currentPage - 1) * pageSize + 1);
	const endItem = $derived(Math.min(currentPage * pageSize, totalItems));
	
	// 页码跳转处理
	function handlePageChange(page: number) {
		console.log('🔍 [Pagination] handlePageChange called', { page, currentPage, totalPages, disabled });
		if (page === currentPage || page < 1 || page > totalPages || disabled) {
			console.log('🔍 [Pagination] handlePageChange early return', { reason: page === currentPage ? 'same page' : page < 1 ? 'page < 1' : page > totalPages ? 'page > totalPages' : 'disabled' });
			return;
		}
		console.log('🔍 [Pagination] calling onPageChange with page:', page);
		onPageChange?.(page);
	}

	// 每页数量变更处理
	function handlePageSizeChange(newPageSize: number) {
		if (newPageSize === pageSize || disabled) {
			return;
		}
		onPageSizeChange?.(newPageSize);
	}

	// 页面跳转输入框处理
	let jumpToPage = $state<string>('');
	function handleJumpToPage(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			const page = parseInt(jumpToPage);
			if (!isNaN(page) && page >= 1 && page <= totalPages) {
				handlePageChange(page);
				jumpToPage = '';
			}
		}
	}

	function handleJumpClick() {
		const page = parseInt(jumpToPage);
		if (!isNaN(page) && page >= 1 && page <= totalPages) {
			handlePageChange(page);
			jumpToPage = '';
		}
	}
</script>

<div class="flex flex-col gap-4 {className}">
	{#if totalItems === 0}
		<!-- 无数据时不显示分页组件 -->
	{:else}
	<!-- 分页信息 -->
	{#if showPageInfo && totalItems > 0}
		<div class="flex items-center text-sm text-gray-600">
			<span>
				显示 <span class="font-medium text-gray-900">{startItem}</span> 到{' '}
				<span class="font-medium text-gray-900">{endItem}</span> 条，共{' '}
				<span class="font-medium text-gray-900">{totalItems}</span> 条数据
			</span>
		</div>
	{/if}

	<!-- 分页控制区域 -->
	{#if totalItems > 0}
		<div class="flex items-center justify-between">
			<!-- 每页数量选择器 -->
			{#if showPageSizeSelector}
				<div class="flex items-center gap-2 text-sm text-gray-600">
					<span>每页显示</span>
					<select
						value={pageSize.toString()}
						disabled={disabled}
						class="px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 w-20 h-8"
						onchange={(e) => {
							const target = e.target as HTMLSelectElement;
							handlePageSizeChange(parseInt(target.value));
						}}
					>
						{#each pageSizeOptions as option}
							<option value={option.toString()}>{option}</option>
						{/each}
					</select>
					<span>条</span>
				</div>
			{:else}
				<div></div>
			{/if}

			<!-- 分页导航和跳转 -->
			<div class="flex items-center gap-4">
				<!-- 分页导航 -->
				<div class="flex items-center gap-1">
					<!-- 首页按钮 -->
					{#if showFirstLast}
						<button
							onclick={() => handlePageChange(1)}
							disabled={currentPage === 1 || disabled}
							class="px-3 h-8 text-sm font-medium rounded-md transition-all bg-white border border-gray-300 text-gray-900 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
						>
							<Icon name="chevron-left" size={16} />
							<Icon name="chevron-left" size={16} class="-ml-1" />
						</button>
					{/if}

					<!-- 上一页按钮 -->
					<button
						onclick={() => handlePageChange(currentPage - 1)}
						disabled={currentPage === 1 || disabled}
						class="px-3 h-8 text-sm font-medium rounded-md transition-all bg-white border border-gray-300 text-gray-900 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
					>
						<Icon name="chevron-left" size={16} />
						上一页
					</button>

					<!-- 页码按钮 -->
					<div class="flex items-center gap-1">
						{#if visiblePages[0] > 1}
							{@const isFirstPageActive = 1 === currentPage}
							<Button
								variant={isFirstPageActive ? 'default' : 'outline'}
								size="sm"
								onclick={() => handlePageChange(1)}
								disabled={disabled}
								class="w-10 h-8"
								style={isFirstPageActive ? 'background-color: #2563eb !important; color: white !important;' : 'background-color: white !important; border: 1px solid #d1d5db !important; color: #111827 !important;'}
							>
								1
							</Button>
							{#if visiblePages[0] > 2}
								<span class="px-2 text-gray-400">...</span>
							{/if}
						{/if}

						{#each visiblePages as page}
							{@const isCurrentPage = page === currentPage}
							<button
								onclick={() => handlePageChange(page)}
								disabled={disabled}
								class="w-10 h-8 text-sm font-medium rounded-md transition-all disabled:opacity-50 {isCurrentPage ? 'bg-blue-600 text-white' : 'bg-white border border-gray-300 text-gray-900 hover:bg-gray-50'}"
								data-debug={`page-${page}-current-${currentPage}-active-${isCurrentPage}`}
							>
								{page}
							</button>
						{/each}

						{#if visiblePages[visiblePages.length - 1] < totalPages}
							{#if visiblePages[visiblePages.length - 1] < totalPages - 1}
								<span class="px-2 text-gray-400">...</span>
							{/if}
							{@const isLastPageActive = totalPages === currentPage}
							<Button
								variant={isLastPageActive ? 'default' : 'outline'}
								size="sm"
								onclick={() => handlePageChange(totalPages)}
								disabled={disabled}
								class="w-10 h-8"
								style={isLastPageActive ? 'background-color: #2563eb !important; color: white !important;' : 'background-color: white !important; border: 1px solid #d1d5db !important; color: #111827 !important;'}
							>
								{totalPages}
							</Button>
						{/if}
					</div>

					<!-- 下一页按钮 -->
					<button
						onclick={() => handlePageChange(currentPage + 1)}
						disabled={currentPage === totalPages || disabled}
						class="px-3 h-8 text-sm font-medium rounded-md transition-all bg-white border border-gray-300 text-gray-900 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
					>
						下一页
						<Icon name="chevron-right" size={16} />
					</button>

					<!-- 末页按钮 -->
					{#if showFirstLast}
						<button
							onclick={() => handlePageChange(totalPages)}
							disabled={currentPage === totalPages || disabled}
							class="px-3 h-8 text-sm font-medium rounded-md transition-all bg-white border border-gray-300 text-gray-900 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
						>
							<Icon name="chevron-right" size={16} class="-mr-1" />
							<Icon name="chevron-right" size={16} />
						</button>
					{/if}
				</div>

				<!-- 快速跳转 -->
				{#if totalPages > 1}
					<div class="flex items-center gap-2 text-sm text-gray-600">
						<span>跳转到第</span>
						<input
							type="number"
							bind:value={jumpToPage}
							onkeydown={handleJumpToPage}
							placeholder="1"
							min="1"
							max={totalPages}
							disabled={disabled}
							class="w-16 h-8 px-2 py-1 text-center text-sm border border-gray-300 rounded 
							       focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent 
							       disabled:opacity-50 disabled:cursor-not-allowed"
						/>
						<span>页</span>
						<Button
							variant="outline"
							size="sm"
							onclick={handleJumpClick}
							disabled={disabled || !jumpToPage || isNaN(parseInt(jumpToPage))}
							class="h-8 px-2 text-xs"
						>
							跳转
						</Button>
					</div>
				{:else}
					<div></div>
				{/if}
			</div>
		</div>
	{/if}
	{/if}
</div>

<style>
	/* 确保按钮样式一致 */
	:global(.pagination-button) {
		@apply transition-all duration-150;
	}
</style>
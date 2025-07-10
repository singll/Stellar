<!--
分页组件
提供页码导航和翻页功能
-->
<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	interface Props {
		currentPage: number;
		totalPages: number;
		total: number;
		pageSize: number;
		showInfo?: boolean;
		maxVisiblePages?: number;
	}

	let {
		currentPage,
		totalPages,
		total,
		pageSize,
		showInfo = true,
		maxVisiblePages = 5
	}: Props = $props();

	const dispatch = createEventDispatcher<{
		pageChange: number;
	}>();

	// 计算显示的页码范围
	let visiblePages = $derived(() => {
		const halfVisible = Math.floor(maxVisiblePages / 2);
		let start = Math.max(1, currentPage - halfVisible);
		let end = Math.min(totalPages, start + maxVisiblePages - 1);

		// 调整起始位置以确保显示足够的页码
		if (end - start + 1 < maxVisiblePages) {
			start = Math.max(1, end - maxVisiblePages + 1);
		}

		const pages = [];
		for (let i = start; i <= end; i++) {
			pages.push(i);
		}
		return pages;
	});

	// 计算当前页的数据范围
	let dataRange = $derived(() => {
		const start = (currentPage - 1) * pageSize + 1;
		const end = Math.min(currentPage * pageSize, total);
		return { start, end };
	});

	function goToPage(page: number) {
		if (page < 1 || page > totalPages || page === currentPage) return;
		dispatch('pageChange', page);
	}

	function goToPrevious() {
		if (currentPage > 1) {
			goToPage(currentPage - 1);
		}
	}

	function goToNext() {
		if (currentPage < totalPages) {
			goToPage(currentPage + 1);
		}
	}
</script>

<div
	class="flex items-center justify-between px-4 py-3 sm:px-6 border-t border-gray-200 dark:border-gray-700"
>
	<!-- 信息显示 -->
	{#if showInfo}
		<div class="flex-1 flex justify-between sm:hidden">
			<button
				onclick={goToPrevious}
				disabled={currentPage <= 1}
				class="relative inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 text-sm font-medium rounded-md text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
			>
				上一页
			</button>
			<button
				onclick={goToNext}
				disabled={currentPage >= totalPages}
				class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 text-sm font-medium rounded-md text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
			>
				下一页
			</button>
		</div>

		<div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
			<div>
				<p class="text-sm text-gray-700 dark:text-gray-300">
					显示第 <span class="font-medium">{dataRange.start}</span> 到
					<span class="font-medium">{dataRange.end}</span>
					项， 共 <span class="font-medium">{total}</span> 项
				</p>
			</div>
			<div>
				<nav
					class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px"
					aria-label="Pagination"
				>
					<!-- 上一页 -->
					<button
						onclick={goToPrevious}
						disabled={currentPage <= 1}
						class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm font-medium text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<span class="sr-only">上一页</span>
						<i class="fas fa-chevron-left w-5 h-5" aria-hidden="true"></i>
					</button>

					<!-- 首页 -->
					{#if visiblePages[0] > 1}
						<button
							onclick={() => goToPage(1)}
							class="relative inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700"
						>
							1
						</button>
						{#if visiblePages[0] > 2}
							<span
								class="relative inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm font-medium text-gray-700 dark:text-gray-200"
							>
								...
							</span>
						{/if}
					{/if}

					<!-- 可见页码 -->
					{#each visiblePages as page}
						<button
							onclick={() => goToPage(page)}
							class="relative inline-flex items-center px-4 py-2 border text-sm font-medium {page ===
							currentPage
								? 'z-10 bg-blue-50 dark:bg-blue-900 border-blue-500 text-blue-600 dark:text-blue-300'
								: 'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'}"
						>
							{page}
						</button>
					{/each}

					<!-- 尾页 -->
					{#if visiblePages[visiblePages.length - 1] < totalPages}
						{#if visiblePages[visiblePages.length - 1] < totalPages - 1}
							<span
								class="relative inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm font-medium text-gray-700 dark:text-gray-200"
							>
								...
							</span>
						{/if}
						<button
							onclick={() => goToPage(totalPages)}
							class="relative inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700"
						>
							{totalPages}
						</button>
					{/if}

					<!-- 下一页 -->
					<button
						onclick={goToNext}
						disabled={currentPage >= totalPages}
						class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm font-medium text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<span class="sr-only">下一页</span>
						<i class="fas fa-chevron-right w-5 h-5" aria-hidden="true"></i>
					</button>
				</nav>
			</div>
		</div>
	{/if}
</div>

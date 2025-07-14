<script lang="ts">
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';

	let {
		data = [],
		columns = [],
		options = {}
	}: {
		data?: any[];
		columns?: any[];
		options?: any;
	} = $props();

	// 简化的数据表实现，避免 svelte-headless-table 的类型问题
	let currentPage = $state(1);
	let pageSize = $state(10);

	let paginatedData = $derived(() => {
		const start = (currentPage - 1) * pageSize;
		const end = start + pageSize;
		return data.slice(start, end);
	});

	let totalPages = $derived(() => Math.ceil(data.length / pageSize));

	// 添加用于模板访问的数组变量
	let paginatedDataArray = $derived(paginatedData());
	let totalPagesNumber = $derived(totalPages());
</script>

<div class="rounded-md border">
	<Table>
		<TableHeader>
			<TableRow>
				{#each columns as column}
					<TableHead>{column.header}</TableHead>
				{/each}
			</TableRow>
		</TableHeader>
		<TableBody>
			{#each paginatedDataArray as row, i}
				<TableRow>
					{#each columns as column}
						<TableCell>
							{#if column.cell}
								{@render column.cell(row)}
							{:else}
								{(row as any)[column.accessorKey] || ''}
							{/if}
						</TableCell>
					{/each}
				</TableRow>
			{/each}
		</TableBody>
	</Table>
</div>

<!-- 简单的分页控件 -->
{#if totalPagesNumber > 1}
	<div class="flex items-center justify-between px-2 py-4">
		<div class="text-sm text-gray-700">
			显示 {(currentPage - 1) * pageSize + 1} 到 {Math.min(currentPage * pageSize, data.length)} 项，共
			{data.length} 项
		</div>
		<div class="flex gap-2">
			<button
				class="px-3 py-1 border rounded disabled:opacity-50"
				disabled={currentPage === 1}
				onclick={() => currentPage--}
			>
				上一页
			</button>
			<span class="px-3 py-1">第 {currentPage} 页，共 {totalPagesNumber} 页</span>
			<button
				class="px-3 py-1 border rounded disabled:opacity-50"
				disabled={currentPage === totalPagesNumber}
				onclick={() => currentPage++}
			>
				下一页
			</button>
		</div>
	</div>
{/if}

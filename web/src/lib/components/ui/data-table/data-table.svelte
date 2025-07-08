<script lang="ts">
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import { createTable, type TableOptions, type ColumnDef } from 'svelte-headless-table';
	import { addSortBy, addTableFilter, addPagination } from 'svelte-headless-table/plugins';

	export let data: any[] = [];
	export let columns: ColumnDef<any>[] = [];
	export let options: Partial<TableOptions<any>> = {};

	const table = createTable(data, {
		columns,
		...options,
		plugins: [
			addSortBy(),
			addTableFilter(),
			addPagination({ initialPageSize: 10 }),
			...(options.plugins || [])
		]
	});

	const { headerRows, pageRows, tableAttrs, tableBodyAttrs, pluginStates } = table;
	const { pageSize, pageIndex } = pluginStates.pagination;
</script>

<div class="rounded-md border">
	<Table {...tableAttrs}>
		<TableHeader>
			{#each $headerRows as headerRow}
				<TableRow>
					{#each headerRow.cells as cell}
						<TableHead {...cell.attrs}>
							{#if cell.column.sort}
								<button
									class="inline-flex items-center gap-2"
									on:click={() => cell.column.sort?.toggle()}
								>
									{cell.column.header}
									{#if cell.column.sort?.order === 'asc'}
										<span>↑</span>
									{:else if cell.column.sort?.order === 'desc'}
										<span>↓</span>
									{/if}
								</button>
							{:else}
								{cell.column.header}
							{/if}
						</TableHead>
					{/each}
				</TableRow>
			{/each}
		</TableHeader>
		<TableBody {...tableBodyAttrs}>
			{#each $pageRows as row}
				<TableRow>
					{#each row.cells as cell}
						<TableCell {...cell.attrs}>
							{cell.value}
						</TableCell>
					{/each}
				</TableRow>
			{/each}
		</TableBody>
	</Table>
</div>

<div class="flex items-center justify-between space-x-2 py-4">
	<div class="flex-1 text-sm text-muted-foreground">
		第 {$pageIndex * $pageSize + 1} - {Math.min(($pageIndex + 1) * $pageSize, data.length)} 条，共 {data.length}
		条
	</div>
	<div class="space-x-2">
		<button
			class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2"
			disabled={$pageIndex === 0}
			on:click={() => pluginStates.pagination.previousPage()}
		>
			上一页
		</button>
		<button
			class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2"
			disabled={($pageIndex + 1) * $pageSize >= data.length}
			on:click={() => pluginStates.pagination.nextPage()}
		>
			下一页
		</button>
	</div>
</div>

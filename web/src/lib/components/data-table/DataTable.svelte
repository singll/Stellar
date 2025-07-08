<!-- DataTable.svelte -->
<script lang="ts">
	import { Table } from '$lib/components/ui/table';
	import type { TableHeader } from './types';
	import { createEventDispatcher } from 'svelte';
	import { fade } from 'svelte/transition';

	export let data: any[] = [];
	export let headers: TableHeader[] = [];
	export let pageSize: number = 10;
	export let currentPage: number = 1;
	export let totalItems: number = 0;
	export let loading: boolean = false;

	const dispatch = createEventDispatcher<{
		sort: { column: string; direction: 'asc' | 'desc' };
		page: { page: number };
		filter: { column: string; value: string };
	}>();

	let sortColumn: string | null = null;
	let sortDirection: 'asc' | 'desc' = 'asc';

	function handleSort(column: string) {
		if (sortColumn === column) {
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			sortColumn = column;
			sortDirection = 'asc';
		}
		dispatch('sort', { column, direction: sortDirection });
	}

	function handlePageChange(page: number) {
		if (page !== currentPage) {
			dispatch('page', { page });
		}
	}

	$: totalPages = Math.ceil(totalItems / pageSize);
	$: pages = Array.from({ length: totalPages }, (_, i) => i + 1);
</script>

<div class="w-full">
	{#if loading}
		<div class="flex justify-center p-4" transition:fade>
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
		</div>
	{:else}
		<div class="rounded-md border">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						{#each headers as header}
							<Table.Head
								class="cursor-pointer"
								on:click={() => header.sortable && handleSort(header.key)}
							>
								<div class="flex items-center gap-2">
									{header.label}
									{#if header.sortable && sortColumn === header.key}
										<span class="text-xs">
											{sortDirection === 'asc' ? '↑' : '↓'}
										</span>
									{/if}
								</div>
							</Table.Head>
						{/each}
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data as row}
						<Table.Row>
							{#each headers as header}
								<Table.Cell>
									{#if header.format}
										{@html header.format(row[header.key])}
									{:else}
										{row[header.key]}
									{/if}
								</Table.Cell>
							{/each}
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="flex items-center justify-between px-2 py-4">
				<button
					class="btn btn-sm"
					disabled={currentPage === 1}
					on:click={() => handlePageChange(currentPage - 1)}
				>
					Previous
				</button>
				<div class="flex gap-2">
					{#each pages as page}
						<button
							class="btn btn-sm {currentPage === page ? 'btn-primary' : ''}"
							on:click={() => handlePageChange(page)}
						>
							{page}
						</button>
					{/each}
				</div>
				<button
					class="btn btn-sm"
					disabled={currentPage === totalPages}
					on:click={() => handlePageChange(currentPage + 1)}
				>
					Next
				</button>
			</div>
		{/if}
	{/if}
</div>

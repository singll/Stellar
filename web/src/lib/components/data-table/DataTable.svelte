<!-- DataTable.svelte -->
<script lang="ts">
	import {
		Root as TableRoot,
		Header as TableHeaderRow,
		Row as TableRow,
		Head as TableHead,
		Body as TableBody,
		Cell as TableCell
	} from '$lib/components/ui/table';
	import type { TableHeader } from './types';
	import { createEventDispatcher } from 'svelte';
	import { fade } from 'svelte/transition';

	let {
		data = [],
		headers = [],
		pageSize = 10,
		currentPage = 1,
		totalItems = 0,
		loading = false
	}: {
		data?: any[];
		headers?: TableHeader[];
		pageSize?: number;
		currentPage?: number;
		totalItems?: number;
		loading?: boolean;
	} = $props();

	const dispatch = createEventDispatcher<{
		sort: { column: string; direction: 'asc' | 'desc' };
		page: { page: number };
		filter: { column: string; value: string };
	}>();

	let sortColumn = $state<string | null>(null);
	let sortDirection = $state<'asc' | 'desc'>('asc');

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

	let totalPages = $derived(Math.ceil(totalItems / pageSize));
	let pages = $derived(Array.from({ length: totalPages }, (_, i) => i + 1));
</script>

<div class="w-full">
	{#if loading}
		<div class="flex justify-center p-4" transition:fade>
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
		</div>
	{:else}
		<div class="rounded-md border">
			<TableRoot>
				<TableHeaderRow>
					<TableRow>
						{#each headers as header}
							<TableHead
								class="cursor-pointer"
								onclick={() => header.sortable && handleSort(header.key)}
							>
								<div class="flex items-center gap-2">
									{header.label}
									{#if header.sortable && sortColumn === header.key}
										<span class="text-xs">
											{sortDirection === 'asc' ? '↑' : '↓'}
										</span>
									{/if}
								</div>
							</TableHead>
						{/each}
					</TableRow>
				</TableHeaderRow>
				<TableBody>
					{#each data as row}
						<TableRow>
							{#each headers as header}
								<TableCell>
									{#if header.format}
										{@html header.format(row[header.key])}
									{:else}
										{row[header.key]}
									{/if}
								</TableCell>
							{/each}
						</TableRow>
					{/each}
				</TableBody>
			</TableRoot>
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="flex items-center justify-between px-2 py-4">
				<button
					class="btn btn-sm"
					disabled={currentPage === 1}
					onclick={() => handlePageChange(currentPage - 1)}
				>
					Previous
				</button>
				<div class="flex gap-2">
					{#each pages as page}
						<button
							class="btn btn-sm {currentPage === page ? 'btn-primary' : ''}"
							onclick={() => handlePageChange(page)}
						>
							{page}
						</button>
					{/each}
				</div>
				<button
					class="btn btn-sm"
					disabled={currentPage === totalPages}
					onclick={() => handlePageChange(currentPage + 1)}
				>
					Next
				</button>
			</div>
		{/if}
	{/if}
</div>

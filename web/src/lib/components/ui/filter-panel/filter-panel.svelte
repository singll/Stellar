<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		Select,
		SelectContent,
		SelectItem,
		SelectTrigger,
		SelectValue
	} from '$lib/components/ui/select';
	import { cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';

	let {
		filters = [],
		className = ''
	}: {
		filters?: Array<{
			id: string;
			label: string;
			type: 'text' | 'select';
			value: string;
			options?: Array<{ value: string; label: string }>;
		}>;
		className?: string;
	} = $props();

	const dispatch = createEventDispatcher<{
		filter: { filters: typeof filters };
		reset: undefined;
	}>();

	function handleFilter() {
		dispatch('filter', { filters });
	}

	function handleReset() {
		filters = filters.map((filter) => ({ ...filter, value: '' }));
		dispatch('reset');
	}
</script>

<div class={cn('space-y-4 p-4', className)}>
	<div class="grid gap-4">
		{#each filters as filter}
			<div class="grid gap-2">
				<Label for={filter.id}>{filter.label}</Label>
				{#if filter.type === 'select'}
					<Select bind:value={filter.value}>
						<SelectTrigger>
							<SelectValue placeholder="选择..." />
						</SelectTrigger>
						<SelectContent>
							{#each filter.options || [] as option}
								<SelectItem value={option.value}>
									{option.label}
								</SelectItem>
							{/each}
						</SelectContent>
					</Select>
				{:else}
					<Input id={filter.id} bind:value={filter.value} placeholder={`输入${filter.label}...`} />
				{/if}
			</div>
		{/each}
	</div>

	<div class="flex items-center space-x-2">
		<Button variant="outline" onclick={handleReset}>重置</Button>
		<Button onclick={handleFilter}>筛选</Button>
	</div>
</div>

<!-- SearchFilter.svelte -->
<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { createEventDispatcher } from 'svelte';

	interface Filter {
		id: string;
		label: string;
		type: 'text' | 'number' | 'boolean' | 'select';
		value: string | boolean;
		options?: { value: string; label: string }[];
	}

	export let filters: Filter[] = [];

	const dispatch = createEventDispatcher<{
		change: { id: string; value: string | boolean };
	}>();

	function handleChange(id: string, value: string | boolean) {
		dispatch('change', { id, value });
	}
</script>

<div class="space-y-4">
	{#each filters as filter}
		<div>
			<Label for={filter.id}>{filter.label}</Label>
			{#if filter.type === 'text' || filter.type === 'number'}
				<Input
					type={filter.type}
					id={filter.id}
					value={filter.value || ''}
					on:input={(e) => handleChange(filter.id, e.currentTarget.value)}
				/>
			{:else if filter.type === 'boolean'}
				<input
					type="checkbox"
					id={filter.id}
					checked={filter.value || false}
					on:change={(e) => handleChange(filter.id, e.currentTarget.checked)}
				/>
			{:else if filter.type === 'select' && filter.options}
				<select
					id={filter.id}
					value={filter.value || ''}
					on:change={(e) => handleChange(filter.id, e.currentTarget.value)}
				>
					<option value="">请选择</option>
					{#each filter.options as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			{/if}
		</div>
	{/each}
</div>

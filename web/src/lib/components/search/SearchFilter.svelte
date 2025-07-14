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

	let {
		filters = []
	}: {
		filters?: Filter[];
	} = $props();

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
					value={typeof filter.value === 'string' ? filter.value : ''}
					oninput={(e) => handleChange(filter.id, e.currentTarget.value)}
				/>
			{:else if filter.type === 'boolean'}
				<input
					type="checkbox"
					id={filter.id}
					checked={typeof filter.value === 'boolean' ? filter.value : false}
					onchange={(e) => handleChange(filter.id, e.currentTarget.checked)}
				/>
			{:else if filter.type === 'select' && filter.options}
				<select
					id={filter.id}
					value={filter.value || ''}
					onchange={(e) => handleChange(filter.id, e.currentTarget.value)}
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

<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLSelectAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	export type SelectProps = HTMLSelectAttributes & {
		class?: string;
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		children?: Snippet;
		options?: Array<{ value: string; label: string; disabled?: boolean }>;
	};

	let {
		class: className,
		value = $bindable(''),
		placeholder = '请选择...',
		disabled = false,
		children,
		options = [],
		...restProps
	}: SelectProps = $props();

	// 简化的选择框样式
	const selectClasses =
		'flex h-10 w-full items-center justify-between rounded-md border border-gray-300 bg-white px-3 py-2 text-sm ring-offset-background placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50';
</script>

<select class={cn(selectClasses, className)} bind:value {disabled} {...restProps}>
	{#if placeholder}
		<option value="" disabled>{placeholder}</option>
	{/if}

	{#if options.length > 0}
		{#each options as option}
			<option value={option.value} disabled={option.disabled}>
				{option.label}
			</option>
		{/each}
	{:else if children}
		{@render children()}
	{/if}
</select>

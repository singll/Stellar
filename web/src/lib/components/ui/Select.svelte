<!--
简单的选择器组件
-->
<script lang="ts">
	import { cn } from '$lib/utils';

	interface Props {
		value?: string;
		options: Array<{ value: string; label: string }>;
		placeholder?: string;
		disabled?: boolean;
		class?: string;
		onselect?: (value: string) => void;
	}

	let {
		value = $bindable(''),
		options,
		placeholder,
		disabled = false,
		class: className = '',
		onselect
	}: Props = $props();

	function handleChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		value = target.value;
		onselect?.(value);
	}
</script>

<select
	bind:value
	{disabled}
	class={cn(
		'flex h-10 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-gray-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
		className
	)}
	onchange={handleChange}
>
	{#if placeholder}
		<option value="" disabled selected={value === ''}>{placeholder}</option>
	{/if}
	{#each options as option}
		<option value={option.value}>{option.label}</option>
	{/each}
</select>

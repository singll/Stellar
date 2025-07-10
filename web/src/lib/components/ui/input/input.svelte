<script lang="ts">
	import type { HTMLInputAttributes, HTMLInputTypeAttribute } from 'svelte/elements';
	import { cn } from '$lib/utils';

	type InputType = Exclude<HTMLInputTypeAttribute, 'file'>;

	type Props = HTMLInputAttributes & {
		type?: InputType;
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		readonly?: boolean;
		class?: string;
		name?: string;
		id?: string;
		ref?: HTMLInputElement | null;
	};

	let {
		ref = $bindable(null),
		value = $bindable(''),
		type = 'text',
		placeholder,
		disabled = false,
		readonly = false,
		class: className,
		name,
		id,
		...restProps
	}: Props = $props();

	// 简化的输入框样式
	const inputClasses =
		'flex h-9 w-full rounded-md border border-gray-300 bg-white px-3 py-1 text-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-gray-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50';
</script>

<input
	bind:this={ref}
	bind:value
	{type}
	{placeholder}
	{disabled}
	{readonly}
	{name}
	{id}
	class={cn(inputClasses, className)}
	{...restProps}
/>

<!--
文本域组件
多行文本输入组件
-->
<script lang="ts">
	import { cn } from '$lib/utils';

	interface Props {
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		readonly?: boolean;
		required?: boolean;
		rows?: number;
		cols?: number;
		class?: string;
		id?: string;
		name?: string;
		resize?: 'none' | 'vertical' | 'horizontal' | 'both';
		onchange?: (event: Event) => void;
		oninput?: (event: Event) => void;
		onkeydown?: (event: KeyboardEvent) => void;
		onkeyup?: (event: KeyboardEvent) => void;
		onfocus?: (event: FocusEvent) => void;
		onblur?: (event: FocusEvent) => void;
	}

	let {
		value = $bindable(''),
		placeholder,
		disabled = false,
		readonly = false,
		required = false,
		rows = 3,
		cols,
		class: className = '',
		id,
		name,
		resize = 'vertical',
		onchange,
		oninput,
		onkeydown,
		onkeyup,
		onfocus,
		onblur
	}: Props = $props();

	const baseClasses =
		'block w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:focus:border-blue-400 dark:focus:ring-blue-400 disabled:opacity-50 disabled:cursor-not-allowed';

	let textareaClasses = $derived(cn(baseClasses, `resize-${resize}`, className));

	function handleChange(event: Event) {
		const target = event.target as HTMLTextAreaElement;
		value = target.value;
		onchange?.(event);
	}

	function handleInput(event: Event) {
		const target = event.target as HTMLTextAreaElement;
		value = target.value;
		oninput?.(event);
	}
</script>

<textarea
	{id}
	{name}
	{placeholder}
	{disabled}
	{readonly}
	{required}
	{rows}
	{cols}
	class={textareaClasses}
	bind:value
	onchange={handleChange}
	oninput={handleInput}
	{onkeydown}
	{onkeyup}
	{onfocus}
	{onblur}
></textarea>

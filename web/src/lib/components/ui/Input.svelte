<!--
输入框组件
基础输入框组件
-->
<script lang="ts">
	import { cn } from '$lib/utils';

	interface Props {
		value?: string | number;
		type?: 'text' | 'password' | 'email' | 'number' | 'tel' | 'url' | 'datetime-local';
		placeholder?: string;
		disabled?: boolean;
		readonly?: boolean;
		required?: boolean;
		min?: number;
		max?: number;
		step?: number;
		class?: string;
		id?: string;
		name?: string;
		onchange?: (event: Event) => void;
		oninput?: (event: Event) => void;
		onkeydown?: (event: KeyboardEvent) => void;
		onkeyup?: (event: KeyboardEvent) => void;
		onfocus?: (event: FocusEvent) => void;
		onblur?: (event: FocusEvent) => void;
	}

	let {
		value = $bindable(''),
		type = 'text',
		placeholder,
		disabled = false,
		readonly = false,
		required = false,
		min,
		max,
		step,
		class: className = '',
		id,
		name,
		onchange,
		oninput,
		onkeydown,
		onkeyup,
		onfocus,
		onblur
	}: Props = $props();

	const baseClasses =
		'block w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:focus:border-blue-400 dark:focus:ring-blue-400 disabled:opacity-50 disabled:cursor-not-allowed';

	let inputClasses = $derived(cn(baseClasses, className));

	function handleChange(event: Event) {
		const target = event.target as HTMLInputElement;
		if (type === 'number') {
			value = target.value === '' ? '' : Number(target.value);
		} else {
			value = target.value;
		}
		onchange?.(event);
	}

	function handleInput(event: Event) {
		const target = event.target as HTMLInputElement;
		if (type === 'number') {
			value = target.value === '' ? '' : Number(target.value);
		} else {
			value = target.value;
		}
		oninput?.(event);
	}
</script>

<input
	{id}
	{name}
	{type}
	{placeholder}
	{disabled}
	{readonly}
	{required}
	{min}
	{max}
	{step}
	class={inputClasses}
	bind:value
	onchange={handleChange}
	oninput={handleInput}
	{onkeydown}
	{onkeyup}
	{onfocus}
	{onblur}
/>

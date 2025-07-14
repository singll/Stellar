<!--
搜索输入框组件
带有搜索图标和清除按钮的搜索输入框
-->
<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	interface Props {
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		class?: string;
		onenter?: () => void;
	}

	let {
		value = $bindable(''),
		placeholder = '搜索...',
		disabled = false,
		class: className = '',
		onenter
	}: Props = $props();

	const dispatch = createEventDispatcher<{
		search: string;
		clear: void;
		enter: string;
	}>();

	function handleInput(event: Event) {
		const target = event.target as HTMLInputElement;
		value = target.value;
		dispatch('search', value);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			onenter?.();
			dispatch('enter', value);
			dispatch('search', value);
		}
	}

	function handleClear() {
		value = '';
		dispatch('clear');
		dispatch('search', '');
	}
</script>

<div class="relative {className}">
	<div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
		<i class="fas fa-search text-gray-400 text-sm"></i>
	</div>

	<input
		type="text"
		bind:value
		{placeholder}
		{disabled}
		class="block w-full pl-10 pr-10 py-2 border border-gray-300 dark:border-gray-600 rounded-md leading-5 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
		oninput={handleInput}
		onkeydown={handleKeydown}
	/>

	{#if value}
		<div class="absolute inset-y-0 right-0 pr-3 flex items-center">
			<button
				type="button"
				onclick={handleClear}
				class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 focus:outline-none focus:text-gray-600 dark:focus:text-gray-300"
				aria-label="清除搜索"
			>
				<i class="fas fa-times text-sm"></i>
			</button>
		</div>
	{/if}
</div>

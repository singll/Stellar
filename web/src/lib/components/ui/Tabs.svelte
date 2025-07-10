<!--
标签页组件
提供标签页切换功能
-->
<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	interface Tab {
		id: string;
		label: string;
		icon?: string;
		disabled?: boolean;
	}

	interface Props {
		tabs: Tab[];
		activeTab?: string;
		class?: string;
		children?: any;
	}

	let {
		tabs,
		activeTab = $bindable(tabs[0]?.id || ''),
		class: className = '',
		children
	}: Props = $props();

	const dispatch = createEventDispatcher<{
		change: string;
	}>();

	function handleTabClick(tabId: string) {
		if (tabs.find((tab) => tab.id === tabId)?.disabled) return;

		activeTab = tabId;
		dispatch('change', tabId);
	}

	function handleKeydown(event: KeyboardEvent, tabId: string) {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			handleTabClick(tabId);
		}
	}
</script>

<div class="w-full {className}">
	<div class="border-b border-gray-200 dark:border-gray-700">
		<nav class="-mb-px flex space-x-8" aria-label="Tabs">
			{#each tabs as tab}
				<button
					type="button"
					class="whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors flex items-center space-x-2
						{activeTab === tab.id
						? 'border-blue-500 text-blue-600 dark:text-blue-400'
						: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300 dark:hover:border-gray-600'}
						{tab.disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer'}"
					onclick={() => handleTabClick(tab.id)}
					onkeydown={(e) => handleKeydown(e, tab.id)}
					disabled={tab.disabled}
					aria-selected={activeTab === tab.id}
					role="tab"
					tabindex={activeTab === tab.id ? 0 : -1}
				>
					{#if tab.icon}
						<span>{tab.icon}</span>
					{/if}
					<span>{tab.label}</span>
				</button>
			{/each}
		</nav>
	</div>

	<div class="mt-4">
		{@render children?.()}
	</div>
</div>

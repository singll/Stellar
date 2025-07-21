<script lang="ts">
	import { getContext } from 'svelte';
	import type { Snippet } from 'svelte';

	interface Props {
		children?: Snippet;
		class?: string;
	}

	let { children, class: className = '' }: Props = $props();
	
	// Get popover context
	const popoverContext = getContext('popover') as { 
		open: boolean; 
		close: () => void; 
		toggle: () => void; 
	} | undefined;

	const isOpen = popoverContext?.open ?? false;

	function handleContentClick(event: MouseEvent) {
		event.stopPropagation();
	}
</script>

{#if isOpen}
	<div 
		class="absolute z-50 mt-1 rounded-md border bg-white shadow-md {className}"
		onclick={handleContentClick}
		role="dialog"
		aria-modal="true"
	>
		{#if children}
			{@render children()}
		{/if}
	</div>
{/if}

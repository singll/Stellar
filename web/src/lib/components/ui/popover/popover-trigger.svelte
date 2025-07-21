<script lang="ts">
	import { getContext } from 'svelte';
	import type { Snippet } from 'svelte';

	interface Props {
		children?: Snippet<[{ builder: { 'data-trigger': string; onclick: (e: MouseEvent) => void } }]>;
		asChild?: boolean;
	}

	let { children, asChild = false }: Props = $props();
	
	// Get popover context
	const popoverContext = getContext('popover') as { 
		open: boolean; 
		close: () => void; 
		toggle: () => void; 
	} | undefined;

	function handleClick(event: MouseEvent) {
		event.stopPropagation();
		if (popoverContext) {
			popoverContext.toggle();
		}
	}

	// Create builder object for trigger functionality
	const builder = {
		'data-trigger': 'popover-trigger',
		onclick: handleClick
	};
</script>

{#if children}
	{@render children({ builder })}
{/if}

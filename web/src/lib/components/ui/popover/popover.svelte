<script lang="ts">
	import { createEventDispatcher, setContext } from 'svelte';
	import type { Snippet } from 'svelte';

	interface Props {
		open?: boolean;
		children?: Snippet;
	}

	let { open = $bindable(false), children }: Props = $props();
	
	const dispatch = createEventDispatcher();
	
	// Create context for popover state
	setContext('popover', {
		open,
		close: () => { open = false; },
		toggle: () => { open = !open; }
	});

	// Close on outside click
	function handleOutsideClick(event: MouseEvent) {
		if (open && event.target) {
			const target = event.target as Element;
			const popoverElement = target.closest('.popover-container');
			if (!popoverElement) {
				open = false;
			}
		}
	}
</script>

<svelte:window onclick={handleOutsideClick} />

<div class="popover-container relative">
	{#if children}
		{@render children()}
	{/if}
</div>

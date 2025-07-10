<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	interface Props extends HTMLAttributes<HTMLDivElement> {
		children?: Snippet;
		ref?: HTMLDivElement | null;
	}

	let { ref = $bindable(null), class: className, children, ...restProps }: Props = $props();
</script>

<div bind:this={ref} class={cn('flex items-center p-6 pt-0', className)} {...restProps}>
	{#if children}
		{@render children()}
	{/if}
</div>

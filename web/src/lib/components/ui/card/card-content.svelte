<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	interface Props extends HTMLAttributes<HTMLDivElement> {
		children?: Snippet;
		ref?: HTMLDivElement | null;
	}

	let { ref = $bindable(null), class: className, children, ...restProps }: Props = $props();
</script>

<div bind:this={ref} class={cn('px-6', className)} {...restProps}>
	{#if children}
		{@render children()}
	{/if}
</div>

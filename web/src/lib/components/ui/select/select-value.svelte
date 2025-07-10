<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	interface Props extends HTMLAttributes<HTMLSpanElement> {
		children?: Snippet;
		placeholder?: string;
		ref?: HTMLSpanElement | null;
	}

	let {
		ref = $bindable(null),
		class: className,
		children,
		placeholder,
		...restProps
	}: Props = $props();
</script>

<span bind:this={ref} class={cn('block truncate', className)} {...restProps}>
	{#if children}
		{@render children()}
	{:else if placeholder}
		<span class="text-muted-foreground">{placeholder}</span>
	{/if}
</span>

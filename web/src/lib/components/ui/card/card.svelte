<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	interface Props extends HTMLAttributes<HTMLDivElement> {
		children?: Snippet;
		ref?: HTMLDivElement | null;
	}

	let { ref = $bindable(null), class: className, children, ...restProps }: Props = $props();

	// 简化的卡片样式
	const cardClasses =
		'bg-white text-gray-900 flex flex-col gap-4 rounded-lg border border-gray-200 p-6 shadow-sm';
</script>

<div bind:this={ref} class={cn(cardClasses, className)} {...restProps}>
	{#if children}
		{@render children()}
	{/if}
</div>

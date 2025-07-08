<script lang="ts">
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	interface Props extends WithElementRef<HTMLAttributes<HTMLDivElement>> {
		children?: Snippet;
	}

	let { ref = $bindable(null), class: className, children, ...restProps }: Props = $props();
</script>

<div
	bind:this={ref}
	data-slot="card-footer"
	class={cn('[.border-t]:pt-6 flex items-center px-6', className)}
	{...restProps}
>
	{@render children?.()}
</div>

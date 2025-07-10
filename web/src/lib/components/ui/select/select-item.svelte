<script lang="ts">
	import { cn } from '$lib/utils';
	import Icon from '$lib/components/ui/Icon.svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	interface Props extends HTMLAttributes<HTMLDivElement> {
		children: Snippet;
		value: string;
		selected?: boolean;
		disabled?: boolean;
		ref?: HTMLDivElement | null;
	}

	let {
		ref = $bindable(null),
		class: className,
		children,
		value,
		selected = false,
		disabled = false,
		...restProps
	}: Props = $props();
</script>

<div
	bind:this={ref}
	class={cn(
		'relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
		className
	)}
	data-disabled={disabled}
	{...restProps}
>
	{#if selected}
		<span class="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
			<Icon name="check" class="h-4 w-4" />
		</span>
	{/if}
	{@render children()}
</div>

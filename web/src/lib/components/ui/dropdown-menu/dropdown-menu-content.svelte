<script lang="ts">
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils';
	import { fade } from 'svelte/transition';
	import type { Snippet } from 'svelte';

	interface Props extends DropdownMenuPrimitive.ContentProps {
		children: Snippet;
		align?: 'start' | 'end' | 'center';
		sideOffset?: number;
		ref?: HTMLDivElement | null;
	}

	let {
		ref = $bindable(null),
		class: className,
		children,
		align = 'center',
		sideOffset = 4,
		...restProps
	}: Props = $props();
</script>

<div
	bind:this={ref}
	class={cn(
		'z-50 min-w-[8rem] overflow-hidden rounded-md border bg-popover p-1 text-popover-foreground shadow-md data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
		className
	)}
	transition:fade={{ duration: 100 }}
>
	<DropdownMenuPrimitive.Content {align} {sideOffset} {...restProps}>
		{@render children()}
	</DropdownMenuPrimitive.Content>
</div>

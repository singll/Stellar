<script lang="ts">
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	interface Props extends DropdownMenuPrimitive.ItemProps {
		children: Snippet;
		disabled?: boolean;
		ref?: HTMLDivElement | null;
	}

	let {
		ref = $bindable(null),
		class: className,
		children,
		disabled = false,
		...restProps
	}: Props = $props();
</script>

<DropdownMenuPrimitive.Item
	bind:ref
	{disabled}
	class={cn(
		'relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
		className
	)}
	{...restProps}
>
	{@render children()}
</DropdownMenuPrimitive.Item>

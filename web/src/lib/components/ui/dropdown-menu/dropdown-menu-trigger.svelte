<script lang="ts">
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	interface Props {
		children: Snippet;
		class?: string;
		child?: boolean;
		ref?: HTMLButtonElement | null;
		[key: string]: any;
	}

	let {
		ref = $bindable(null),
		class: className,
		children,
		child = false,
		...restProps
	}: Props = $props();
</script>

<DropdownMenuPrimitive.Trigger
	bind:ref
	class={cn(
		child
			? ''
			: 'inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=open]:bg-accent data-[state=open]:text-accent-foreground',
		className
	)}
	{...restProps}
>
	{@render children()}
</DropdownMenuPrimitive.Trigger>

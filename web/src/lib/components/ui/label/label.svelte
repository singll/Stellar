<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Label as LabelPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	interface Props extends LabelPrimitive.RootProps {
		children: Snippet;
	}

	let { ref = $bindable(null), class: className, children, ...restProps }: Props = $props();
</script>

<LabelPrimitive.Root
	bind:ref
	data-slot="label"
	class={cn(
		'flex select-none items-center gap-2 text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-50 group-data-[disabled=true]:pointer-events-none group-data-[disabled=true]:opacity-50',
		className
	)}
	{...restProps}
>
	{@render children()}
</LabelPrimitive.Root>

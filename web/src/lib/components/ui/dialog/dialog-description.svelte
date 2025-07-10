<!--
  @component
  Dialog description component from shadcn-svelte
-->
<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';
	import type { Snippet } from 'svelte';

	interface Props extends DialogPrimitive.DescriptionProps {
		children: Snippet;
		ref?: HTMLParagraphElement | null;
	}

	let { ref = $bindable(null), class: className, children, ...restProps }: Props = $props();
</script>

<DialogPrimitive.Description
	bind:ref
	class={cn('text-sm text-muted-foreground', className)}
	{...restProps}
>
	{@render children()}
</DialogPrimitive.Description>

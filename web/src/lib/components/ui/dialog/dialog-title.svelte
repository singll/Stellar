<!--
  @component
  Dialog title component from shadcn-svelte
-->
<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';
	import type { Snippet } from 'svelte';

	interface Props extends DialogPrimitive.TitleProps {
		children: Snippet;
		ref?: HTMLHeadingElement | null;
	}

	let { ref = $bindable(null), class: className, children, ...restProps }: Props = $props();
</script>

<DialogPrimitive.Title
	bind:ref
	class={cn('text-lg font-semibold leading-none tracking-tight', className)}
	{...restProps}
>
	{@render children()}
</DialogPrimitive.Title>

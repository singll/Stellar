<!--
  @component
  Dialog trigger component from shadcn-svelte
-->
<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import type { Snippet } from 'svelte';

	interface Props extends DialogPrimitive.TriggerProps {
		children: Snippet;
		ref?: HTMLButtonElement | null;
	}

	let { ref = $bindable(null), children, ...restProps }: Props = $props();
</script>

<DialogPrimitive.Trigger bind:ref {...restProps}>
	{@render children()}
</DialogPrimitive.Trigger>

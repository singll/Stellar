<script lang="ts">
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';
	import type { Snippet } from 'svelte';

	interface Props extends DropdownMenuPrimitive.RootProps {
		children: Snippet;
		open?: boolean;
	}

	let { open = $bindable(false), children, ...restProps }: Props = $props();
</script>

<DropdownMenuPrimitive.Root bind:open {...restProps}>
	{@render children()}
</DropdownMenuPrimitive.Root>

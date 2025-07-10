<script lang="ts">
	import { cn } from '$lib/utils';
	import Icon from '$lib/components/ui/Icon.svelte';

	let { class: className, children, ...restProps }: { class?: string; children?: any } = $props();
</script>

<button
	class={cn(
		'flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 [&>span]:line-clamp-1',
		className
	)}
	{...restProps}
>
	{@render children?.()}
	<Icon name="chevron-down" class="h-4 w-4 opacity-50" />
</button>

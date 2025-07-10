<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { cn } from '$lib/utils';
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	type StatusType = 'success' | 'warning' | 'error' | 'info';

	interface Props extends HTMLAttributes<HTMLDivElement> {
		children: Snippet;
		status?: StatusType;
		pulse?: boolean;
		ref?: HTMLDivElement | null;
	}

	let {
		ref = $bindable(null),
		class: className,
		children,
		status = 'info',
		pulse = false,
		...restProps
	}: Props = $props();

	const getStatusColor = (status: StatusType) => {
		switch (status) {
			case 'success':
				return 'bg-success text-success-foreground';
			case 'warning':
				return 'bg-warning text-warning-foreground';
			case 'error':
				return 'bg-destructive text-destructive-foreground';
			case 'info':
				return 'bg-primary text-primary-foreground';
			default:
				return 'bg-primary text-primary-foreground';
		}
	};
</script>

<Badge
	bind:ref
	class={cn(getStatusColor(status), pulse && 'animate-pulse', className)}
	{...restProps}
>
	{@render children()}
</Badge>

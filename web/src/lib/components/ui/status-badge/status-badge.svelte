<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { cn } from '$lib/utils';

	let className: string | undefined = undefined;
	export { className as class };

	type StatusType = 'success' | 'warning' | 'error' | 'info';
	export let status: StatusType = 'info';
	export let pulse = false;

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

<Badge class={cn(getStatusColor(status), pulse && 'animate-pulse', className)} {...$$restProps}>
	<slot />
</Badge>

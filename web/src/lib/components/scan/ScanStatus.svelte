<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Progress } from '$lib/components/ui/progress';
	import { cn } from '$lib/utils';

	type ScanStatus = 'pending' | 'running' | 'completed' | 'failed';

	export let status: ScanStatus = 'pending';
	export let progress: number = 0;
	export let message: string = '';
	export let className: string = '';

	const getStatusVariant = (status: ScanStatus) => {
		switch (status) {
			case 'pending':
				return 'secondary';
			case 'running':
				return 'default';
			case 'completed':
				return 'default';
			case 'failed':
				return 'destructive';
			default:
				return 'default';
		}
	};

	const getStatusText = (status: ScanStatus) => {
		switch (status) {
			case 'pending':
				return '等待中';
			case 'running':
				return '扫描中';
			case 'completed':
				return '已完成';
			case 'failed':
				return '失败';
			default:
				return '未知';
		}
	};
</script>

<div class={cn('space-y-2', className)}>
	<div class="flex items-center justify-between">
		<Badge variant={getStatusVariant(status)}>
			{getStatusText(status)}
		</Badge>
		{#if status === 'running'}
			<span class="text-sm text-muted-foreground">{progress}%</span>
		{/if}
	</div>

	{#if status === 'running'}
		<Progress value={progress} />
	{/if}

	{#if message}
		<p class="text-sm text-muted-foreground">{message}</p>
	{/if}
</div>

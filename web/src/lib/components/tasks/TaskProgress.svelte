<!--
任务进度组件
显示任务的执行进度和详细信息
-->
<script lang="ts">
	import type { Task } from '$lib/types/task';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';

	interface Props {
		task: Task;
	}

	let { task }: Props = $props();

	// 计算执行时间
	let executionTime = $derived(() => {
		if (!task.startedAt) return null;

		const startTime = new Date(task.startedAt);
		const endTime = task.completedAt ? new Date(task.completedAt) : new Date();
		const diff = endTime.getTime() - startTime.getTime();

		const hours = Math.floor(diff / (1000 * 60 * 60));
		const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
		const seconds = Math.floor((diff % (1000 * 60)) / 1000);

		if (hours > 0) {
			return `${hours}h ${minutes}m ${seconds}s`;
		} else if (minutes > 0) {
			return `${minutes}m ${seconds}s`;
		} else {
			return `${seconds}s`;
		}
	});

	// 获取进度条颜色
	function getProgressColor(status: string) {
		switch (status) {
			case 'running':
				return 'bg-blue-500';
			case 'completed':
				return 'bg-green-500';
			case 'failed':
				return 'bg-red-500';
			case 'cancelled':
				return 'bg-yellow-500';
			default:
				return 'bg-gray-500';
		}
	}

	// 获取状态图标
	function getStatusIcon(status: string) {
		switch (status) {
			case 'pending':
				return 'fas fa-clock';
			case 'running':
				return 'fas fa-play';
			case 'completed':
				return 'fas fa-check';
			case 'failed':
				return 'fas fa-times';
			case 'cancelled':
				return 'fas fa-ban';
			default:
				return 'fas fa-question';
		}
	}

	// 获取状态文本
	function getStatusText(status: string) {
		switch (status) {
			case 'pending':
				return '等待中';
			case 'running':
				return '运行中';
			case 'completed':
				return '已完成';
			case 'failed':
				return '失败';
			case 'cancelled':
				return '已取消';
			default:
				return '未知';
		}
	}
</script>

<Card>
	<CardHeader>
		<CardTitle>执行进度</CardTitle>
	</CardHeader>
	<CardContent>
		<div class="space-y-4">
			<!-- 状态信息 -->
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<i class="{getStatusIcon(task.status)} text-gray-600 dark:text-gray-400"></i>
					<span class="font-medium">{getStatusText(task.status)}</span>
				</div>

				{#if task.progress !== undefined}
					<span class="text-sm text-gray-600 dark:text-gray-400">
						{task.progress}%
					</span>
				{/if}
			</div>

			<!-- 进度条 -->
			{#if task.progress !== undefined}
				<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
					<div
						class="{getProgressColor(task.status)} h-2 rounded-full transition-all duration-300"
						style="width: {task.progress}%"
					></div>
				</div>
			{/if}

			<!-- 时间信息 -->
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
				<div>
					<div class="text-gray-500 dark:text-gray-400">创建时间</div>
					<div class="font-medium">{new Date(task.createdAt).toLocaleString()}</div>
				</div>

				{#if task.startedAt}
					<div>
						<div class="text-gray-500 dark:text-gray-400">开始时间</div>
						<div class="font-medium">{new Date(task.startedAt).toLocaleString()}</div>
					</div>
				{/if}

				{#if task.completedAt}
					<div>
						<div class="text-gray-500 dark:text-gray-400">完成时间</div>
						<div class="font-medium">{new Date(task.completedAt).toLocaleString()}</div>
					</div>
				{/if}

				{#if executionTime}
					<div>
						<div class="text-gray-500 dark:text-gray-400">执行时间</div>
						<div class="font-medium">{executionTime}</div>
					</div>
				{/if}
			</div>

			<!-- 任务详情 -->
			<div class="border-t pt-4">
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
					<div>
						<div class="text-gray-500 dark:text-gray-400">任务ID</div>
						<div class="font-mono text-xs">{task.id}</div>
					</div>

					<div>
						<div class="text-gray-500 dark:text-gray-400">项目ID</div>
						<div class="font-mono text-xs">{task.projectId}</div>
					</div>

					<div>
						<div class="text-gray-500 dark:text-gray-400">超时时间</div>
						<div class="font-medium">{task.timeout}秒</div>
					</div>

					<div>
						<div class="text-gray-500 dark:text-gray-400">最大重试</div>
						<div class="font-medium">{task.maxRetries}次</div>
					</div>
				</div>
			</div>

			<!-- 错误信息 -->
			{#if task.error}
				<div class="border-t pt-4">
					<div class="text-gray-500 dark:text-gray-400 mb-2">错误信息</div>
					<div
						class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-3"
					>
						<div class="text-red-800 dark:text-red-200 text-sm font-mono">
							{task.error}
						</div>
					</div>
				</div>
			{/if}

			<!-- 结果预览 -->
			{#if task.result}
				<div class="border-t pt-4">
					<div class="text-gray-500 dark:text-gray-400 mb-2">执行结果</div>
					<div
						class="bg-gray-50 dark:bg-gray-900/20 border border-gray-200 dark:border-gray-700 rounded-md p-3"
					>
						<div class="text-gray-800 dark:text-gray-200 text-sm">
							{#if typeof task.result === 'string'}
								<pre class="whitespace-pre-wrap font-mono">{task.result}</pre>
							{:else}
								<pre class="whitespace-pre-wrap font-mono">{JSON.stringify(
										task.result,
										null,
										2
									)}</pre>
							{/if}
						</div>
					</div>
				</div>
			{/if}

			<!-- 标签 -->
			{#if task.tags && task.tags.length > 0}
				<div class="border-t pt-4">
					<div class="text-gray-500 dark:text-gray-400 mb-2">标签</div>
					<div class="flex flex-wrap gap-2">
						{#each task.tags as tag}
							<Badge variant="outline">{tag}</Badge>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	</CardContent>
</Card>

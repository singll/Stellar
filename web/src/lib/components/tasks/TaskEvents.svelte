<!--
任务事件组件
显示任务的执行事件和日志
-->
<script lang="ts">
	import type { Task } from '$lib/types/task';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	interface Props {
		task: Task;
	}

	let { task }: Props = $props();

	// 模拟任务事件数据 (在实际应用中，这应该从API获取)
	let events = $state([
		{
			id: '1',
			type: 'info',
			message: '任务已创建',
			timestamp: task.createdAt,
			details: null
		},
		{
			id: '2',
			type: 'info',
			message: '任务开始执行',
			timestamp: task.startedAt || task.createdAt,
			details: null
		}
	]);

	// 获取事件类型的图标
	function getEventIcon(type: string): string {
		switch (type) {
			case 'info':
				return 'fas fa-info-circle';
			case 'warning':
				return 'fas fa-exclamation-triangle';
			case 'error':
				return 'fas fa-times-circle';
			case 'success':
				return 'fas fa-check-circle';
			case 'debug':
				return 'fas fa-bug';
			default:
				return 'fas fa-circle';
		}
	}

	// 获取事件类型的颜色
	function getEventColor(type: string): string {
		switch (type) {
			case 'info':
				return 'text-blue-600 dark:text-blue-400';
			case 'warning':
				return 'text-yellow-600 dark:text-yellow-400';
			case 'error':
				return 'text-red-600 dark:text-red-400';
			case 'success':
				return 'text-green-600 dark:text-green-400';
			case 'debug':
				return 'text-gray-600 dark:text-gray-400';
			default:
				return 'text-gray-600 dark:text-gray-400';
		}
	}

	// 获取事件类型的Badge样式
	function getEventBadgeVariant(type: string): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (type) {
			case 'error':
				return 'destructive';
			case 'warning':
				return 'outline';
			case 'success':
				return 'default';
			default:
				return 'secondary';
		}
	}

	// 格式化时间戳
	function formatTimestamp(timestamp: string): string {
		return new Date(timestamp).toLocaleString();
	}

	// 清空事件日志
	function clearEvents() {
		events = [];
	}

	// 导出事件日志
	function exportEvents() {
		const content = events
			.map(
				(event) =>
					`[${formatTimestamp(event.timestamp)}] ${event.type.toUpperCase()}: ${event.message}${event.details ? '\n' + event.details : ''}`
			)
			.join('\n\n');

		const blob = new Blob([content], { type: 'text/plain' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `task_${task.id}_events.txt`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	// 自动滚动到底部
	let eventsContainer: HTMLElement | undefined;
	let autoScroll = $state(true);

	function scrollToBottom() {
		if (autoScroll && eventsContainer) {
			eventsContainer.scrollTop = eventsContainer.scrollHeight;
		}
	}

	// 监听事件变化，自动滚动
	$effect(() => {
		if (events.length > 0) {
			scrollToBottom();
		}
	});
</script>

<Card>
	<CardHeader>
		<div class="flex items-center justify-between">
			<CardTitle>执行事件</CardTitle>

			<div class="flex items-center gap-2">
				<div class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
					<input
						type="checkbox"
						bind:checked={autoScroll}
						id="auto-scroll"
						class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
					/>
					<label for="auto-scroll">自动滚动</label>
				</div>

				<Button variant="outline" size="sm" onclick={exportEvents} disabled={events.length === 0}>
					<i class="fas fa-download mr-2"></i>
					导出
				</Button>

				<Button variant="outline" size="sm" onclick={clearEvents} disabled={events.length === 0}>
					<i class="fas fa-trash mr-2"></i>
					清空
				</Button>
			</div>
		</div>
	</CardHeader>
	<CardContent>
		{#if events.length === 0}
			<div class="text-center py-8 text-gray-500 dark:text-gray-400">
				<i class="fas fa-list text-2xl mb-2"></i>
				<p>暂无事件记录</p>
			</div>
		{:else}
			<div bind:this={eventsContainer} class="space-y-4 max-h-96 overflow-y-auto pr-2">
				{#each events as event}
					<div class="flex items-start gap-3 p-3 bg-gray-50 dark:bg-gray-900/20 rounded-lg">
						<div class="flex-shrink-0">
							<i class="{getEventIcon(event.type)} {getEventColor(event.type)} text-sm"></i>
						</div>

						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2 mb-1">
								<Badge variant={getEventBadgeVariant(event.type)} class="text-xs">
									{event.type.toUpperCase()}
								</Badge>
								<span class="text-xs text-gray-500 dark:text-gray-400">
									{formatTimestamp(event.timestamp)}
								</span>
							</div>

							<div class="text-sm text-gray-900 dark:text-white mb-1">
								{event.message}
							</div>

							{#if event.details}
								<div
									class="text-xs text-gray-600 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 rounded p-2 font-mono"
								>
									<pre class="whitespace-pre-wrap">{event.details}</pre>
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</CardContent>
</Card>

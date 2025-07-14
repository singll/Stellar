<!--
任务日志组件
显示任务的执行日志
-->
<script lang="ts">
	import type { Task } from '$lib/types/task';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';

	interface Props {
		task: Task;
	}

	let { task }: Props = $props();

	// 模拟日志数据 (在实际应用中，这应该从API获取)
	let logs = $state([
		{
			id: '1',
			level: 'info',
			message: '任务初始化完成',
			timestamp: task.createdAt,
			source: 'system'
		},
		{
			id: '2',
			level: 'info',
			message: '开始执行任务',
			timestamp: task.startedAt || task.createdAt,
			source: 'executor'
		}
	]);

	// 日志级别过滤
	let logLevelFilter = $state('all');
	let searchQuery = $state('');

	// 过滤后的日志
	let filteredLogs = $derived(() => {
		return logs.filter((log) => {
			const matchesLevel = logLevelFilter === 'all' || log.level === logLevelFilter;
			const matchesSearch =
				searchQuery === '' || log.message.toLowerCase().includes(searchQuery.toLowerCase());
			return matchesLevel && matchesSearch;
		}) as {
			id: string;
			level: string;
			message: string;
			timestamp: string;
			source: string;
		}[];
	});

	// 获取日志级别的图标
	function getLogLevelIcon(level: string): string {
		switch (level) {
			case 'debug':
				return 'fas fa-bug';
			case 'info':
				return 'fas fa-info-circle';
			case 'warning':
				return 'fas fa-exclamation-triangle';
			case 'error':
				return 'fas fa-times-circle';
			case 'fatal':
				return 'fas fa-skull-crossbones';
			default:
				return 'fas fa-circle';
		}
	}

	// 获取日志级别的颜色
	function getLogLevelColor(level: string): string {
		switch (level) {
			case 'debug':
				return 'text-gray-600 dark:text-gray-400';
			case 'info':
				return 'text-blue-600 dark:text-blue-400';
			case 'warning':
				return 'text-yellow-600 dark:text-yellow-400';
			case 'error':
				return 'text-red-600 dark:text-red-400';
			case 'fatal':
				return 'text-red-800 dark:text-red-300';
			default:
				return 'text-gray-600 dark:text-gray-400';
		}
	}

	// 格式化时间戳
	function formatTimestamp(timestamp: string): string {
		return new Date(timestamp).toLocaleString();
	}

	// 清空日志
	function clearLogs() {
		logs = [];
	}

	// 导出日志
	function exportLogs() {
		const content = filteredLogs()
			.map(
				(log) =>
					`[${formatTimestamp(log.timestamp)}] ${log.level.toUpperCase()} [${log.source}]: ${log.message}`
			)
			.join('\n');

		const blob = new Blob([content], { type: 'text/plain' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `task_${task.id}_logs.txt`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	// 自动滚动到底部
	let logsContainer = $state<HTMLElement | undefined>();
	let autoScroll = $state(true);

	function scrollToBottom() {
		if (autoScroll && logsContainer) {
			logsContainer.scrollTop = logsContainer.scrollHeight;
		}
	}

	// 监听日志变化，自动滚动
	$effect(() => {
		if (filteredLogs().length > 0) {
			scrollToBottom();
		}
	});

	// 日志级别选项
	const logLevelOptions = [
		{ value: 'all', label: '全部' },
		{ value: 'debug', label: 'DEBUG' },
		{ value: 'info', label: 'INFO' },
		{ value: 'warning', label: 'WARNING' },
		{ value: 'error', label: 'ERROR' },
		{ value: 'fatal', label: 'FATAL' }
	];
</script>

<Card>
	<CardHeader>
		<div class="flex items-center justify-between">
			<CardTitle>执行日志</CardTitle>

			<div class="flex items-center gap-2">
				<div class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
					<input
						type="checkbox"
						bind:checked={autoScroll}
						id="auto-scroll-logs"
						class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
					/>
					<label for="auto-scroll-logs">自动滚动</label>
				</div>

				<Button
					variant="outline"
					size="sm"
					onclick={exportLogs}
					disabled={filteredLogs().length === 0}
				>
					<i class="fas fa-download mr-2"></i>
					导出
				</Button>

				<Button variant="outline" size="sm" onclick={clearLogs} disabled={logs.length === 0}>
					<i class="fas fa-trash mr-2"></i>
					清空
				</Button>
			</div>
		</div>
	</CardHeader>
	<CardContent>
		<!-- 过滤控件 -->
		<div class="flex items-center gap-4 mb-4 p-3 bg-gray-50 dark:bg-gray-900/20 rounded-lg">
			<div class="flex items-center gap-2">
				<label for="log-level-filter" class="text-sm text-gray-600 dark:text-gray-400">级别:</label>
				<select
					id="log-level-filter"
					bind:value={logLevelFilter}
					class="text-sm border border-gray-300 dark:border-gray-600 rounded px-2 py-1 bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
				>
					{#each logLevelOptions as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<div class="flex items-center gap-2 flex-1">
				<label for="log-search" class="text-sm text-gray-600 dark:text-gray-400">搜索:</label>
				<input
					id="log-search"
					type="text"
					bind:value={searchQuery}
					placeholder="搜索日志..."
					class="flex-1 text-sm border border-gray-300 dark:border-gray-600 rounded px-2 py-1 bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
				/>
			</div>

			<div class="text-sm text-gray-600 dark:text-gray-400">
				显示 {filteredLogs().length} / {logs.length} 条日志
			</div>
		</div>

		{#if filteredLogs().length === 0}
			<div class="text-center py-8 text-gray-500 dark:text-gray-400">
				<i class="fas fa-file-alt text-2xl mb-2"></i>
				<p>暂无日志记录</p>
			</div>
		{:else}
			<div
				bind:this={logsContainer}
				class="space-y-1 max-h-96 overflow-y-auto pr-2 font-mono text-sm"
			>
				{#each filteredLogs() as log}
					<div
						class="flex items-start gap-3 p-2 hover:bg-gray-50 dark:hover:bg-gray-800/50 rounded"
					>
						<div class="flex-shrink-0 w-20 text-xs text-gray-500 dark:text-gray-400">
							{formatTimestamp(log.timestamp).split(' ')[1]}
						</div>

						<div class="flex-shrink-0 w-16">
							<Badge
								variant={log.level === 'error' || log.level === 'fatal'
									? 'destructive'
									: 'secondary'}
								class="text-xs"
							>
								{log.level.toUpperCase()}
							</Badge>
						</div>

						<div class="flex-shrink-0 w-20 text-xs text-gray-600 dark:text-gray-400">
							[{log.source}]
						</div>

						<div class="flex-1 text-gray-900 dark:text-white">
							{log.message}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</CardContent>
</Card>

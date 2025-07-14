<!--
任务结果组件
显示任务的执行结果
-->
<script lang="ts">
	import type { Task } from '$lib/types/task';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';

	interface Props {
		task: Task;
		result?: any; // 添加 result 属性
	}

	let { task, result }: Props = $props();

	// 格式化结果数据
	function formatResult(result: any): string {
		if (!result) return '';

		if (typeof result === 'string') {
			return result;
		}

		return JSON.stringify(result, null, 2);
	}

	// 判断是否为JSON格式
	function isJsonResult(result: any): boolean {
		return result && typeof result === 'object';
	}

	// 下载结果
	function downloadResult(format: 'json' | 'txt' = 'json') {
		if (!task.result) return;

		let content = '';
		let filename = `task_${task.id}_result`;
		let mimeType = 'application/json';

		if (format === 'json') {
			content = JSON.stringify(task.result, null, 2);
			filename += '.json';
			mimeType = 'application/json';
		} else {
			content = formatResult(task.result);
			filename += '.txt';
			mimeType = 'text/plain';
		}

		const blob = new Blob([content], { type: mimeType });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = filename;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	// 复制结果到剪贴板
	async function copyResult() {
		if (!task.result) return;

		try {
			const content = formatResult(task.result);
			await navigator.clipboard.writeText(content);
			// 这里可以添加成功提示
		} catch (error) {
			console.error('复制失败:', error);
			// 这里可以添加错误提示
		}
	}

	// 获取结果统计信息
	let resultStats = $derived(() => {
		if (!task.result) return null;

		const result = task.result as any;

		if (typeof result === 'string') {
			return {
				type: 'text',
				lines: result.split('\n').length,
				size: new Blob([result]).size
			};
		}

		if (Array.isArray(result)) {
			return {
				type: 'array',
				items: result.length,
				size: new Blob([JSON.stringify(result)]).size
			};
		}

		if (typeof result === 'object') {
			return {
				type: 'object',
				keys: Object.keys(result).length,
				size: new Blob([JSON.stringify(result)]).size
			};
		}

		return {
			type: 'unknown',
			size: new Blob([String(result)]).size
		};
	});

	// 获取统计信息值，用于模板访问
	let stats = $derived(resultStats());

	// 格式化文件大小
	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 B';

		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));

		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}
</script>

<Card>
	<CardHeader>
		<div class="flex items-center justify-between">
			<CardTitle>执行结果</CardTitle>

			{#if task.result}
				<div class="flex items-center gap-2">
					{#if stats}
						<div class="flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400">
							{#if stats.type === 'text' && stats.lines !== undefined}
								<span>{stats.lines} 行</span>
							{:else if stats.type === 'array' && stats.items !== undefined}
								<span>{stats.items} 项</span>
							{:else if stats.type === 'object' && stats.keys !== undefined}
								<span>{stats.keys} 键</span>
							{/if}
							<span>{formatFileSize(stats.size)}</span>
						</div>
					{/if}

					<Button variant="outline" size="sm" onclick={copyResult}>
						<i class="fas fa-copy mr-2"></i>
						复制
					</Button>

					<Button variant="outline" size="sm" onclick={() => downloadResult('txt')}>
						<i class="fas fa-download mr-2"></i>
						下载
					</Button>

					{#if isJsonResult(task.result)}
						<Button variant="outline" size="sm" onclick={() => downloadResult('json')}>
							<i class="fas fa-file-code mr-2"></i>
							JSON
						</Button>
					{/if}
				</div>
			{/if}
		</div>
	</CardHeader>
	<CardContent>
		{#if task.status === 'pending'}
			<div class="text-center py-8 text-gray-500 dark:text-gray-400">
				<i class="fas fa-clock text-2xl mb-2"></i>
				<p>任务尚未开始执行</p>
			</div>
		{:else if task.status === 'running'}
			<div class="text-center py-8 text-gray-500 dark:text-gray-400">
				<i class="fas fa-spinner fa-spin text-2xl mb-2"></i>
				<p>任务正在执行中...</p>
				{#if task.progress !== undefined}
					<div class="mt-4 w-full max-w-md mx-auto">
						<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
							<div
								class="bg-blue-600 h-2 rounded-full transition-all duration-300"
								style="width: {task.progress}%"
							></div>
						</div>
						<div class="mt-2 text-sm">{task.progress}%</div>
					</div>
				{/if}
			</div>
		{:else if task.status === 'cancelled'}
			<div class="text-center py-8 text-gray-500 dark:text-gray-400">
				<i class="fas fa-ban text-2xl mb-2"></i>
				<p>任务已被取消</p>
			</div>
		{:else if task.status === 'failed'}
			<div class="text-center py-8">
				<i class="fas fa-exclamation-triangle text-2xl mb-2 text-red-500"></i>
				<p class="text-red-600 dark:text-red-400 mb-4">任务执行失败</p>

				{#if task.error}
					<div
						class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 text-left"
					>
						<div class="text-red-800 dark:text-red-200 text-sm">
							<div class="font-medium mb-2">错误信息:</div>
							<pre class="whitespace-pre-wrap font-mono">{task.error}</pre>
						</div>
					</div>
				{/if}
			</div>
		{:else if !task.result}
			<div class="text-center py-8 text-gray-500 dark:text-gray-400">
				<i class="fas fa-file-alt text-2xl mb-2"></i>
				<p>暂无执行结果</p>
			</div>
		{:else}
			<div class="space-y-4">
				<!-- 结果统计 -->
				{#if stats}
					<div
						class="flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400 pb-4 border-b border-gray-200 dark:border-gray-700"
					>
						<div class="flex items-center gap-2">
							<i class="fas fa-info-circle"></i>
							<span>类型: {stats.type}</span>
						</div>

						{#if stats.type === 'text' && stats.lines !== undefined}
							<span>行数: {stats.lines}</span>
						{:else if stats.type === 'array' && stats.items !== undefined}
							<span>数量: {stats.items}</span>
						{:else if stats.type === 'object' && stats.keys !== undefined}
							<span>属性: {stats.keys}</span>
						{/if}

						<span>大小: {formatFileSize(stats.size)}</span>
					</div>
				{/if}

				<!-- 结果内容 -->
				<div
					class="bg-gray-50 dark:bg-gray-900/20 border border-gray-200 dark:border-gray-700 rounded-md"
				>
					<div class="p-4">
						{#if isJsonResult(task.result)}
							<pre
								class="text-sm text-gray-800 dark:text-gray-200 whitespace-pre-wrap overflow-x-auto font-mono max-h-96 overflow-y-auto">{formatResult(
									task.result
								)}</pre>
						{:else}
							<div
								class="text-sm text-gray-800 dark:text-gray-200 whitespace-pre-wrap overflow-x-auto font-mono max-h-96 overflow-y-auto"
							>
								{formatResult(task.result)}
							</div>
						{/if}
					</div>
				</div>
			</div>
		{/if}
	</CardContent>
</Card>

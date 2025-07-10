<!--
任务头部组件
显示任务的基本信息和操作按钮
-->
<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Task } from '$lib/types/task';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';

	interface Props {
		task: Task;
		loading?: boolean;
	}

	let { task, loading = false }: Props = $props();

	const dispatch = createEventDispatcher<{
		start: void;
		stop: void;
		cancel: void;
		restart: void;
		delete: void;
		edit: void;
	}>();

	// 获取任务状态的样式类
	function getStatusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (status) {
			case 'running':
				return 'default';
			case 'completed':
				return 'secondary';
			case 'failed':
			case 'cancelled':
				return 'destructive';
			default:
				return 'outline';
		}
	}

	// 获取任务类型信息
	function getTaskTypeInfo(type: string) {
		const typeMap = {
			subdomain_enum: { label: '子域名枚举', icon: 'fas fa-globe' },
			port_scan: { label: '端口扫描', icon: 'fas fa-network-wired' },
			vuln_scan: { label: '漏洞扫描', icon: 'fas fa-bug' },
			asset_discovery: { label: '资产发现', icon: 'fas fa-search' },
			dir_scan: { label: '目录扫描', icon: 'fas fa-folder' },
			web_crawl: { label: 'Web爬虫', icon: 'fas fa-spider' }
		};
		return typeMap[type] || { label: type, icon: 'fas fa-cog' };
	}

	// 获取优先级信息
	function getPriorityInfo(priority: string) {
		const priorityMap = {
			low: { label: '低', color: 'text-gray-500' },
			normal: { label: '正常', color: 'text-blue-500' },
			high: { label: '高', color: 'text-yellow-500' },
			critical: { label: '紧急', color: 'text-red-500' }
		};
		return priorityMap[priority] || { label: priority, color: 'text-gray-500' };
	}

	let typeInfo = $derived(getTaskTypeInfo(task.type));
	let priorityInfo = $derived(getPriorityInfo(task.priority));

	// 判断是否可以执行操作
	let canStart = $derived(task.status === 'pending' || task.status === 'failed');
	let canStop = $derived(task.status === 'running');
	let canCancel = $derived(task.status === 'running' || task.status === 'pending');
	let canRestart = $derived(
		task.status === 'completed' || task.status === 'failed' || task.status === 'cancelled'
	);
</script>

<div class="bg-white dark:bg-gray-800 shadow rounded-lg p-6 mb-6">
	<div class="flex items-start justify-between mb-4">
		<div class="flex-1">
			<div class="flex items-center gap-3 mb-2">
				<div class="flex items-center gap-2">
					<i class="{typeInfo.icon} text-blue-600 dark:text-blue-400"></i>
					<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{task.name}</h1>
				</div>
				<Badge variant={getStatusVariant(task.status)}>{task.status}</Badge>
				<Badge variant="outline">{typeInfo.label}</Badge>
			</div>

			{#if task.description}
				<p class="text-gray-600 dark:text-gray-300 mb-3">{task.description}</p>
			{/if}

			<div class="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
				<div class="flex items-center gap-1">
					<i class="fas fa-flag {priorityInfo.color}"></i>
					<span>优先级: {priorityInfo.label}</span>
				</div>
				<div class="flex items-center gap-1">
					<i class="fas fa-clock"></i>
					<span>创建时间: {new Date(task.createdAt).toLocaleString()}</span>
				</div>
				{#if task.startedAt}
					<div class="flex items-center gap-1">
						<i class="fas fa-play"></i>
						<span>开始时间: {new Date(task.startedAt).toLocaleString()}</span>
					</div>
				{/if}
				{#if task.completedAt}
					<div class="flex items-center gap-1">
						<i class="fas fa-check"></i>
						<span>完成时间: {new Date(task.completedAt).toLocaleString()}</span>
					</div>
				{/if}
			</div>
		</div>

		<!-- 操作按钮 -->
		<div class="flex items-center gap-2">
			{#if canStart}
				<Button variant="default" size="sm" onclick={() => dispatch('start')} disabled={loading}>
					<i class="fas fa-play mr-2"></i>
					开始
				</Button>
			{/if}

			{#if canStop}
				<Button variant="outline" size="sm" onclick={() => dispatch('stop')} disabled={loading}>
					<i class="fas fa-stop mr-2"></i>
					停止
				</Button>
			{/if}

			{#if canCancel}
				<Button variant="outline" size="sm" onclick={() => dispatch('cancel')} disabled={loading}>
					<i class="fas fa-ban mr-2"></i>
					取消
				</Button>
			{/if}

			{#if canRestart}
				<Button variant="outline" size="sm" onclick={() => dispatch('restart')} disabled={loading}>
					<i class="fas fa-redo mr-2"></i>
					重启
				</Button>
			{/if}

			<Button variant="outline" size="sm" onclick={() => dispatch('edit')} disabled={loading}>
				<i class="fas fa-edit mr-2"></i>
				编辑
			</Button>

			<Button variant="destructive" size="sm" onclick={() => dispatch('delete')} disabled={loading}>
				<i class="fas fa-trash mr-2"></i>
				删除
			</Button>
		</div>
	</div>

	<!-- 进度条 -->
	{#if task.status === 'running' && task.progress !== undefined}
		<div class="mt-4">
			<div class="flex items-center justify-between mb-2">
				<span class="text-sm text-gray-600 dark:text-gray-400">执行进度</span>
				<span class="text-sm text-gray-600 dark:text-gray-400">{task.progress}%</span>
			</div>
			<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
				<div
					class="bg-blue-600 h-2 rounded-full transition-all duration-300"
					style="width: {task.progress}%"
				></div>
			</div>
		</div>
	{/if}
</div>

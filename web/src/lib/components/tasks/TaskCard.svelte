<!--
任务卡片组件
在任务列表中显示单个任务的信息
-->
<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Task } from '$lib/types/task';
	import { Badge } from '$lib/components/ui/badge';
	import ProgressBar from '$lib/components/ui/ProgressBar.svelte';
	import { formatRelativeTime, formatDateTime } from '$lib/utils/date';

	interface Props {
		task: Task;
		selected?: boolean;
		onclick?: () => void;
		onselect?: (selected: boolean) => void;
	}

	let { task, selected = false, onclick, onselect }: Props = $props();

	const dispatch = createEventDispatcher();

	// 处理卡片点击
	function handleClick(event: MouseEvent) {
		// 如果点击的是复选框或按钮，不触发卡片点击
		const target = event.target as HTMLElement;
		if ((target as HTMLInputElement).type === 'checkbox' || target.closest('button')) {
			return;
		}

		onclick?.();
		dispatch('click');
	}

	// 处理键盘事件
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			onclick?.();
			dispatch('click');
		}
	}

	// 处理选择状态变化
	function handleSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		onselect?.(target.checked);
		dispatch('select', { selected: target.checked });
	}

	// 阻止事件冒泡
	function stopPropagation(event: Event) {
		event.stopPropagation();
	}

	// 获取任务类型信息
	function getTaskTypeInfo(type: string) {
		const typeMap: Record<string, { label: string; icon: string; color: string }> = {
			subdomain_enum: { label: '子域名枚举', icon: 'fas fa-globe', color: 'blue' },
			port_scan: { label: '端口扫描', icon: 'fas fa-network-wired', color: 'green' },
			vuln_scan: { label: '漏洞扫描', icon: 'fas fa-bug', color: 'red' },
			asset_discovery: { label: '资产发现', icon: 'fas fa-search', color: 'purple' },
			dir_scan: { label: '目录扫描', icon: 'fas fa-folder', color: 'yellow' },
			web_crawl: { label: 'Web爬虫', icon: 'fas fa-spider', color: 'indigo' },
			sensitive_scan: { label: '敏感信息扫描', icon: 'fas fa-eye', color: 'orange' },
			page_monitor: { label: '页面监控', icon: 'fas fa-monitor', color: 'teal' }
		};
		return typeMap[type] || { label: type, icon: 'fas fa-cog', color: 'gray' };
	}

	// 获取状态信息
	function getStatusInfo(status: string) {
		const statusMap: Record<
			string,
			{ label: string; color: string; icon: string; pulse?: boolean }
		> = {
			pending: { label: '等待中', color: 'gray', icon: 'fas fa-clock' },
			queued: { label: '队列中', color: 'blue', icon: 'fas fa-hourglass-half' },
			running: { label: '运行中', color: 'yellow', icon: 'fas fa-play', pulse: true },
			completed: { label: '已完成', color: 'green', icon: 'fas fa-check-circle' },
			failed: { label: '失败', color: 'red', icon: 'fas fa-exclamation-circle' },
			canceled: { label: '已取消', color: 'gray', icon: 'fas fa-ban' },
			timeout: { label: '超时', color: 'orange', icon: 'fas fa-clock' }
		};
		return statusMap[status] || { label: status, color: 'gray', icon: 'fas fa-question' };
	}

	// 获取优先级信息
	function getPriorityInfo(priority: string) {
		const priorityMap: Record<string, { label: string; color: string }> = {
			low: { label: '低', color: 'gray' },
			normal: { label: '正常', color: 'blue' },
			high: { label: '高', color: 'yellow' },
			critical: { label: '紧急', color: 'red' }
		};
		return priorityMap[priority] || { label: priority, color: 'gray' };
	}

	let typeInfo = $derived(getTaskTypeInfo(task.type));
	let statusInfo = $derived(getStatusInfo(task.status));
	let priorityInfo = $derived(getPriorityInfo(task.priority));

	// 将color映射到Badge variant
	function colorToVariant(color: string): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (color) {
			case 'red':
				return 'destructive';
			case 'gray':
			case 'grey':
				return 'secondary';
			case 'green':
			case 'blue':
			case 'yellow':
			case 'purple':
			case 'indigo':
			case 'orange':
			case 'teal':
			default:
				return 'default';
		}
	}
</script>

<div
	class="p-4 hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer transition-colors {selected
		? 'bg-blue-50 dark:bg-blue-900/20 border-l-4 border-l-blue-500'
		: ''}"
	onclick={handleClick}
	role="button"
	tabindex="0"
	onkeydown={handleKeydown}
>
	<div class="flex items-start gap-4">
		<!-- 选择框 -->
		<div
			class="flex-shrink-0 pt-1"
			onclick={stopPropagation}
			onkeydown={(e) => e.key === 'Enter' && stopPropagation(e)}
			role="button"
			tabindex="0"
		>
			<input
				type="checkbox"
				checked={selected}
				onchange={handleSelect}
				class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
			/>
		</div>

		<!-- 任务类型图标 -->
		<div class="flex-shrink-0 pt-1">
			<div
				class="w-10 h-10 rounded-lg bg-{typeInfo.color}-100 dark:bg-{typeInfo.color}-900/20 flex items-center justify-center"
			>
				<i class="{typeInfo.icon} text-{typeInfo.color}-600 dark:text-{typeInfo.color}-400"></i>
			</div>
		</div>

		<!-- 任务信息 -->
		<div class="flex-1 min-w-0">
			<div class="flex items-start justify-between">
				<div class="flex-1 min-w-0">
					<!-- 任务名称和状态 -->
					<div class="flex items-center gap-2 mb-1">
						<h3 class="font-medium text-gray-900 dark:text-white truncate">
							{task.name}
						</h3>
						<Badge variant={colorToVariant(statusInfo.color)}>
							<i class="{statusInfo.icon} mr-1"></i>
							{statusInfo.label}
						</Badge>
					</div>

					<!-- 任务描述 -->
					{#if task.description}
						<p class="text-gray-600 dark:text-gray-400 text-sm mb-2 line-clamp-2">
							{task.description}
						</p>
					{/if}

					<!-- 任务详情 -->
					<div class="flex flex-wrap items-center gap-4 text-sm text-gray-600 dark:text-gray-400">
						<div class="flex items-center gap-1">
							<i class={typeInfo.icon}></i>
							<span>{typeInfo.label}</span>
						</div>

						<div class="flex items-center gap-1">
							<i class="fas fa-flag"></i>
							<Badge variant={colorToVariant(priorityInfo.color)}>
								{priorityInfo.label}
							</Badge>
						</div>

						{#if task.tags?.length > 0}
							<div class="flex items-center gap-1">
								<i class="fas fa-tags"></i>
								<div class="flex gap-1">
									{#each task.tags.slice(0, 2) as tag}
										<Badge variant="secondary">{tag}</Badge>
									{/each}
									{#if task.tags.length > 2}
										<Badge variant="secondary">+{task.tags.length - 2}</Badge>
									{/if}
								</div>
							</div>
						{/if}

						<div class="flex items-center gap-1">
							<i class="fas fa-clock"></i>
							<span title={formatDateTime(task.createdAt)}>
								{formatRelativeTime(task.createdAt)}
							</span>
						</div>
					</div>
				</div>

				<!-- 右侧信息 -->
				<div class="flex-shrink-0 text-right">
					<!-- 进度条 -->
					{#if task.status === 'running' && task.progress > 0}
						<div class="w-24 mb-2">
							<ProgressBar value={task.progress} size="sm" />
							<div class="text-xs text-gray-600 dark:text-gray-400 mt-1 text-center">
								{Math.round(task.progress)}%
							</div>
						</div>
					{/if}

					<!-- 时间信息 -->
					<div class="text-xs text-gray-600 dark:text-gray-400">
						{#if task.status === 'running' && task.startedAt}
							<div>开始: {formatRelativeTime(task.startedAt)}</div>
						{:else if task.status === 'completed' && task.completedAt}
							<div>完成: {formatRelativeTime(task.completedAt)}</div>
						{:else if task.status === 'failed' && task.completedAt}
							<div>失败: {formatRelativeTime(task.completedAt)}</div>
						{/if}

						{#if task.retryCount > 0}
							<div class="mt-1">
								<i class="fas fa-redo mr-1"></i>
								重试 {task.retryCount}
							</div>
						{/if}
					</div>
				</div>
			</div>
		</div>
	</div>
</div>

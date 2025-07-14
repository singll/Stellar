<!--
任务统计卡片组件
显示任务的统计信息
-->
<script lang="ts">
	import type { TaskStats } from '$lib/types/task';
	import StatCard from '$lib/components/ui/StatCard.svelte';

	interface Props {
		stats: TaskStats;
	}

	let { stats }: Props = $props();

	// 计算完成率
	let completionRate = $derived(
		stats.total > 0 ? Math.round((stats.completed / stats.total) * 100) : 0
	);

	// 计算失败率
	let failureRate = $derived(stats.total > 0 ? Math.round((stats.failed / stats.total) * 100) : 0);

	// 计算活跃任务数（运行中 + 队列中）
	let activeTasks = $derived(stats.running + stats.queued);

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
</script>

<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-7 gap-4 mb-6">
	<!-- 总任务数 -->
	<StatCard title="总任务" value={stats.total} icon="fas fa-tasks" color="blue" />

	<!-- 运行中 -->
	<StatCard
		title="运行中"
		value={stats.running}
		icon="fas fa-play"
		color="yellow"
		pulse={stats.running > 0}
	/>

	<!-- 队列中 -->
	<StatCard title="队列中" value={stats.queued} icon="fas fa-hourglass-half" color="blue" />

	<!-- 已完成 -->
	<StatCard
		title="已完成"
		value={stats.completed}
		icon="fas fa-check-circle"
		color="green"
		subtitle="{completionRate}%"
	/>

	<!-- 失败 -->
	<StatCard
		title="失败"
		value={stats.failed}
		icon="fas fa-exclamation-circle"
		color="red"
		subtitle="{failureRate}%"
	/>

	<!-- 已取消 -->
	<StatCard title="已取消" value={stats.cancelled} icon="fas fa-ban" color="gray" />

	<!-- 超时 -->
	<StatCard title="超时" value={stats.timeout} icon="fas fa-clock" color="orange" />
</div>

<!-- 按类型统计（如果有多个类型） -->
{#if Object.keys(stats.byType).length > 1}
	<div
		class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 mb-6"
	>
		<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">按类型统计</h3>
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
			{#each Object.entries(stats.byType) as [type, count]}
				{#if count > 0}
					{@const typeInfo = getTaskTypeInfo(type)}
					<div class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
						<div
							class="w-8 h-8 rounded-lg bg-{typeInfo.color}-100 dark:bg-{typeInfo.color}-900/20 flex items-center justify-center"
						>
							<i
								class="{typeInfo.icon} text-{typeInfo.color}-600 dark:text-{typeInfo.color}-400 text-sm"
							></i>
						</div>
						<div>
							<div class="font-medium text-gray-900 dark:text-white">
								{count}
							</div>
							<div class="text-sm text-gray-600 dark:text-gray-400">
								{typeInfo.label}
							</div>
						</div>
					</div>
				{/if}
			{/each}
		</div>
	</div>
{/if}

<!-- 按优先级统计（如果有多个优先级） -->
{#if Object.keys(stats.byPriority).length > 1}
	<div
		class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 mb-6"
	>
		<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">按优先级统计</h3>
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
			{#each Object.entries(stats.byPriority) as [priority, count]}
				{#if count > 0}
					{@const priorityInfo = getPriorityInfo(priority)}
					<div class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
						<div
							class="w-8 h-8 rounded-lg bg-{priorityInfo.color}-100 dark:bg-{priorityInfo.color}-900/20 flex items-center justify-center"
						>
							<i
								class="fas fa-flag text-{priorityInfo.color}-600 dark:text-{priorityInfo.color}-400 text-sm"
							></i>
						</div>
						<div>
							<div class="font-medium text-gray-900 dark:text-white">
								{count}
							</div>
							<div class="text-sm text-gray-600 dark:text-gray-400">
								{priorityInfo.label}
							</div>
						</div>
					</div>
				{/if}
			{/each}
		</div>
	</div>
{/if}

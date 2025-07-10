<!--
任务过滤器组件
提供任务状态、类型、优先级、项目等过滤选项
-->
<script lang="ts">
	import type { TaskStatus, TaskType, TaskPriority } from '$lib/types/task';
	import type { Project } from '$lib/types/project';
	import { Button } from '$lib/components/ui/button';
	import { Select } from '$lib/components/ui/select';
	import { createEventDispatcher } from 'svelte';

	interface Props {
		status?: TaskStatus | '';
		type?: TaskType | '';
		priority?: TaskPriority | '';
		projectId?: string;
		projects?: Project[];
	}

	let {
		status = $bindable(''),
		type = $bindable(''),
		priority = $bindable(''),
		projectId = $bindable(''),
		projects = []
	}: Props = $props();

	const dispatch = createEventDispatcher<{
		filter: {
			status: TaskStatus | '';
			type: TaskType | '';
			priority: TaskPriority | '';
			projectId: string;
		};
	}>();

	const statusOptions = [
		{ value: '', label: '全部状态' },
		{ value: 'pending', label: '等待中' },
		{ value: 'running', label: '运行中' },
		{ value: 'completed', label: '已完成' },
		{ value: 'failed', label: '失败' },
		{ value: 'cancelled', label: '已取消' }
	];

	const typeOptions = [
		{ value: '', label: '全部类型' },
		{ value: 'subdomain_enum', label: '子域名枚举' },
		{ value: 'port_scan', label: '端口扫描' },
		{ value: 'vuln_scan', label: '漏洞扫描' },
		{ value: 'asset_discovery', label: '资产发现' },
		{ value: 'dir_scan', label: '目录扫描' },
		{ value: 'web_crawl', label: 'Web爬虫' }
	];

	const priorityOptions = [
		{ value: '', label: '全部优先级' },
		{ value: 'low', label: '低' },
		{ value: 'normal', label: '正常' },
		{ value: 'high', label: '高' },
		{ value: 'critical', label: '紧急' }
	];

	let projectOptions = $derived([
		{ value: '', label: '全部项目' },
		...(projects?.map((p) => ({ value: p.id, label: p.name })) || [])
	]);

	function handleFilterChange() {
		dispatch('filter', {
			status: status as TaskStatus | '',
			type: type as TaskType | '',
			priority: priority as TaskPriority | '',
			projectId
		});
	}

	function resetFilters() {
		status = '';
		type = '';
		priority = '';
		projectId = '';
		handleFilterChange();
	}

	let hasActiveFilters = $derived(
		status !== '' || type !== '' || priority !== '' || projectId !== ''
	);
</script>

<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
	<!-- 状态过滤 -->
	<div>
		<div class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">状态</div>
		<Select bind:value={status} options={statusOptions} onchange={handleFilterChange} />
	</div>

	<!-- 类型过滤 -->
	<div>
		<div class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">类型</div>
		<Select bind:value={type} options={typeOptions} onchange={handleFilterChange} />
	</div>

	<!-- 优先级过滤 -->
	<div>
		<div class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">优先级</div>
		<Select bind:value={priority} options={priorityOptions} onchange={handleFilterChange} />
	</div>

	<!-- 项目过滤 -->
	<div>
		<div class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">项目</div>
		<Select bind:value={projectId} options={projectOptions} onchange={handleFilterChange} />
	</div>

	<!-- 重置按钮 -->
	<div class="flex items-end">
		<Button
			variant="outline"
			size="sm"
			onclick={resetFilters}
			disabled={!hasActiveFilters}
			class="w-full"
		>
			<i class="fas fa-undo mr-2"></i>
			重置
		</Button>
	</div>
</div>

<!-- 活动过滤器提示 -->
{#if hasActiveFilters}
	<div class="mt-4 flex flex-wrap gap-2">
		<span class="text-sm text-gray-600 dark:text-gray-400">活动过滤器:</span>

		{#if status}
			<span
				class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200"
			>
				状态: {statusOptions.find((opt) => opt.value === status)?.label}
				<button
					onclick={() => {
						status = '';
						handleFilterChange();
					}}
					class="ml-1 hover:text-blue-600 dark:hover:text-blue-400"
					aria-label="清除状态过滤器"
				>
					<i class="fas fa-times"></i>
				</button>
			</span>
		{/if}

		{#if type}
			<span
				class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200"
			>
				类型: {typeOptions.find((opt) => opt.value === type)?.label}
				<button
					onclick={() => {
						type = '';
						handleFilterChange();
					}}
					class="ml-1 hover:text-green-600 dark:hover:text-green-400"
					aria-label="清除类型过滤器"
				>
					<i class="fas fa-times"></i>
				</button>
			</span>
		{/if}

		{#if priority}
			<span
				class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200"
			>
				优先级: {priorityOptions.find((opt) => opt.value === priority)?.label}
				<button
					onclick={() => {
						priority = '';
						handleFilterChange();
					}}
					class="ml-1 hover:text-yellow-600 dark:hover:text-yellow-400"
					aria-label="清除优先级过滤器"
				>
					<i class="fas fa-times"></i>
				</button>
			</span>
		{/if}

		{#if projectId}
			<span
				class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200"
			>
				项目: {projectOptions.find((opt) => opt.value === projectId)?.label}
				<button
					onclick={() => {
						projectId = '';
						handleFilterChange();
					}}
					class="ml-1 hover:text-purple-600 dark:hover:text-purple-400"
					aria-label="清除项目过滤器"
				>
					<i class="fas fa-times"></i>
				</button>
			</span>
		{/if}
	</div>
{/if}

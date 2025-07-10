<!-- 端口扫描任务列表页面 -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { portScanStore, type PortScanTask } from '$lib/stores/portscan';
	import { toastStore } from '$lib/stores/toast';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import SearchInput from '$lib/components/ui/SearchInput.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import StatCard from '$lib/components/ui/StatCard.svelte';
	import { formatDateTime, formatDuration } from '$lib/utils/date';

	// 响应式状态
	let searchTerm = $state('');
	let selectedStatus = $state('all');
	let selectedProject = $state('all');
	let showFilters = $state(false);
	let isRefreshing = $state(false);

	// Store 订阅
	const tasks = $derived(portScanStore.tasks);
	const loading = $derived(portScanStore.loading);
	const pagination = $derived(portScanStore.pagination);
	const taskSummary = $derived(portScanStore.taskSummary);
	const runningTasks = $derived(portScanStore.runningTasks);

	// 状态选项
	const statusOptions = [
		{ value: 'all', label: '全部状态' },
		{ value: 'pending', label: '待执行' },
		{ value: 'queued', label: '已排队' },
		{ value: 'running', label: '运行中' },
		{ value: 'completed', label: '已完成' },
		{ value: 'failed', label: '失败' },
		{ value: 'canceled', label: '已取消' }
	];

	// 项目选项（模拟数据）
	const projectOptions = [
		{ value: 'all', label: '全部项目' },
		{ value: 'proj1', label: '项目 1' },
		{ value: 'proj2', label: '项目 2' }
	];

	// 初始化
	onMount(() => {
		loadTasks();

		// 设置定时刷新（每30秒）
		const interval = setInterval(() => {
			if ($runningTasks.length > 0) {
				refreshTasks();
			}
		}, 30000);

		return () => clearInterval(interval);
	});

	// 加载任务列表
	async function loadTasks() {
		try {
			const filters = buildFilters();
			await portScanStore.actions.loadTasks({
				page: $pagination.page,
				limit: $pagination.limit,
				filters
			});
		} catch (error) {
			console.error('加载任务失败:', error);
		}
	}

	// 刷新任务列表
	async function refreshTasks() {
		isRefreshing = true;
		try {
			await portScanStore.actions.refresh();
		} catch (error) {
			console.error('刷新任务失败:', error);
		} finally {
			isRefreshing = false;
		}
	}

	// 构建过滤器
	function buildFilters() {
		const filters: any = {};

		if (selectedStatus !== 'all') {
			filters.status = selectedStatus;
		}

		if (selectedProject !== 'all') {
			filters.projectId = selectedProject;
		}

		if (searchTerm.trim()) {
			filters.target = searchTerm.trim();
		}

		return filters;
	}

	// 应用过滤器
	async function applyFilters() {
		await loadTasks();
	}

	// 清除过滤器
	async function clearFilters() {
		searchTerm = '';
		selectedStatus = 'all';
		selectedProject = 'all';
		await loadTasks();
	}

	// 分页处理
	async function handlePageChange(event: CustomEvent<{ page: number }>) {
		const filters = buildFilters();
		await portScanStore.actions.loadTasks({
			page: event.detail.page,
			limit: $pagination.limit,
			filters
		});
	}

	// 任务操作
	async function handleTaskAction(action: string, task: PortScanTask) {
		try {
			switch (action) {
				case 'view':
					await goto(`/portscan/${task.id}`);
					break;
				case 'cancel':
					await portScanStore.actions.cancelTask(task.id);
					toastStore.success('任务已取消');
					break;
				case 'retry':
					await portScanStore.actions.retryTask(task.id);
					toastStore.success('任务已重新启动');
					break;
				case 'delete':
					if (confirm('确定要删除这个任务吗？')) {
						await portScanStore.actions.deleteTask(task.id);
						toastStore.success('任务已删除');
					}
					break;
				case 'export':
					await portScanStore.actions.exportResults(task.id, 'csv');
					toastStore.success('结果导出成功');
					break;
			}
		} catch (error) {
			console.error('任务操作失败:', error);
			toastStore.error('操作失败: ' + (error as Error).message);
		}
	}

	// 批量操作
	let selectedTasks = $state<string[]>([]);

	function handleTaskSelect(taskId: string, checked: boolean) {
		if (checked) {
			selectedTasks = [...selectedTasks, taskId];
		} else {
			selectedTasks = selectedTasks.filter((id) => id !== taskId);
		}
	}

	function handleSelectAll(checked: boolean) {
		if (checked) {
			selectedTasks = $tasks.map((task) => task.id);
		} else {
			selectedTasks = [];
		}
	}

	async function handleBulkDelete() {
		if (selectedTasks.length === 0) return;

		if (confirm(`确定要删除选中的 ${selectedTasks.length} 个任务吗？`)) {
			try {
				for (const taskId of selectedTasks) {
					await portScanStore.actions.deleteTask(taskId);
				}
				selectedTasks = [];
				toastStore.success('批量删除成功');
				await loadTasks();
			} catch (error) {
				console.error('批量删除失败:', error);
				toastStore.error('批量删除失败');
			}
		}
	}

	// 获取任务状态样式
	function getStatusVariant(status: string) {
		switch (status) {
			case 'completed':
				return 'success';
			case 'running':
				return 'warning';
			case 'failed':
				return 'destructive';
			case 'canceled':
				return 'secondary';
			default:
				return 'default';
		}
	}

	// 获取任务状态文本
	function getStatusText(status: string) {
		switch (status) {
			case 'pending':
				return '待执行';
			case 'queued':
				return '已排队';
			case 'running':
				return '运行中';
			case 'completed':
				return '已完成';
			case 'failed':
				return '失败';
			case 'canceled':
				return '已取消';
			default:
				return status;
		}
	}

	// 获取进度百分比
	function getProgressPercentage(task: PortScanTask) {
		if (!task.progress) return 0;
		return Math.round((task.progress.completed / task.progress.total) * 100);
	}
</script>

<svelte:head>
	<title>端口扫描任务 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6">
	<!-- 页面标题 -->
	<div class="flex justify-between items-center mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">端口扫描任务</h1>
			<p class="text-gray-600 mt-1">管理和监控端口扫描任务</p>
		</div>
		<div class="flex gap-2">
			<Button variant="outline" onclick={refreshTasks} disabled={isRefreshing}>
				{#if isRefreshing}
					<LoadingSpinner size="sm" />
				{:else}
					刷新
				{/if}
			</Button>
			<Button onclick={() => goto('/portscan/create')}>新建任务</Button>
		</div>
	</div>

	<!-- 统计卡片 -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
		<StatCard title="总任务数" value={$taskSummary.total} variant="default" />
		<StatCard title="运行中" value={$taskSummary.running} variant="warning" />
		<StatCard title="已完成" value={$taskSummary.completed} variant="success" />
		<StatCard title="成功率" value={`${$taskSummary.successRate}%`} variant="info" />
	</div>

	<!-- 过滤器 -->
	<div class="bg-white rounded-lg shadow-sm border p-4 mb-6">
		<div class="flex flex-wrap gap-4 items-center">
			<SearchInput bind:value={searchTerm} placeholder="搜索目标地址..." onchange={applyFilters} />

			<Select bind:value={selectedStatus} options={statusOptions} onchange={applyFilters} />

			<Select bind:value={selectedProject} options={projectOptions} onchange={applyFilters} />

			<Button variant="outline" onclick={clearFilters}>清除过滤</Button>

			<Button variant="outline" onclick={() => (showFilters = !showFilters)}>
				{showFilters ? '隐藏' : '显示'}高级过滤
			</Button>
		</div>

		{#if showFilters}
			<div class="mt-4 pt-4 border-t">
				<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">创建时间</label>
						<input
							type="date"
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">结束时间</label>
						<input
							type="date"
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">创建者</label>
						<input
							type="text"
							placeholder="输入创建者..."
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- 批量操作 -->
	{#if selectedTasks.length > 0}
		<div class="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
			<div class="flex items-center justify-between">
				<span class="text-blue-800">
					已选择 {selectedTasks.length} 个任务
				</span>
				<div class="flex gap-2">
					<Button variant="outline" size="sm" onclick={handleBulkDelete}>批量删除</Button>
					<Button variant="outline" size="sm" onclick={() => (selectedTasks = [])}>取消选择</Button>
				</div>
			</div>
		</div>
	{/if}

	<!-- 任务列表 -->
	<div class="bg-white rounded-lg shadow-sm border">
		{#if $loading.tasks}
			<div class="flex justify-center items-center py-12">
				<LoadingSpinner size="lg" />
			</div>
		{:else if $tasks.length === 0}
			<div class="text-center py-12">
				<div class="text-gray-400 text-lg mb-2">暂无任务</div>
				<p class="text-gray-500 mb-4">创建您的第一个端口扫描任务</p>
				<Button onclick={() => goto('/portscan/create')}>创建任务</Button>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-4 py-3 text-left">
								<input
									type="checkbox"
									onchange={(e) => handleSelectAll(e.target.checked)}
									class="rounded border-gray-300"
								/>
							</th>
							<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">任务名称</th>
							<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">目标</th>
							<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">状态</th>
							<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">进度</th>
							<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">开放端口</th>
							<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">创建时间</th>
							<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">操作</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200">
						{#each $tasks as task}
							<tr class="hover:bg-gray-50">
								<td class="px-4 py-3">
									<input
										type="checkbox"
										checked={selectedTasks.includes(task.id)}
										onchange={(e) => handleTaskSelect(task.id, e.target.checked)}
										class="rounded border-gray-300"
									/>
								</td>
								<td class="px-4 py-3">
									<div class="font-medium text-gray-900">{task.name}</div>
									{#if task.description}
										<div class="text-sm text-gray-500">{task.description}</div>
									{/if}
								</td>
								<td class="px-4 py-3">
									<div class="text-sm text-gray-900">{task.config.target}</div>
									<div class="text-xs text-gray-500">{task.config.ports}</div>
								</td>
								<td class="px-4 py-3">
									<Badge variant={getStatusVariant(task.status)}>
										{getStatusText(task.status)}
									</Badge>
								</td>
								<td class="px-4 py-3">
									{#if task.status === 'running' && task.progress}
										<div class="flex items-center gap-2">
											<div class="w-16 bg-gray-200 rounded-full h-2">
												<div
													class="bg-blue-500 h-2 rounded-full transition-all"
													style="width: {getProgressPercentage(task)}%"
												></div>
											</div>
											<span class="text-xs text-gray-600">{getProgressPercentage(task)}%</span>
										</div>
									{:else if task.status === 'completed'}
										<span class="text-sm text-green-600">完成</span>
									{:else}
										<span class="text-sm text-gray-400">-</span>
									{/if}
								</td>
								<td class="px-4 py-3">
									{#if task.result}
										<span class="text-sm text-gray-900">{task.result.openCount}</span>
									{:else}
										<span class="text-sm text-gray-400">-</span>
									{/if}
								</td>
								<td class="px-4 py-3">
									<div class="text-sm text-gray-900">{formatDateTime(task.createdAt)}</div>
									{#if task.endTime}
										<div class="text-xs text-gray-500">
											耗时: {formatDuration(task.startTime, task.endTime)}
										</div>
									{/if}
								</td>
								<td class="px-4 py-3">
									<div class="flex gap-1">
										<Button
											size="sm"
											variant="outline"
											onclick={() => handleTaskAction('view', task)}
										>
											查看
										</Button>

										{#if task.status === 'running'}
											<Button
												size="sm"
												variant="outline"
												onclick={() => handleTaskAction('cancel', task)}
											>
												取消
											</Button>
										{/if}

										{#if task.status === 'failed'}
											<Button
												size="sm"
												variant="outline"
												onclick={() => handleTaskAction('retry', task)}
											>
												重试
											</Button>
										{/if}

										{#if task.status === 'completed'}
											<Button
												size="sm"
												variant="outline"
												onclick={() => handleTaskAction('export', task)}
											>
												导出
											</Button>
										{/if}

										{#if task.status !== 'running'}
											<Button
												size="sm"
												variant="outline"
												onclick={() => handleTaskAction('delete', task)}
											>
												删除
											</Button>
										{/if}
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>

	<!-- 分页 -->
	{#if $tasks.length > 0}
		<div class="mt-6">
			<Pagination
				currentPage={$pagination.page}
				totalPages={$pagination.totalPages}
				totalItems={$pagination.total}
				itemsPerPage={$pagination.limit}
				onPageChange={handlePageChange}
			/>
		</div>
	{/if}
</div>

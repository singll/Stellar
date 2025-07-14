<!--
任务管理 - 任务列表页面
提供任务的列表、搜索、过滤、批量操作等功能
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { taskStore, taskActions } from '$lib/stores/tasks';
	import { projectStore, projectActions } from '$lib/stores/projects';
	import type { TaskStatus, TaskType, TaskPriority } from '$lib/types/task';
	import type { Project } from '$lib/types/project';

	import TaskCard from '$lib/components/tasks/TaskCard.svelte';
	import TaskFilters from '$lib/components/tasks/TaskFilters.svelte';
	import TaskStatsCards from '$lib/components/tasks/TaskStatsCards.svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import SearchInput from '$lib/components/ui/SearchInput.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';

	// 响应式状态
	let selectedTasks = $state<string[]>([]);
	let showFilters = $state(false);
	let searchQuery = $state('');

	// Store 状态使用 runes
	let store = $state({
		tasks: taskActions.tasks,
		loading: taskActions.loading,
		error: taskActions.error,
		pagination: taskActions.pagination,
		taskStats: null as any
	});

	let projects = $state<{ projects: Project[]; loading: boolean }>({
		projects: [],
		loading: false
	});

	// 分页和过滤参数
	let currentPage = $state(1);
	let pageSize = $state(20);
	let statusFilter = $state<TaskStatus | ''>('');
	let typeFilter = $state<TaskType | ''>('');
	let priorityFilter = $state<TaskPriority | ''>('');
	let projectFilter = $state('');

	// 更新 store 状态的副作用
	$effect(() => {
		store = {
			tasks: taskActions.tasks,
			loading: taskActions.loading,
			error: taskActions.error,
			pagination: taskActions.pagination,
			taskStats: null
		};
	});

	// 更新项目状态的副作用
	$effect(() => {
		// 从 projectStore 获取数据
		const unsubscribe = projectStore.subscribe((value) => {
			projects = value;
		});
		return unsubscribe;
	});

	// 组件挂载时初始化
	onMount(async () => {
		// 从URL查询参数初始化过滤器
		const params = new URLSearchParams($page.url.search);
		currentPage = parseInt(params.get('page') || '1');
		statusFilter = (params.get('status') as TaskStatus) || '';
		typeFilter = (params.get('type') as TaskType) || '';
		priorityFilter = (params.get('priority') as TaskPriority) || '';
		projectFilter = params.get('project') || '';
		searchQuery = params.get('search') || '';

		// 加载项目列表
		await projectActions.loadProjects();

		// 加载任务统计
		await taskActions.loadTaskStats();

		// 加载任务列表
		await loadTasks();
	});

	// 加载任务列表
	async function loadTasks() {
		const params = {
			page: currentPage,
			pageSize,
			status: statusFilter || undefined,
			type: typeFilter || undefined,
			priority: priorityFilter || undefined,
			projectId: projectFilter || undefined,
			search: searchQuery || undefined
		};

		await taskActions.loadTasks(params);

		// 更新URL
		updateURL();
	}

	// 更新URL查询参数
	function updateURL() {
		const params = new URLSearchParams();
		if (currentPage > 1) params.set('page', currentPage.toString());
		if (statusFilter) params.set('status', statusFilter);
		if (typeFilter) params.set('type', typeFilter);
		if (priorityFilter) params.set('priority', priorityFilter);
		if (projectFilter) params.set('project', projectFilter);
		if (searchQuery) params.set('search', searchQuery);

		const url = params.toString() ? `?${params.toString()}` : '';
		goto(`/tasks${url}`, { replaceState: true });
	}

	// 处理搜索
	function handleSearch() {
		currentPage = 1;
		loadTasks();
	}

	// 处理过滤器变化
	function handleFilterChange() {
		currentPage = 1;
		loadTasks();
	}

	// 处理分页变化
	function handlePageChange(event: CustomEvent<number>) {
		currentPage = event.detail;
		loadTasks();
	}

	// 处理任务选择
	function handleTaskSelect(taskId: string, selected: boolean) {
		if (selected) {
			selectedTasks = [...selectedTasks, taskId];
		} else {
			selectedTasks = selectedTasks.filter((id) => id !== taskId);
		}
	}

	// 全选/取消全选
	function handleSelectAll(selected: boolean) {
		if (selected) {
			selectedTasks = store?.tasks?.map((task) => task.id) || [];
		} else {
			selectedTasks = [];
		}
	}

	// 批量操作
	async function handleBatchAction(action: 'start' | 'cancel' | 'delete') {
		if (selectedTasks.length === 0) return;

		switch (action) {
			case 'start':
				// 逐个启动任务
				for (const taskId of selectedTasks) {
					await taskActions.restartTask(taskId);
				}
				break;
			case 'cancel':
				await taskActions.batchCancel(selectedTasks);
				break;
			case 'delete':
				await taskActions.batchDelete(selectedTasks);
				break;
		}

		selectedTasks = [];
		await loadTasks();
	}

	// 跳转到任务详情
	function viewTask(taskId: string) {
		goto(`/tasks/${taskId}`);
	}

	// 跳转到创建任务页面
	function createTask() {
		goto('/tasks/create');
	}

	// 计算属性
	let isAllSelected = $derived(
		store?.tasks?.length > 0 && selectedTasks.length === store.tasks.length
	);

	let isPartialSelected = $derived(
		selectedTasks.length > 0 && selectedTasks.length < (store?.tasks?.length || 0)
	);
</script>

<svelte:head>
	<title>任务管理 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6">
	<!-- 页面标题和操作 -->
	<div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">任务管理</h1>
			<p class="text-gray-600 dark:text-gray-400 mt-1">管理和监控安全扫描任务的执行状态</p>
		</div>
		<div class="flex gap-2">
			<Button variant="outline" onclick={() => (showFilters = !showFilters)}>
				<i class="fas fa-filter mr-2"></i>
				过滤器
			</Button>
			<Button onclick={createTask}>
				<i class="fas fa-plus mr-2"></i>
				创建任务
			</Button>
		</div>
	</div>

	<!-- 任务统计卡片 -->
	{#if store?.taskStats}
		<TaskStatsCards stats={store.taskStats} />
	{/if}

	<!-- 搜索和过滤器 -->
	<div
		class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 mb-6"
	>
		<div class="flex flex-col sm:flex-row gap-4">
			<div class="flex-1">
				<SearchInput
					bind:value={searchQuery}
					placeholder="搜索任务名称、描述..."
					onenter={handleSearch}
				/>
			</div>
			<div class="flex gap-2">
				<Button variant="outline" onclick={handleSearch}>
					<i class="fas fa-search mr-2"></i>
					搜索
				</Button>
			</div>
		</div>

		{#if showFilters}
			<div class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
				<TaskFilters
					bind:status={statusFilter}
					bind:type={typeFilter}
					bind:priority={priorityFilter}
					bind:projectId={projectFilter}
					projects={projects?.projects || []}
					onchange={handleFilterChange}
				/>
			</div>
		{/if}
	</div>

	<!-- 批量操作工具栏 -->
	{#if selectedTasks.length > 0}
		<div
			class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-700 rounded-lg p-4 mb-4"
		>
			<div class="flex items-center justify-between">
				<span class="text-blue-700 dark:text-blue-300">
					已选择 {selectedTasks.length} 个任务
				</span>
				<div class="flex gap-2">
					<Button size="sm" variant="outline" onclick={() => handleBatchAction('start')}>
						<i class="fas fa-play mr-1"></i>
						批量启动
					</Button>
					<Button size="sm" variant="outline" onclick={() => handleBatchAction('cancel')}>
						<i class="fas fa-stop mr-1"></i>
						批量取消
					</Button>
					<Button size="sm" variant="destructive" onclick={() => handleBatchAction('delete')}>
						<i class="fas fa-trash mr-1"></i>
						批量删除
					</Button>
				</div>
			</div>
		</div>
	{/if}

	<!-- 任务列表 -->
	<div
		class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700"
	>
		{#if store?.loading}
			<div class="p-8 text-center">
				<LoadingSpinner />
				<p class="text-gray-600 dark:text-gray-400 mt-2">加载任务列表...</p>
			</div>
		{:else if store?.error}
			<div class="p-8 text-center">
				<div class="text-red-500 mb-2">
					<i class="fas fa-exclamation-triangle text-2xl"></i>
				</div>
				<p class="text-red-600 dark:text-red-400">{store.error}</p>
				<Button variant="outline" onclick={loadTasks} class="mt-4">重新加载</Button>
			</div>
		{:else if !store?.tasks?.length}
			<div class="p-8 text-center">
				<div class="text-gray-400 mb-2">
					<i class="fas fa-tasks text-3xl"></i>
				</div>
				<p class="text-gray-600 dark:text-gray-400 mb-4">暂无任务</p>
				<Button onclick={createTask}>
					<i class="fas fa-plus mr-2"></i>
					创建第一个任务
				</Button>
			</div>
		{:else}
			<!-- 列表头部 -->
			<div class="border-b border-gray-200 dark:border-gray-700 p-4">
				<div class="flex items-center gap-4">
					<label class="flex items-center">
						<input
							type="checkbox"
							checked={isAllSelected}
							indeterminate={isPartialSelected}
							onchange={(e) => handleSelectAll((e.target as HTMLInputElement).checked)}
							class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
						/>
						<span class="ml-2 text-sm text-gray-600 dark:text-gray-400"> 全选 </span>
					</label>
					<div class="text-sm text-gray-600 dark:text-gray-400">
						共 {store.pagination?.total || 0} 个任务
					</div>
				</div>
			</div>

			<!-- 任务列表项 -->
			<div class="divide-y divide-gray-200 dark:divide-gray-700">
				{#each store.tasks as task (task.id)}
					<TaskCard
						{task}
						selected={selectedTasks.includes(task.id)}
						onselect={(selected) => handleTaskSelect(task.id, selected)}
						onclick={() => viewTask(task.id)}
					/>
				{/each}
			</div>

			<!-- 分页 -->
			{#if store.pagination && store.pagination.totalPages > 1}
				<div class="border-t border-gray-200 dark:border-gray-700 p-4">
					<Pagination
						currentPage={store.pagination.page}
						totalPages={store.pagination.totalPages}
						total={store.pagination.total}
						pageSize={store.pagination.pageSize}
						on:pageChange={handlePageChange}
					/>
				</div>
			{/if}
		{/if}
	</div>
</div>

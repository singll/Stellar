<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { taskApi } from '$lib/api/tasks';
	import type { Task } from '$lib/types/task';

	import PageLayout from '$lib/components/ui/page-layout/PageLayout.svelte';
	import StatsGrid from '$lib/components/ui/stats-grid/StatsGrid.svelte';
	import DataList from '$lib/components/ui/data-list/DataList.svelte';
	import { Button } from '$lib/components/ui/button';
	import { notifications } from '$lib/stores/notifications';

	// 响应式状态
	let tasks = $state<Task[]>([]);
	let taskStats = $state({
		total: 0,
		pending: 0,
		running: 0,
		completed: 0,
		failed: 0
	});
	let loading = $state(false);
	let searchQuery = $state('');
	let currentPage = $state(1);
	let totalPages = $state(0);

	// 加载任务数据
	async function loadTasks() {
		try {
			loading = true;
			
			// 并行获取任务列表和统计信息
			const [tasksResponse, statsResponse] = await Promise.all([
				taskApi.getTasks({
					page: currentPage,
					pageSize: 20,
					search: searchQuery || undefined
				}),
				taskApi.getTaskStats()
			]);

			tasks = tasksResponse.data || [];
			taskStats = statsResponse || {
				total: 0,
				pending: 0,
				running: 0,
				completed: 0,
				failed: 0
			};
			
			totalPages = Math.ceil((tasksResponse.total || 0) / 20);
		} catch (error) {
			console.error('加载任务数据失败:', error);
			notifications.add({
				type: 'error',
				message: '加载任务数据失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
			
			// 出错时显示空数据
			tasks = [];
			taskStats = {
				total: 0,
				pending: 0,
				running: 0,
				completed: 0,
				failed: 0
			};
		} finally {
			loading = false;
		}
	}

	// 获取任务状态颜色
	function getTaskStatusColor(status: string) {
		switch (status) {
			case 'pending':
				return 'bg-gray-100 text-gray-800';
			case 'running':
				return 'bg-blue-100 text-blue-800';
			case 'completed':
				return 'bg-green-100 text-green-800';
			case 'failed':
				return 'bg-red-100 text-red-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	}

	// 获取任务类型显示名称
	function getTaskTypeDisplay(type: string) {
		const typeMap: { [key: string]: string } = {
			subdomain: '子域名扫描',
			portscan: '端口扫描',
			vulnerability: '漏洞检测',
			webscan: 'Web扫描'
		};
		return typeMap[type] || type;
	}

	// 获取优先级颜色
	function getPriorityColor(priority: string) {
		switch (priority) {
			case 'high':
				return 'bg-red-100 text-red-800';
			case 'medium':
				return 'bg-yellow-100 text-yellow-800';
			case 'low':
				return 'bg-green-100 text-green-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	}

	// 格式化时间
	function formatTime(dateString: string) {
		return new Date(dateString).toLocaleString('zh-CN');
	}

	// 搜索处理
	const handleSearch = async (query?: string) => {
		if (query !== undefined) {
			searchQuery = query;
		}
		currentPage = 1; // 重置到第一页
		await loadTasks();
	};

	// 分页处理
	const handlePageChange = async (newPage: number) => {
		currentPage = newPage;
		await loadTasks();
	};

	// 处理编辑任务
	const handleEditTask = (taskId: string) => {
		goto(`/tasks/${taskId}/edit`);
	};

	// 删除任务
	const handleDeleteTask = async (taskId: string) => {
		if (!confirm('确定要删除这个任务吗？此操作不可逆。')) return;

		try {
			await taskApi.deleteTask(taskId);
			tasks = tasks.filter((t) => t.id !== taskId);
			notifications.add({
				type: 'success',
				message: '任务删除成功'
			});
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '删除任务失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		}
	};

	// 准备统计数据
	const statsData = $derived([
		{
			title: '总任务',
			value: taskStats.total,
			icon: 'list',
			color: 'blue' as const
		},
		{
			title: '运行中',
			value: taskStats.running,
			icon: 'play',
			color: 'blue' as const
		},
		{
			title: '已完成',
			value: taskStats.completed,
			icon: 'check-circle',
			color: 'green' as const
		},
		{
			title: '失败',
			value: taskStats.failed,
			icon: 'x-circle',
			color: 'red' as const
		}
	]);

	// 准备表格列配置
	const columns = [
		{
			key: 'name',
			title: '任务名称',
			render: (value: any, row: any) => {
				return `
					<div class="flex items-center gap-3">
						<div class="flex items-center justify-center w-8 h-8 rounded-full bg-purple-100">
							<svg class="h-4 w-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
							</svg>
						</div>
						<div>
							<span class="font-medium text-blue-600 hover:text-blue-800 cursor-pointer">${row.name}</span>
							<p class="text-xs text-gray-500 mt-1">${getTaskTypeDisplay(row.type)}</p>
						</div>
					</div>
				`;
			}
		},
		{
			key: 'status',
			title: '状态',
			render: (value: any, row: any) => {
				const statusColor = getTaskStatusColor(row.status);
				return `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColor}">${row.status}</span>`;
			}
		},
		{
			key: 'priority',
			title: '优先级',
			render: (value: any, row: any) => {
				const priorityColor = getPriorityColor(row.priority);
				return `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${priorityColor}">${row.priority}</span>`;
			}
		},
		{
			key: 'project',
			title: '所属项目',
			render: (value: any, row: any) => {
				return `<span class="font-mono text-xs bg-gray-100 px-2 py-1 rounded">${row.project}</span>`;
			}
		},
		{
			key: 'progress',
			title: '进度',
			render: (value: any, row: any) => {
				return `
					<div class="flex items-center gap-2">
						<div class="flex-1 bg-gray-200 rounded-full h-2">
							<div class="bg-blue-600 h-2 rounded-full" style="width: ${row.progress}%"></div>
						</div>
						<span class="text-xs text-gray-500">${row.progress}%</span>
					</div>
				`;
			}
		},
		{
			key: 'updated_at',
			title: '更新时间',
			render: (value: any, row: any) => {
				return `<span class="text-gray-500 text-sm">${formatTime(row.updated_at)}</span>`;
			}
		}
	];

	onMount(() => {
		loadTasks();
	});
</script>

<svelte:head>
	<title>任务管理 - Stellar</title>
</svelte:head>

<PageLayout
	title="任务管理"
	description="管理和监控安全扫描任务的执行状态"
	icon="play"
	showStats={!loading && tasks.length > 0}
	actions={[
		{
			text: '创建任务',
			icon: 'plus',
			variant: 'default',
			onClick: () => goto('/tasks/create')
		}
	]}
>
	{#snippet stats()}
		<StatsGrid stats={statsData} columns={4} />
	{/snippet}

	<DataList
		title=""
		{columns}
		data={tasks}
		{loading}
		searchPlaceholder="搜索任务名称、类型..."
		searchValue={searchQuery}
		onSearch={handleSearch}
		emptyStateTitle="暂无任务"
		emptyStateDescription="您还没有创建任何任务，开始创建第一个任务吧"
		emptyStateAction={{
			text: '创建第一个任务',
			icon: 'plus',
			onClick: () => goto('/tasks/create')
		}}
		onRowClick={(task) => goto(`/tasks/${task.id}`)}
		rowActions={(row) => [
			{
				icon: 'edit',
				title: '编辑任务',
				variant: 'ghost',
				onClick: () => handleEditTask(row.id)
			},
			{
				icon: 'trash',
				title: '删除任务',
				variant: 'ghost',
				color: 'red',
				onClick: () => handleDeleteTask(row.id)
			}
		]}
	/>

	<!-- 分页 -->
	{#if totalPages > 1}
		<div class="flex justify-center items-center gap-2 mt-8">
			<Button
				variant="outline"
				size="sm"
				disabled={currentPage === 1}
				onclick={() => handlePageChange(currentPage - 1)}
			>
				上一页
			</Button>

			{#each Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
				if (totalPages <= 7) return i + 1;
				if (currentPage <= 4) return i + 1;
				if (currentPage >= totalPages - 3) return totalPages - 6 + i;
				return currentPage - 3 + i;
			}) as page}
				{#if page === currentPage}
					<Button size="sm">{page}</Button>
				{:else}
					<Button variant="outline" size="sm" onclick={() => handlePageChange(page)}>
						{page}
					</Button>
				{/if}
			{/each}

			<Button
				variant="outline"
				size="sm"
				disabled={currentPage === totalPages}
				onclick={() => handlePageChange(currentPage + 1)}
			>
				下一页
			</Button>
		</div>
	{/if}
</PageLayout>

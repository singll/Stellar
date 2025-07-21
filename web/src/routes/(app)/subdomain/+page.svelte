<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { subdomainApi } from '$lib/api/subdomain';
	import type { SubdomainTask } from '$lib/types/subdomain';
	import PageLayout from '$lib/components/ui/page-layout/PageLayout.svelte';
	import StatsGrid from '$lib/components/ui/stats-grid/StatsGrid.svelte';
	import DataList from '$lib/components/ui/data-list/DataList.svelte';
	import { Button } from '$lib/components/ui/button';
	import { notifications } from '$lib/stores/notifications';

	// 响应式状态
	let tasks = $state<SubdomainTask[]>([]);
	let taskStats = $state({
		total: 0,
		running: 0,
		completed: 0,
		failed: 0,
		successRate: 0
	});
	let loading = $state(false);
	let searchQuery = $state('');
	let currentPage = $state(1);
	let totalPages = $state(0);

	// 加载子域名任务数据
	async function loadTasks() {
		try {
			loading = true;
			
			// 并行获取任务列表和统计信息
			const [tasksResponse, statsResponse] = await Promise.all([
				subdomainApi.getTasks({
					page: currentPage,
					limit: 20,
					target: searchQuery || undefined
				}),
				subdomainApi.getStatistics()
			]);

			tasks = tasksResponse.tasks || [];
			taskStats = {
				total: statsResponse.totalTasks || 0,
				running: tasks.filter(t => t.status === 'running').length,
				completed: statsResponse.completedTasks || 0,
				failed: statsResponse.failedTasks || 0,
				successRate: statsResponse.totalTasks > 0 ? Math.round((statsResponse.completedTasks / statsResponse.totalTasks) * 100) : 0
			};
			
			totalPages = Math.ceil((tasksResponse.total || 0) / 20);
		} catch (error) {
			console.error('加载子域名任务数据失败:', error);
			notifications.add({
				type: 'error',
				message: '加载子域名任务数据失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
			
			// 出错时显示空数据
			tasks = [];
			taskStats = {
				total: 0,
				running: 0,
				completed: 0,
				failed: 0,
				successRate: 0
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
			title: '成功率',
			value: `${taskStats.successRate}%`,
			icon: 'trending-up',
			color: 'purple' as const
		}
	]);

	// 准备表格列配置
	const columns = [
		{
			key: 'name',
			title: '任务名称',
			render: (value: any, row: SubdomainTask) => {
				return `
					<div class="flex items-center gap-3">
						<div class="flex items-center justify-center w-8 h-8 rounded-full bg-indigo-100">
							<svg class="h-4 w-4 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9v-9m0-9v9"></path>
							</svg>
						</div>
						<div>
							<span class="font-medium text-blue-600 hover:text-blue-800 cursor-pointer">${row.name}</span>
							<p class="text-xs text-gray-500 mt-1">目标: ${row.rootDomain}</p>
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
			key: 'methods',
			title: '枚举方法',
			render: (value: any, row: SubdomainTask) => {
				const methods = row.config?.methods || [];
				return `<span class="font-mono text-xs bg-gray-100 px-2 py-1 rounded">${Array.isArray(methods) ? methods.join(', ') : '-'}</span>`;
			}
		},
		{
			key: 'progress',
			title: '进度',
			render: (value: any, row: SubdomainTask) => {
				const progress = row.progress || 0;
				if (row.status === 'running' && progress > 0) {
					return `
						<div class="flex items-center gap-2">
							<div class="flex-1 bg-gray-200 rounded-full h-2">
								<div class="bg-blue-600 h-2 rounded-full" style="width: ${progress}%"></div>
							</div>
							<span class="text-xs text-gray-500">${progress}%</span>
						</div>
					`;
				} else if (row.status === 'completed') {
					return '<span class="text-sm text-green-600">完成</span>';
				}
				return '<span class="text-sm text-gray-400">-</span>';
			}
		},
		{
			key: 'subdomains_found',
			title: '发现子域名',
			render: (value: any, row: SubdomainTask) => {
				return `<span class="font-semibold text-green-600">${row.subdomainsFound || 0}</span>`;
			}
		},
		{
			key: 'created_at',
			title: '创建时间',
			render: (value: any, row: SubdomainTask) => {
				return `<span class="text-gray-500 text-sm">${formatTime(row.createdAt)}</span>`;
			}
		}
	];

	onMount(() => {
		loadTasks();
	});
</script>

<svelte:head>
	<title>子域名枚举任务 - Stellar</title>
</svelte:head>

<PageLayout
	title="子域名枚举任务"
	description="管理和监控子域名枚举任务"
	icon="globe"
	showStats={!loading && tasks.length > 0}
	actions={[
		{
			text: '新建任务',
			icon: 'plus',
			variant: 'default',
			onClick: () => goto('/subdomain/create')
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
		searchPlaceholder="搜索子域名任务、目标域名..."
		searchValue={searchQuery}
		onSearch={handleSearch}
		emptyStateTitle="暂无子域名任务"
		emptyStateDescription="您还没有创建任何子域名枚举任务，开始创建第一个任务吧"
		emptyStateAction={{
			text: '创建第一个任务',
			icon: 'plus',
			onClick: () => goto('/subdomain/create')
		}}
		onRowClick={(task) => goto(`/subdomain/${task.id}`)}
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

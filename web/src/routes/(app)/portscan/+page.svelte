<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { portScanApi } from '$lib/api/portscan';
	import type { PortScanTask } from '$lib/types/portscan';
	import PageLayout from '$lib/components/ui/page-layout/PageLayout.svelte';
	import StatsGrid from '$lib/components/ui/stats-grid/StatsGrid.svelte';
	import DataList from '$lib/components/ui/data-list/DataList.svelte';
	import { Button } from '$lib/components/ui/button';
	import { notifications } from '$lib/stores/notifications';

	// 响应式状态
	let tasks = $state<PortScanTask[]>([]);
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

	// 模拟端口扫描任务数据加载
	async function loadTasks() {
		try {
			loading = true;
			// 模拟API调用延迟
			await new Promise(resolve => setTimeout(resolve, 500));
			
			// 模拟端口扫描任务数据
			tasks = [
				{
					id: '1',
					name: '192.168.1.0/24 端口扫描',
					target: '192.168.1.0/24',
					ports: '1-1000',
					status: 'completed',
					open_ports: 23,
					created_at: '2024-01-15T10:00:00Z',
					completed_at: '2024-01-15T10:15:00Z'
				},
				{
					id: '2',
					name: '服务器端口检测',
					target: '10.0.0.50',
					ports: '80,443,22,3306',
					status: 'running',
					open_ports: 2,
					created_at: '2024-01-15T11:00:00Z',
					progress: 75
				},
				{
					id: '3',
					name: '全端口扫描',
					target: '192.168.0.100',
					ports: '1-65535',
					status: 'pending',
					open_ports: 0,
					created_at: '2024-01-15T11:30:00Z'
				}
			];

			// 更新统计数据
			taskStats = {
				total: tasks.length,
				running: tasks.filter(t => t.status === 'running').length,
				completed: tasks.filter(t => t.status === 'completed').length,
				failed: tasks.filter(t => t.status === 'failed').length,
				successRate: Math.round((tasks.filter(t => t.status === 'completed').length / tasks.length) * 100)
			};
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '加载端口扫描任务失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
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
			render: (value: any, row: any) => {
				return `
					<div class="flex items-center gap-3">
						<div class="flex items-center justify-center w-8 h-8 rounded-full bg-orange-100">
							<svg class="h-4 w-4 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"></path>
							</svg>
						</div>
						<div>
							<span class="font-medium text-blue-600 hover:text-blue-800 cursor-pointer">${row.name}</span>
							<p class="text-xs text-gray-500 mt-1">目标: ${row.target}</p>
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
			key: 'ports',
			title: '端口范围',
			render: (value: any, row: any) => {
				return `<span class="font-mono text-xs bg-gray-100 px-2 py-1 rounded">${row.ports}</span>`;
			}
		},
		{
			key: 'progress',
			title: '进度',
			render: (value: any, row: any) => {
				if (row.status === 'running' && row.progress) {
					return `
						<div class="flex items-center gap-2">
							<div class="flex-1 bg-gray-200 rounded-full h-2">
								<div class="bg-blue-600 h-2 rounded-full" style="width: ${row.progress}%"></div>
							</div>
							<span class="text-xs text-gray-500">${row.progress}%</span>
						</div>
					`;
				} else if (row.status === 'completed') {
					return '<span class="text-sm text-green-600">完成</span>';
				}
				return '<span class="text-sm text-gray-400">-</span>';
			}
		},
		{
			key: 'open_ports',
			title: '开放端口',
			render: (value: any, row: any) => {
				return `<span class="font-semibold text-green-600">${row.open_ports || 0}</span>`;
			}
		},
		{
			key: 'created_at',
			title: '创建时间',
			render: (value: any, row: any) => {
				return `<span class="text-gray-500 text-sm">${formatTime(row.created_at)}</span>`;
			}
		}
	];

	onMount(() => {
		loadTasks();
	});
</script>

<svelte:head>
	<title>端口扫描任务 - Stellar</title>
</svelte:head>

<PageLayout
	title="端口扫描任务"
	description="管理和监控端口扫描任务"
	icon="wifi"
	showStats={!loading && tasks.length > 0}
	actions={[
		{
			text: '新建任务',
			icon: 'plus',
			variant: 'default',
			onClick: () => goto('/portscan/create')
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
		searchPlaceholder="搜索端口扫描任务、目标地址..."
		emptyStateTitle="暂无端口扫描任务"
		emptyStateDescription="您还没有创建任何端口扫描任务，开始创建第一个任务吧"
		emptyStateAction={{
			text: '创建第一个任务',
			icon: 'plus',
			onClick: () => goto('/portscan/create')
		}}
		onRowClick={(task) => goto(`/portscan/${task.id}`)}
	/>
</PageLayout>

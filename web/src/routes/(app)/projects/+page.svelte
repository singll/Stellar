<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { ProjectAPI } from '$lib/api/projects';
	import type { Project, ProjectFilters } from '$lib/types/project';
	import { notifications } from '$lib/stores/notifications';
	import { Button } from '$lib/components/ui/button';
	import PageLayout from '$lib/components/ui/page-layout/PageLayout.svelte';
	import StatsGrid from '$lib/components/ui/stats-grid/StatsGrid.svelte';
	import DataList from '$lib/components/ui/data-list/DataList.svelte';
	import { onMount } from 'svelte';

	// 响应式状态
	let projects = $state<Project[]>([]);
	let stats = $state({
		total_projects: 0,
		active_projects: 0,
		total_assets: 0,
		total_vulnerabilities: 0,
		total_tasks: 0
	});
	let loading = $state(false);
	let searchQuery = $state('');
	let currentPage = $state(1);
	let totalPages = $state(0);

	// 加载项目数据
	async function loadProjects() {
		try {
			loading = true;
			
			// 并行获取项目列表和统计信息
			const [projectsResponse, statsResponse] = await Promise.all([
				ProjectAPI.getProjects({
					page: currentPage,
					limit: 20,
					search: searchQuery || undefined
				}),
				ProjectAPI.getProjectStats()
			]);

			projects = projectsResponse.data || [];
			stats = statsResponse || {
				total_projects: 0,
				active_projects: 0,
				total_assets: 0,
				total_vulnerabilities: 0,
				total_tasks: 0
			};
			
			totalPages = Math.ceil((projectsResponse.total || 0) / 20);
		} catch (error) {
			console.error('加载项目数据失败:', error);
			notifications.add({
				type: 'error',
				message: '加载项目数据失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
			
			// 出错时显示空数据
			projects = [];
			stats = {
				total_projects: 0,
				active_projects: 0,
				total_assets: 0,
				total_vulnerabilities: 0,
				total_tasks: 0
			};
		} finally {
			loading = false;
		}
	}

	// 格式化日期
	const formatDate = (dateString: string) => {
		return new Date(dateString).toLocaleDateString('zh-CN');
	};

	// 搜索处理
	const handleSearch = async (query?: string) => {
		if (query !== undefined) {
			searchQuery = query;
		}
		currentPage = 1; // 重置到第一页
		await loadProjects();
	};

	// 删除项目
	const handleDeleteProject = async (projectId: string) => {
		if (!confirm('确定要删除这个项目吗？此操作不可逆。')) return;

		try {
			await ProjectAPI.deleteProject(projectId);
			projects = projects.filter((p) => p.id !== projectId);
			notifications.add({
				type: 'success',
				message: '项目删除成功'
			});
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '删除项目失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		}
	};

	// 复制项目
	const handleDuplicateProject = async (project: Project) => {
		const newName = prompt('请输入新项目名称:', `${project.name} - 副本`);
		if (!newName) return;

		try {
			const newProject = await ProjectAPI.duplicateProject(project.id, newName);
			projects = [newProject, ...projects];
			notifications.add({
				type: 'success',
				message: '项目复制成功'
			});
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '复制项目失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		}
	};

	// 分页处理
	const handlePageChange = async (newPage: number) => {
		currentPage = newPage;
		await loadProjects();
	};

	// 获取扫描状态颜色
	const getScanStatusColor = (status: string) => {
		switch (status) {
			case 'running':
				return 'bg-blue-100 text-blue-800';
			case 'completed':
				return 'bg-green-100 text-green-800';
			case 'failed':
				return 'bg-red-100 text-red-800';
			case 'paused':
				return 'bg-yellow-100 text-yellow-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	};

	// 准备统计数据
	const statsData = $derived([
		{
			title: '总项目',
			value: stats.total_projects,
			icon: 'folder',
			color: 'blue' as const
		},
		{
			title: '活跃项目',
			value: stats.active_projects,
			icon: 'activity',
			color: 'green' as const
		},
		{
			title: '总资产',
			value: stats.total_assets,
			icon: 'database',
			color: 'purple' as const
		},
		{
			title: '发现漏洞',
			value: stats.total_vulnerabilities,
			icon: 'alert-triangle',
			color: 'red' as const
		}
	]);

	// 准备表格列配置
	const columns = [
		{
			key: 'name',
			title: '项目名称',
			render: (value: any, row: Project) => {
				return `
					<div class="flex items-center gap-3">
						<div class="flex items-center justify-center w-8 h-8 rounded-full bg-blue-100">
							<svg class="h-4 w-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2z"></path>
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5a2 2 0 012-2h4a2 2 0 012 2v3H8V5z"></path>
							</svg>
						</div>
						<div>
							<span class="font-medium text-blue-600 hover:text-blue-800 cursor-pointer">${row.name}</span>
							${row.description ? `<p class="text-xs text-gray-500 mt-1 line-clamp-2">${row.description}</p>` : ''}
						</div>
					</div>
				`;
			}
		},
		{
			key: 'target',
			title: '目标',
			render: (value: any, row: Project) => {
				return row.target ? `<span class="font-mono text-xs bg-gray-100 px-2 py-1 rounded">${row.target}</span>` : '-';
			}
		},
		{
			key: 'status',
			title: '状态',
			render: (value: any, row: Project) => {
				const statusColor = getScanStatusColor(row.scan_status || 'unknown');
				return `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColor}">${row.scan_status || '未知'}</span>`;
			}
		},
		{
			key: 'stats',
			title: '统计',
			render: (value: any, row: Project) => {
				return `
					<div class="flex items-center gap-4 text-sm">
						<span class="text-gray-500">资产: <span class="font-semibold text-gray-900">${row.assets_count || 0}</span></span>
						<span class="text-gray-500">漏洞: <span class="font-semibold text-red-600">${row.vulnerabilities_count || 0}</span></span>
					</div>
				`;
			}
		},
		{
			key: 'created_at',
			title: '创建时间',
			render: (value: any, row: Project) => {
				return `<span class="text-gray-500 text-sm">${formatDate(row.created_at)}</span>`;
			}
		}
	];

	onMount(() => {
		loadProjects();
	});
</script>

<svelte:head>
	<title>项目管理 - Stellar</title>
</svelte:head>

<PageLayout
	title="项目管理"
	description="管理和监控您的安全扫描项目"
	icon="briefcase"
	showStats={true}
	actions={[
		{
			text: '创建项目',
			icon: 'plus',
			variant: 'default',
			onClick: () => goto('/projects/create')
		}
	]}
>
	{#snippet stats()}
		<StatsGrid stats={statsData} columns={4} />
	{/snippet}

	<DataList
		title=""
		{columns}
		data={projects}
		{loading}
		searchPlaceholder="搜索项目名称、描述或目标..."
		searchValue={searchQuery}
		onSearch={handleSearch}
		emptyStateTitle="暂无项目"
		emptyStateDescription="您还没有创建任何项目，开始创建第一个项目吧"
		emptyStateAction={{
			text: '创建第一个项目',
			icon: 'plus',
			onClick: () => goto('/projects/create')
		}}
		onRowClick={(project) => goto(`/projects/${project.id}`)}
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

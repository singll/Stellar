<!--
主页仪表盘 - 完善版本
显示系统概览、快速操作、最近活动、统计图表等
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	// API导入
	import { assetApi } from '$lib/api/asset';
	import { taskApi } from '$lib/api/tasks';
	import { nodeAPI } from '$lib/api/nodes';
	import { ProjectAPI } from '$lib/api/projects';

	// 组件导入
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import Button from '$lib/components/ui/Button.svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import ProgressBar from '$lib/components/ui/ProgressBar.svelte';

	// 类型导入
	import type { Asset } from '$lib/types/asset';
	import type { Task } from '$lib/types/task';
	import type { Node, NodeStats } from '$lib/types/node';
	import type { Project } from '$lib/types/project';
	import { formatDateTime } from '$lib/utils/date';

	// 状态变量
	let loading = $state(false);
	let error = $state<string | null>(null);

	// 统计数据
	let stats = $state({
		assets: { total: 0, active: 0, domains: 0, ips: 0 },
		tasks: { total: 0, running: 0, completed: 0, failed: 0 },
		nodes: { total: 0, online: 0, offline: 0 },
		projects: { total: 0, active: 0 }
	});

	// 最近数据
	let recentTasks: Task[] = $state([]);
	let recentAssets: Asset[] = $state([]);
	let nodeStats: NodeStats | null = $state(null);
	let activeProjects: Project[] = $state([]);

	// 系统状态
	let systemHealth = $state({
		overall: 100,
		database: 100,
		cache: 100,
		nodes: 100,
		storage: 85
	});

	// 加载仪表盘数据
	async function loadDashboardData() {
		try {
			loading = true;
			error = null;

			// 并发加载各种数据
			const [
				assetsResponse,
				tasksResponse,
				nodesStatsResponse,
				projectsResponse,
				recentTasksResponse
			] = await Promise.allSettled([
				assetApi.getAssets({ type: 'domain', page: 1, pageSize: 1 }),
				taskApi.getTasks({ page: 1, pageSize: 1 }),
				nodeAPI.getNodeStats(),
				ProjectAPI.getProjects({ page: 1, limit: 5 }),
				taskApi.getTasks({ page: 1, pageSize: 5, sortBy: 'createdAt', sortDesc: true })
			]);

			// 处理资产统计
			if (assetsResponse.status === 'fulfilled') {
				const assetsData = assetsResponse.value;
				stats.assets.total = assetsData.data.total;

				// 获取详细的资产统计
				try {
					const assetStats = await assetApi.getAssetStats();
					stats.assets.active = assetStats.data.total;
					stats.assets.domains = assetStats.data.byType?.domain || 0;
					stats.assets.ips = assetStats.data.byType?.ip || 0;

					// 获取最近资产
					const recentAssetsData = await assetApi.getAssets({
						type: 'domain',
						page: 1,
						pageSize: 5,
						sortBy: 'createdAt',
						sortDesc: true
					});
					recentAssets = recentAssetsData.data.items;
				} catch (err) {
					console.warn('加载资产详细统计失败:', err);
				}
			}

			// 处理任务统计
			if (tasksResponse.status === 'fulfilled') {
				const tasksData = tasksResponse.value;
				stats.tasks.total = tasksData.data.total;

				try {
					const taskStats = await taskApi.getTaskStats();
					stats.tasks.running = taskStats.data.byStatus?.running || 0;
					stats.tasks.completed = taskStats.data.byStatus?.completed || 0;
					stats.tasks.failed = taskStats.data.byStatus?.failed || 0;
				} catch (err) {
					console.warn('加载任务详细统计失败:', err);
				}
			}

			// 处理最近任务
			if (recentTasksResponse.status === 'fulfilled') {
				recentTasks = recentTasksResponse.value.data.items;
			}

			// 处理节点统计
			if (nodesStatsResponse.status === 'fulfilled') {
				nodeStats = nodesStatsResponse.value.data;
				stats.nodes.total = nodeStats.total;
				stats.nodes.online = nodeStats.online;
				stats.nodes.offline = nodeStats.offline;

				// 计算节点健康度
				if (nodeStats.total > 0) {
					systemHealth.nodes = Math.round((nodeStats.online / nodeStats.total) * 100);
				}
			}

			// 处理项目统计
			if (projectsResponse.status === 'fulfilled') {
				const projectsData = projectsResponse.value;
				stats.projects.total = projectsData.total;
				activeProjects = projectsData.data;
				stats.projects.active = projectsData.data.length;
			}

			// 计算整体系统健康度
			systemHealth.overall = Math.round(
				(systemHealth.database + systemHealth.cache + systemHealth.nodes + systemHealth.storage) / 4
			);
		} catch (err) {
			error = err instanceof Error ? err.message : '加载仪表盘数据失败';
		} finally {
			loading = false;
		}
	}

	// 获取任务状态标签样式
	function getTaskStatusVariant(status: string) {
		switch (status) {
			case 'running':
				return 'warning';
			case 'completed':
				return 'success';
			case 'failed':
				return 'danger';
			case 'pending':
				return 'secondary';
			default:
				return 'secondary';
		}
	}

	// 获取任务状态文本
	function getTaskStatusText(status: string) {
		switch (status) {
			case 'running':
				return '运行中';
			case 'completed':
				return '已完成';
			case 'failed':
				return '失败';
			case 'pending':
				return '等待中';
			case 'paused':
				return '已暂停';
			case 'cancelled':
				return '已取消';
			default:
				return '未知';
		}
	}

	// 获取任务类型文本
	function getTaskTypeText(type: string) {
		switch (type) {
			case 'subdomain_enum':
				return '子域名枚举';
			case 'port_scan':
				return '端口扫描';
			case 'vuln_scan':
				return '漏洞扫描';
			case 'asset_discovery':
				return '资产发现';
			default:
				return type;
		}
	}

	// 获取健康度颜色
	function getHealthColor(score: number) {
		if (score >= 90) return 'text-green-600';
		if (score >= 70) return 'text-yellow-600';
		return 'text-red-600';
	}

	// 获取健康度进度条样式
	function getHealthVariant(score: number) {
		if (score >= 90) return 'success';
		if (score >= 70) return 'warning';
		return 'danger';
	}

	// 页面加载
	onMount(() => {
		loadDashboardData();

		// 定时刷新数据（每30秒）
		const interval = setInterval(loadDashboardData, 30000);

		return () => {
			clearInterval(interval);
		};
	});
</script>

<svelte:head>
	<title>仪表盘 - Stellar</title>
</svelte:head>

<div class="space-y-8">
	<!-- 页面标题 -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-4xl font-bold bg-gradient-to-r from-slate-800 to-blue-600 bg-clip-text text-transparent">仪表盘</h1>
			<p class="text-slate-600 mt-2">系统概览和快速操作</p>
		</div>
		<div class="flex items-center space-x-3">
			<button 
				onclick={loadDashboardData} 
				disabled={loading}
				class="modern-btn-secondary"
			>
				{loading ? '刷新中...' : '🔄 刷新'}
			</button>
			<button onclick={() => goto('/tasks/create')} class="modern-btn-primary">
				⚡ 创建任务
			</button>
		</div>
	</div>

	<!-- 错误提示 -->
	{#if error}
		<div class="notification-error">
			<div class="flex items-center justify-between">
				<div class="flex items-center space-x-2">
					<span>❌</span>
					<span>{error}</span>
				</div>
				<button onclick={() => (error = null)} class="modern-btn-ghost text-red-600 hover:text-red-800">
					关闭
				</button>
			</div>
		</div>
	{/if}

	<!-- 核心统计卡片 -->
	<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
		<div class="modern-card hover:scale-105 transition-all duration-300">
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-sm font-medium text-slate-600">资产总数</h3>
				<div class="w-10 h-10 bg-gradient-to-br from-blue-400 to-blue-600 rounded-xl flex items-center justify-center text-white shadow-soft">
					🎯
				</div>
			</div>
			<div class="space-y-2">
				<div class="text-3xl font-bold text-slate-800">{stats.assets.total.toLocaleString()}</div>
				<p class="text-xs text-slate-500">
					域名 {stats.assets.domains} · IP {stats.assets.ips}
				</p>
				<button onclick={() => goto('/assets')} class="modern-btn-ghost text-blue-600 hover:text-blue-700 text-sm">
					查看资产
				</button>
			</div>
		</div>

		<div class="modern-card hover:scale-105 transition-all duration-300">
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-sm font-medium text-slate-600">活跃任务</h3>
				<div class="w-10 h-10 bg-gradient-to-br from-orange-400 to-orange-600 rounded-xl flex items-center justify-center text-white shadow-soft">
					⚡
				</div>
			</div>
			<div class="space-y-2">
				<div class="text-3xl font-bold text-slate-800">{stats.tasks.running}</div>
				<p class="text-xs text-slate-500">
					总计 {stats.tasks.total} · 已完成 {stats.tasks.completed}
				</p>
				<button onclick={() => goto('/tasks')} class="modern-btn-ghost text-orange-600 hover:text-orange-700 text-sm">
					管理任务
				</button>
			</div>
		</div>

		<div class="modern-card hover:scale-105 transition-all duration-300">
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-sm font-medium text-slate-600">在线节点</h3>
				<div class="w-10 h-10 bg-gradient-to-br from-green-400 to-green-600 rounded-xl flex items-center justify-center text-white shadow-soft">
					🖥️
				</div>
			</div>
			<div class="space-y-2">
				<div class="text-3xl font-bold text-slate-800">{stats.nodes.online}/{stats.nodes.total}</div>
				<p class="text-xs text-slate-500">
					离线 {stats.nodes.offline} 个节点
				</p>
				<button onclick={() => goto('/nodes')} class="modern-btn-ghost text-green-600 hover:text-green-700 text-sm">
					节点管理
				</button>
			</div>
		</div>

		<div class="modern-card hover:scale-105 transition-all duration-300">
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-sm font-medium text-slate-600">活跃项目</h3>
				<div class="w-10 h-10 bg-gradient-to-br from-purple-400 to-purple-600 rounded-xl flex items-center justify-center text-white shadow-soft">
					📁
				</div>
			</div>
			<div class="space-y-2">
				<div class="text-3xl font-bold text-slate-800">{stats.projects.active}</div>
				<p class="text-xs text-slate-500">
					总计 {stats.projects.total} 个项目
				</p>
				<button onclick={() => goto('/projects')} class="modern-btn-ghost text-purple-600 hover:text-purple-700 text-sm">
					项目管理
				</button>
			</div>
		</div>
	</div>

	<!-- 主要内容区域 -->
	<div class="grid gap-6 lg:grid-cols-3">
		<!-- 左栏：系统健康状态 + 快速操作 -->
		<div class="space-y-6">
			<!-- 系统健康状态 -->
			<div class="modern-card">
				<div class="flex items-center space-x-2 mb-6">
					<div class="w-8 h-8 bg-gradient-to-br from-green-400 to-emerald-500 rounded-lg flex items-center justify-center">
						🩺
					</div>
					<h3 class="text-lg font-semibold text-slate-800">系统健康状态</h3>
				</div>
				
				<div class="space-y-4">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-slate-700">整体状态</span>
						<span class="text-sm font-bold {getHealthColor(systemHealth.overall)}">
							{systemHealth.overall}%
						</span>
					</div>
					<div class="w-full bg-slate-200 rounded-full h-2">
						<div 
							class="h-2 rounded-full bg-gradient-to-r from-green-400 to-emerald-500 transition-all duration-300"
							style="width: {systemHealth.overall}%"
						></div>
					</div>

					<div class="space-y-3 text-sm">
						<div class="flex justify-between items-center">
							<span class="text-slate-600">数据库</span>
							<div class="flex items-center space-x-2">
								<div class="w-12 bg-slate-200 rounded-full h-1">
									<div 
										class="h-1 rounded-full bg-green-400 transition-all"
										style="width: {systemHealth.database}%"
									></div>
								</div>
								<span class={getHealthColor(systemHealth.database)}>{systemHealth.database}%</span>
							</div>
						</div>
						<div class="flex justify-between items-center">
							<span class="text-slate-600">缓存服务</span>
							<div class="flex items-center space-x-2">
								<div class="w-12 bg-slate-200 rounded-full h-1">
									<div 
										class="h-1 rounded-full bg-blue-400 transition-all"
										style="width: {systemHealth.cache}%"
									></div>
								</div>
								<span class={getHealthColor(systemHealth.cache)}>{systemHealth.cache}%</span>
							</div>
						</div>
						<div class="flex justify-between items-center">
							<span class="text-slate-600">计算节点</span>
							<div class="flex items-center space-x-2">
								<div class="w-12 bg-slate-200 rounded-full h-1">
									<div 
										class="h-1 rounded-full bg-purple-400 transition-all"
										style="width: {systemHealth.nodes}%"
									></div>
								</div>
								<span class={getHealthColor(systemHealth.nodes)}>{systemHealth.nodes}%</span>
							</div>
						</div>
						<div class="flex justify-between items-center">
							<span class="text-slate-600">存储空间</span>
							<div class="flex items-center space-x-2">
								<div class="w-12 bg-slate-200 rounded-full h-1">
									<div 
										class="h-1 rounded-full bg-orange-400 transition-all"
										style="width: {systemHealth.storage}%"
									></div>
								</div>
								<span class={getHealthColor(systemHealth.storage)}>{systemHealth.storage}%</span>
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- 快速操作 -->
			<div class="modern-card">
				<div class="flex items-center space-x-2 mb-6">
					<div class="w-8 h-8 bg-gradient-to-br from-blue-400 to-blue-500 rounded-lg flex items-center justify-center">
						⚡
					</div>
					<h3 class="text-lg font-semibold text-slate-800">快速操作</h3>
				</div>
				
				<div class="space-y-3">
					<button
						onclick={() => goto('/tasks/create')}
						class="modern-btn-ghost w-full justify-start text-blue-600 hover:text-blue-700 hover:bg-blue-50"
					>
						<span class="mr-3">🎯</span>
						创建扫描任务
					</button>
					<button
						onclick={() => goto('/projects/create')}
						class="modern-btn-ghost w-full justify-start text-purple-600 hover:text-purple-700 hover:bg-purple-50"
					>
						<span class="mr-3">📁</span>
						新建项目
					</button>
					<button 
						onclick={() => goto('/assets')} 
						class="modern-btn-ghost w-full justify-start text-green-600 hover:text-green-700 hover:bg-green-50"
					>
						<span class="mr-3">🔍</span>
						导入资产
					</button>
					<button
						onclick={() => goto('/nodes/create')}
						class="modern-btn-ghost w-full justify-start text-orange-600 hover:text-orange-700 hover:bg-orange-50"
					>
						<span class="mr-3">🖥️</span>
						添加节点
					</button>
					<button 
						onclick={() => goto('/settings')} 
						class="modern-btn-ghost w-full justify-start text-slate-600 hover:text-slate-700 hover:bg-slate-50"
					>
						<span class="mr-3">⚙️</span>
						系统设置
					</button>
				</div>
			</div>
		</div>

		<!-- 中栏：最近任务 -->
		<div class="modern-card">
			<div class="flex items-center justify-between mb-6">
				<div class="flex items-center space-x-2">
					<div class="w-8 h-8 bg-gradient-to-br from-orange-400 to-orange-500 rounded-lg flex items-center justify-center">
						📋
					</div>
					<h3 class="text-lg font-semibold text-slate-800">最近任务</h3>
				</div>
				<button onclick={() => goto('/tasks')} class="modern-btn-ghost text-orange-600 hover:text-orange-700 text-sm">
					查看全部
				</button>
			</div>
			
			{#if loading}
				<div class="flex items-center justify-center py-12">
					<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-orange-500"></div>
				</div>
			{:else if recentTasks.length === 0}
				<div class="text-center py-12 text-slate-500">
					<div class="text-6xl mb-4">📝</div>
					<p class="text-lg font-medium mb-2">暂无任务</p>
					<p class="text-sm mb-4">创建您的第一个扫描任务</p>
					<button onclick={() => goto('/tasks/create')} class="modern-btn-primary">
						创建任务
					</button>
				</div>
			{:else}
				<div class="space-y-3">
					{#each recentTasks as task}
						<div class="flex items-center justify-between p-4 border border-slate-200 rounded-xl hover:bg-slate-50 transition-all duration-200">
							<div class="flex-1">
								<div class="flex items-center space-x-2 mb-2">
									<span class="font-medium text-slate-800">{task.name}</span>
									<div class="status-badge status-{getTaskStatusVariant(task.status)}">
										{getTaskStatusText(task.status)}
									</div>
								</div>
								<div class="text-xs text-slate-500">
									{getTaskTypeText(task.type)} · {formatDateTime(task.createdAt)}
								</div>
								{#if task.status === 'running' && task.progress}
									<div class="mt-2">
										<div class="w-full bg-slate-200 rounded-full h-1">
											<div 
												class="h-1 rounded-full bg-gradient-to-r from-orange-400 to-orange-500 transition-all duration-300"
												style="width: {task.progress}%"
											></div>
										</div>
									</div>
								{/if}
							</div>
							<button onclick={() => goto(`/tasks/${task.id}`)} class="modern-btn-ghost text-orange-600 hover:text-orange-700 ml-4">
								查看
							</button>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- 右栏：活跃项目 + 最近资产 -->
		<div class="space-y-6">
			<!-- 活跃项目 -->
			<div class="modern-card">
				<div class="flex items-center justify-between mb-6">
					<div class="flex items-center space-x-2">
						<div class="w-8 h-8 bg-gradient-to-br from-purple-400 to-purple-500 rounded-lg flex items-center justify-center">
							📁
						</div>
						<h3 class="text-lg font-semibold text-slate-800">活跃项目</h3>
					</div>
					<button onclick={() => goto('/projects')} class="modern-btn-ghost text-purple-600 hover:text-purple-700 text-sm">
						查看全部
					</button>
				</div>
				
				{#if activeProjects.length === 0}
					<div class="text-center py-8 text-slate-500">
						<div class="text-4xl mb-3">📂</div>
						<p class="text-sm font-medium mb-2">暂无项目</p>
						<button
							onclick={() => goto('/projects/create')}
							class="modern-btn-ghost text-purple-600 hover:text-purple-700"
						>
							创建项目
						</button>
					</div>
				{:else}
					<div class="space-y-3">
						{#each activeProjects as project}
							<div class="flex items-center justify-between p-3 border border-slate-200 rounded-xl hover:bg-slate-50 transition-all duration-200">
								<div class="flex-1">
									<div class="font-medium text-slate-800 mb-1">{project.name}</div>
									<div class="text-xs text-slate-500">
										{formatDateTime(project.createdAt)}
									</div>
								</div>
								<button
									onclick={() => goto(`/projects/${project.id}`)}
									class="modern-btn-ghost text-purple-600 hover:text-purple-700"
								>
									进入
								</button>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- 最近资产 -->
			<div class="modern-card">
				<div class="flex items-center justify-between mb-6">
					<div class="flex items-center space-x-2">
						<div class="w-8 h-8 bg-gradient-to-br from-blue-400 to-blue-500 rounded-lg flex items-center justify-center">
							🎯
						</div>
						<h3 class="text-lg font-semibold text-slate-800">最近资产</h3>
					</div>
					<button onclick={() => goto('/assets')} class="modern-btn-ghost text-blue-600 hover:text-blue-700 text-sm">
						查看全部
					</button>
				</div>
				
				{#if recentAssets.length === 0}
					<div class="text-center py-8 text-slate-500">
						<div class="text-4xl mb-3">🎯</div>
						<p class="text-sm font-medium mb-2">暂无资产</p>
						<button onclick={() => goto('/assets')} class="modern-btn-ghost text-blue-600 hover:text-blue-700">
							添加资产
						</button>
					</div>
				{:else}
					<div class="space-y-2">
						{#each recentAssets as asset}
							<div class="flex items-center justify-between p-3 border border-slate-200 rounded-lg hover:bg-slate-50 transition-all duration-200">
								<div class="flex-1">
									<div class="font-medium text-slate-800 text-sm mb-1">{asset.value}</div>
									<div class="text-xs text-slate-500 uppercase tracking-wide">{asset.type}</div>
								</div>
								<button onclick={() => goto(`/assets/${asset.id}`)} class="modern-btn-ghost text-blue-600 hover:text-blue-700">
									查看
								</button>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>

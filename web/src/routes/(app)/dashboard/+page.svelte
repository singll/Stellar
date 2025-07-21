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
	import { Button } from '$lib/components/ui/button';
	import ProgressBar from '$lib/components/ui/ProgressBar.svelte';

	// 现代化图标
	import Icon from '@iconify/svelte';

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
			if (assetsResponse.status === 'fulfilled' && assetsResponse.value.data) {
				const assetsData = assetsResponse.value;
				stats.assets.total = assetsData.data.total || 0;

				// 获取详细的资产统计
				try {
					const assetStats = await assetApi.getAssetStats();
					// assetStats.data 是 Record<string, number>，需要解析结构
					const assetData = assetStats.data as Record<string, number> || {};
					stats.assets.active = assetData.total || 0;
					stats.assets.domains = assetData.domain || 0;
					stats.assets.ips = assetData.ip || 0;

					// 获取最近资产
					const recentAssetsData = await assetApi.getAssets({
						type: 'domain',
						page: 1,
						pageSize: 5,
						sortBy: 'createdAt',
						sortDesc: true
					});
					recentAssets = recentAssetsData.data?.items || [];
				} catch (err) {
					console.warn('加载资产详细统计失败:', err);
				}
			}

			// 处理任务统计
			if (tasksResponse.status === 'fulfilled' && tasksResponse.value.data) {
				const tasksData = tasksResponse.value;
				stats.tasks.total = tasksData.data.total || 0;

				try {
					const taskStats = await taskApi.getTaskStats();
					// TaskStats 直接包含状态计数，不需要 byStatus
					stats.tasks.running = taskStats.data.running || 0;
					stats.tasks.completed = taskStats.data.completed || 0;
					stats.tasks.failed = taskStats.data.failed || 0;
				} catch (err) {
					console.warn('加载任务详细统计失败:', err);
				}
			}

			// 处理最近任务
			if (recentTasksResponse.status === 'fulfilled' && recentTasksResponse.value.data) {
				recentTasks = recentTasksResponse.value.data.items || [];
			}

			// 处理节点统计
			if (nodesStatsResponse.status === 'fulfilled' && nodesStatsResponse.value) {
				nodeStats = nodesStatsResponse.value;
				if (nodeStats) {
					stats.nodes.total = nodeStats.total || 0;
					stats.nodes.online = nodeStats.online || 0;
					stats.nodes.offline = nodeStats.offline || 0;

					// 计算节点健康度
					if (nodeStats.total > 0) {
						systemHealth.nodes = Math.round((nodeStats.online / nodeStats.total) * 100);
					}
				}
			}

			// 处理项目统计
			if (projectsResponse.status === 'fulfilled' && projectsResponse.value.data) {
				const projectsData = projectsResponse.value;
				stats.projects.total = projectsData.data.total || 0;
				activeProjects = projectsData.data.items || [];
				stats.projects.active = activeProjects.length;
			}

			// 计算整体系统健康度
			systemHealth.overall = Math.round(
				(systemHealth.database + systemHealth.cache + systemHealth.nodes + systemHealth.storage) / 4
			);
			
			// 确保所有统计值都是有效的数字
			Object.keys(stats).forEach(key => {
				const category = key as keyof typeof stats;
				Object.keys(stats[category]).forEach(subKey => {
					const value = stats[category][subKey as keyof typeof stats[typeof key]];
					if (typeof value !== 'number' || isNaN(value)) {
						stats[category][subKey as keyof typeof stats[typeof key]] = 0;
					}
				});
			});
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
				return 'secondary';
			case 'completed':
				return 'default';
			case 'failed':
				return 'destructive';
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
		if (score >= 90) return 'green';
		if (score >= 70) return 'yellow';
		return 'red';
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
			<h1 class="text-4xl font-bold bg-gradient-to-r from-slate-800 to-blue-600 bg-clip-text text-transparent">
				仪表盘
			</h1>
			<p class="text-slate-600 dark:text-slate-400 mt-2">系统概览和快速操作</p>
		</div>
		<div class="flex items-center space-x-3">
			<Button
				variant="outline"
				onclick={loadDashboardData}
				disabled={loading}
				class="h-10"
			>
				<Icon icon="tabler:refresh" width={16} class="mr-2 {loading ? 'animate-spin' : ''}" />
				{loading ? '刷新中...' : '刷新'}
			</Button>
			<Button
				onclick={() => goto('/tasks/create')}
				class="h-10 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
			>
				<Icon icon="tabler:plus" width={16} class="mr-2" />
				创建任务
			</Button>
		</div>
	</div>

	<!-- 错误提示 -->
	{#if error}
		<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4">
			<div class="flex items-center justify-between">
				<div class="flex items-center space-x-2">
					<Icon icon="tabler:circle-x" width={20} class="text-red-600 dark:text-red-400" />
					<span class="text-red-800 dark:text-red-200">{error}</span>
				</div>
				<Button
					variant="ghost"
					size="sm"
					onclick={() => (error = null)}
					class="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-200"
				>
					关闭
				</Button>
			</div>
		</div>
	{/if}

	<!-- 核心统计卡片 -->
	<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
		<Card class="hover:shadow-lg transition-all duration-300 hover:scale-[1.02] bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm">
			<CardContent class="p-6">
				<div class="flex items-center justify-between mb-4">
					<h3 class="text-sm font-medium text-slate-600 dark:text-slate-400">资产总数</h3>
					<div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl flex items-center justify-center shadow-lg">
						<Icon icon="tabler:target" width={20} class="text-white" />
					</div>
				</div>
				<div class="space-y-2">
					<div class="text-3xl font-bold text-slate-900 dark:text-slate-100">{stats.assets.total.toLocaleString()}</div>
					<p class="text-xs text-slate-500 dark:text-slate-400">
						域名 {stats.assets.domains} · IP {stats.assets.ips}
					</p>
					<Button
						variant="ghost"
						size="sm"
						onclick={() => goto('/assets')}
						class="text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 p-0 h-auto"
					>
						查看资产
						<Icon icon="tabler:arrow-right" width={14} class="ml-1" />
					</Button>
				</div>
			</CardContent>
		</Card>

		<Card class="hover:shadow-lg transition-all duration-300 hover:scale-[1.02] bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm">
			<CardContent class="p-6">
				<div class="flex items-center justify-between mb-4">
					<h3 class="text-sm font-medium text-slate-600 dark:text-slate-400">活跃任务</h3>
					<div class="w-10 h-10 bg-gradient-to-br from-orange-500 to-orange-600 rounded-xl flex items-center justify-center shadow-lg">
						<Icon icon="tabler:bolt" width={20} class="text-white" />
					</div>
				</div>
				<div class="space-y-2">
					<div class="text-3xl font-bold text-slate-900 dark:text-slate-100">{stats.tasks.running}</div>
					<p class="text-xs text-slate-500 dark:text-slate-400">
						总计 {stats.tasks.total} · 已完成 {stats.tasks.completed}
					</p>
					<Button
						variant="ghost"
						size="sm"
						onclick={() => goto('/tasks')}
						class="text-orange-600 hover:text-orange-700 dark:text-orange-400 dark:hover:text-orange-300 p-0 h-auto"
					>
						管理任务
						<Icon icon="tabler:arrow-right" width={14} class="ml-1" />
					</Button>
				</div>
			</CardContent>
		</Card>

		<Card class="hover:shadow-lg transition-all duration-300 hover:scale-[1.02] bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm">
			<CardContent class="p-6">
				<div class="flex items-center justify-between mb-4">
					<h3 class="text-sm font-medium text-slate-600 dark:text-slate-400">在线节点</h3>
					<div class="w-10 h-10 bg-gradient-to-br from-green-500 to-green-600 rounded-xl flex items-center justify-center shadow-lg">
						<Icon icon="tabler:device-desktop" width={20} class="text-white" />
					</div>
				</div>
				<div class="space-y-2">
					<div class="text-3xl font-bold text-slate-900 dark:text-slate-100">
						{stats.nodes.online}/{stats.nodes.total}
					</div>
					<p class="text-xs text-slate-500 dark:text-slate-400">
						离线 {stats.nodes.offline} 个节点
					</p>
					<Button
						variant="ghost"
						size="sm"
						onclick={() => goto('/nodes')}
						class="text-green-600 hover:text-green-700 dark:text-green-400 dark:hover:text-green-300 p-0 h-auto"
					>
						节点管理
						<Icon icon="tabler:arrow-right" width={14} class="ml-1" />
					</Button>
				</div>
			</CardContent>
		</Card>

		<Card class="hover:shadow-lg transition-all duration-300 hover:scale-[1.02] bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm">
			<CardContent class="p-6">
				<div class="flex items-center justify-between mb-4">
					<h3 class="text-sm font-medium text-slate-600 dark:text-slate-400">活跃项目</h3>
					<div class="w-10 h-10 bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg">
						<Icon icon="tabler:folder" width={20} class="text-white" />
					</div>
				</div>
				<div class="space-y-2">
					<div class="text-3xl font-bold text-slate-900 dark:text-slate-100">{stats.projects.active}</div>
					<p class="text-xs text-slate-500 dark:text-slate-400">
						总计 {stats.projects.total} 个项目
					</p>
					<Button
						variant="ghost"
						size="sm"
						onclick={() => goto('/projects')}
						class="text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300 p-0 h-auto"
					>
						项目管理
						<Icon icon="tabler:arrow-right" width={14} class="ml-1" />
					</Button>
				</div>
			</CardContent>
		</Card>
	</div>

	<!-- 主要内容区域 -->
	<div class="grid gap-6 lg:grid-cols-3">
		<!-- 左栏：系统健康状态 + 快速操作 -->
		<div class="space-y-6">
			<!-- 系统健康状态 -->
			<Card class="bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm">
				<CardHeader class="pb-4">
					<div class="flex items-center space-x-2">
						<div class="w-8 h-8 bg-gradient-to-br from-green-500 to-emerald-600 rounded-lg flex items-center justify-center">
							<Icon icon="tabler:activity" width={18} class="text-white" />
						</div>
						<CardTitle class="text-lg text-slate-900 dark:text-slate-100">系统健康状态</CardTitle>
					</div>
				</CardHeader>
				<CardContent class="space-y-4">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-slate-700 dark:text-slate-300">整体状态</span>
						<span class="text-sm font-bold {getHealthColor(systemHealth.overall)}">
							{systemHealth.overall}%
						</span>
					</div>
					<div class="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2">
						<div
							class="h-2 rounded-full bg-gradient-to-r from-green-500 to-emerald-600 transition-all duration-300"
							style="width: {systemHealth.overall}%"
						></div>
					</div>

					<div class="space-y-3 text-sm">
						<div class="flex justify-between items-center">
							<div class="flex items-center space-x-2">
								<Icon icon="tabler:database" width={14} class="text-slate-500" />
								<span class="text-slate-600 dark:text-slate-400">数据库</span>
							</div>
							<div class="flex items-center space-x-2">
								<div class="w-12 bg-slate-200 dark:bg-slate-700 rounded-full h-1">
									<div
										class="h-1 rounded-full bg-green-500 transition-all"
										style="width: {systemHealth.database}%"
									></div>
								</div>
								<span class="{getHealthColor(systemHealth.database)} font-medium">{systemHealth.database}%</span>
							</div>
						</div>
						<div class="flex justify-between items-center">
							<div class="flex items-center space-x-2">
								<Icon icon="tabler:wifi" width={14} class="text-slate-500" />
								<span class="text-slate-600 dark:text-slate-400">缓存服务</span>
							</div>
							<div class="flex items-center space-x-2">
								<div class="w-12 bg-slate-200 dark:bg-slate-700 rounded-full h-1">
									<div
										class="h-1 rounded-full bg-blue-500 transition-all"
										style="width: {systemHealth.cache}%"
									></div>
								</div>
								<span class="{getHealthColor(systemHealth.cache)} font-medium">{systemHealth.cache}%</span>
							</div>
						</div>
						<div class="flex justify-between items-center">
							<div class="flex items-center space-x-2">
								<Icon icon="tabler:device-desktop" width={14} class="text-slate-500" />
								<span class="text-slate-600 dark:text-slate-400">计算节点</span>
							</div>
							<div class="flex items-center space-x-2">
								<div class="w-12 bg-slate-200 dark:bg-slate-700 rounded-full h-1">
									<div
										class="h-1 rounded-full bg-purple-500 transition-all"
										style="width: {systemHealth.nodes}%"
									></div>
								</div>
								<span class="{getHealthColor(systemHealth.nodes)} font-medium">{systemHealth.nodes}%</span>
							</div>
						</div>
						<div class="flex justify-between items-center">
							<div class="flex items-center space-x-2">
								<Icon icon="tabler:device-floppy" width={14} class="text-slate-500" />
								<span class="text-slate-600 dark:text-slate-400">存储空间</span>
							</div>
							<div class="flex items-center space-x-2">
								<div class="w-12 bg-slate-200 dark:bg-slate-700 rounded-full h-1">
									<div
										class="h-1 rounded-full bg-orange-500 transition-all"
										style="width: {systemHealth.storage}%"
									></div>
								</div>
								<span class="{getHealthColor(systemHealth.storage)} font-medium">{systemHealth.storage}%</span>
							</div>
						</div>
					</div>
				</CardContent>
			</Card>

			<!-- 快速操作 -->
			<Card class="bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm">
				<CardHeader class="pb-4">
					<div class="flex items-center space-x-2">
						<div class="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center">
							<Icon icon="tabler:bolt" width={18} class="text-white" />
						</div>
						<CardTitle class="text-lg text-slate-900 dark:text-slate-100">快速操作</CardTitle>
					</div>
				</CardHeader>
				<CardContent class="space-y-3">
					<Button
						variant="ghost"
						onclick={() => goto('/tasks/create')}
						class="w-full justify-start text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 hover:bg-blue-50 dark:hover:bg-blue-900/20"
					>
						<Icon icon="tabler:target" width={16} class="mr-3" />
						创建扫描任务
					</Button>
					<Button
						variant="ghost"
						onclick={() => goto('/projects/create')}
						class="w-full justify-start text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300 hover:bg-purple-50 dark:hover:bg-purple-900/20"
					>
						<Icon icon="tabler:folder" width={16} class="mr-3" />
						新建项目
					</Button>
					<Button
						variant="ghost"
						onclick={() => goto('/assets')}
						class="w-full justify-start text-green-600 hover:text-green-700 dark:text-green-400 dark:hover:text-green-300 hover:bg-green-50 dark:hover:bg-green-900/20"
					>
						<Icon icon="tabler:trending-up" width={16} class="mr-3" />
						导入资产
					</Button>
					<Button
						variant="ghost"
						onclick={() => goto('/nodes/create')}
						class="w-full justify-start text-orange-600 hover:text-orange-700 dark:text-orange-400 dark:hover:text-orange-300 hover:bg-orange-50 dark:hover:bg-orange-900/20"
					>
						<Icon icon="tabler:device-desktop" width={16} class="mr-3" />
						添加节点
					</Button>
					<Button
						variant="ghost"
						onclick={() => goto('/settings')}
						class="w-full justify-start text-slate-600 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800"
					>
						<Icon icon="tabler:settings" width={16} class="mr-3" />
						系统设置
					</Button>
				</CardContent>
			</Card>
		</div>

		<!-- 中栏：最近任务 -->
		<Card class="bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm">
			<CardHeader class="pb-4">
				<div class="flex items-center justify-between">
					<div class="flex items-center space-x-2">
						<div class="w-8 h-8 bg-gradient-to-br from-orange-500 to-orange-600 rounded-lg flex items-center justify-center">
							<Icon icon="tabler:file-text" width={18} class="text-white" />
						</div>
						<CardTitle class="text-lg text-slate-900 dark:text-slate-100">最近任务</CardTitle>
					</div>
					<Button
						variant="ghost"
						size="sm"
						onclick={() => goto('/tasks')}
						class="text-orange-600 hover:text-orange-700 dark:text-orange-400 dark:hover:text-orange-300"
					>
						查看全部
						<Icon icon="tabler:arrow-right" width={14} class="ml-1" />
					</Button>
				</div>
			</CardHeader>
			<CardContent>
				{#if loading}
					<div class="flex items-center justify-center py-12">
						<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-orange-500"></div>
					</div>
				{:else if recentTasks.length === 0}
					<div class="text-center py-12 text-slate-500 dark:text-slate-400">
						<Icon icon="tabler:file-text" width={48} class="mx-auto mb-4 text-slate-300 dark:text-slate-600" />
						<p class="text-lg font-medium mb-2">暂无任务</p>
						<p class="text-sm mb-4">创建您的第一个扫描任务</p>
						<Button onclick={() => goto('/tasks/create')} class="bg-gradient-to-r from-orange-500 to-orange-600">
							<Icon icon="tabler:plus" width={16} class="mr-2" />
							创建任务
						</Button>
					</div>
				{:else}
					<div class="space-y-3">
						{#each recentTasks as task}
							<div class="flex items-center justify-between p-4 border border-slate-200 dark:border-slate-700 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-all duration-200">
								<div class="flex-1">
									<div class="flex items-center space-x-2 mb-2">
										<span class="font-medium text-slate-800 dark:text-slate-200">{task.name}</span>
										<div class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium 
											{task.status === 'running' ? 'bg-green-100 text-green-700 ring-1 ring-green-600/20 dark:bg-green-900/20 dark:text-green-300' : 
											 task.status === 'completed' ? 'bg-blue-100 text-blue-700 ring-1 ring-blue-600/20 dark:bg-blue-900/20 dark:text-blue-300' : 
											 task.status === 'failed' ? 'bg-red-100 text-red-700 ring-1 ring-red-600/20 dark:bg-red-900/20 dark:text-red-300' : 
											 'bg-gray-100 text-gray-700 ring-1 ring-gray-600/20 dark:bg-gray-900/20 dark:text-gray-300'}">
											{#if task.status === 'running'}
												<Icon icon="tabler:player-play" width={10} />
											{:else if task.status === 'completed'}
												<Icon icon="tabler:circle-check" width={10} />
											{:else if task.status === 'failed'}
												<Icon icon="tabler:circle-x" width={10} />
											{:else}
												<Icon icon="tabler:player-pause" width={10} />
											{/if}
											{getTaskStatusText(task.status)}
										</div>
									</div>
									<div class="flex items-center space-x-2 text-xs text-slate-500 dark:text-slate-400">
										<span>{getTaskTypeText(task.type)}</span>
										<span>•</span>
										<div class="flex items-center space-x-1">
											<Icon icon="tabler:clock" width={12} />
											<span>{formatDateTime(task.createdAt)}</span>
										</div>
									</div>
									{#if task.status === 'running' && task.progress}
										<div class="mt-2">
											<div class="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-1">
												<div
													class="h-1 rounded-full bg-gradient-to-r from-orange-500 to-orange-600 transition-all duration-300"
													style="width: {task.progress}%"
												></div>
											</div>
										</div>
									{/if}
								</div>
								<Button
									variant="ghost"
									size="sm"
									onclick={() => goto(`/tasks/${task.id}`)}
									class="text-orange-600 hover:text-orange-700 dark:text-orange-400 dark:hover:text-orange-300 ml-4"
								>
									查看
									<Icon icon="tabler:arrow-right" width={14} class="ml-1" />
								</Button>
							</div>
						{/each}
					</div>
				{/if}
			</CardContent>
		</Card>

		<!-- 右栏：活跃项目 + 最近资产 -->
		<div class="space-y-6">
			<!-- 活跃项目 -->
			<Card class="bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm">
				<CardHeader class="pb-4">
					<div class="flex items-center justify-between">
						<div class="flex items-center space-x-2">
							<div class="w-8 h-8 bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg flex items-center justify-center">
								<Icon icon="tabler:folder" width={18} class="text-white" />
							</div>
							<CardTitle class="text-lg text-slate-900 dark:text-slate-100">活跃项目</CardTitle>
						</div>
						<Button
							variant="ghost"
							size="sm"
							onclick={() => goto('/projects')}
							class="text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300"
						>
							查看全部
							<Icon icon="tabler:arrow-right" width={14} class="ml-1" />
						</Button>
					</div>
				</CardHeader>
				<CardContent>
					{#if activeProjects.length === 0}
						<div class="text-center py-8 text-slate-500 dark:text-slate-400">
							<Icon icon="tabler:folder" width={48} class="mx-auto mb-4 text-slate-300 dark:text-slate-600" />
							<p class="text-sm font-medium mb-2">暂无项目</p>
							<Button
								variant="ghost"
								onclick={() => goto('/projects/create')}
								class="text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300"
							>
								<Icon icon="tabler:plus" width={14} class="mr-1" />
								创建项目
							</Button>
						</div>
					{:else}
						<div class="space-y-3">
							{#each activeProjects as project}
								<div class="flex items-center justify-between p-3 border border-slate-200 dark:border-slate-700 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-all duration-200">
									<div class="flex-1">
										<div class="font-medium text-slate-800 dark:text-slate-200 mb-1">{project.name}</div>
										<div class="flex items-center space-x-1 text-xs text-slate-500 dark:text-slate-400">
											<Icon icon="tabler:clock" width={12} />
											<span>{formatDateTime(project.created_at)}</span>
										</div>
									</div>
									<Button
										variant="ghost"
										size="sm"
										onclick={() => goto(`/projects/${project.id}`)}
										class="text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300"
									>
										进入
										<Icon icon="tabler:arrow-right" width={14} class="ml-1" />
									</Button>
								</div>
							{/each}
						</div>
					{/if}
				</CardContent>
			</Card>

			<!-- 最近资产 -->
			<Card class="bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm">
				<CardHeader class="pb-4">
					<div class="flex items-center justify-between">
						<div class="flex items-center space-x-2">
							<div class="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center">
								<Icon icon="tabler:target" width={18} class="text-white" />
							</div>
							<CardTitle class="text-lg text-slate-900 dark:text-slate-100">最近资产</CardTitle>
						</div>
						<Button
							variant="ghost"
							size="sm"
							onclick={() => goto('/assets')}
							class="text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
						>
							查看全部
							<Icon icon="tabler:arrow-right" width={14} class="ml-1" />
						</Button>
					</div>
				</CardHeader>
				<CardContent>
					{#if recentAssets.length === 0}
						<div class="text-center py-8 text-slate-500 dark:text-slate-400">
							<Icon icon="tabler:target" width={48} class="mx-auto mb-4 text-slate-300 dark:text-slate-600" />
							<p class="text-sm font-medium mb-2">暂无资产</p>
							<Button
								variant="ghost"
								onclick={() => goto('/assets')}
								class="text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
							>
								<Icon icon="tabler:plus" width={14} class="mr-1" />
								添加资产
							</Button>
						</div>
					{:else}
						<div class="space-y-2">
							{#each recentAssets as asset}
								<div class="flex items-center justify-between p-3 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-all duration-200">
									<div class="flex-1">
										<div class="font-medium text-slate-800 dark:text-slate-200 text-sm mb-1">{asset.value}</div>
										<div class="text-xs text-slate-500 dark:text-slate-400 uppercase tracking-wide">{asset.type}</div>
									</div>
									<Button
										variant="ghost"
										size="sm"
										onclick={() => goto(`/assets/${asset.id}`)}
										class="text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
									>
										查看
										<Icon icon="tabler:arrow-right" width={14} class="ml-1" />
									</Button>
								</div>
							{/each}
						</div>
					{/if}
				</CardContent>
			</Card>
		</div>
	</div>
</div>

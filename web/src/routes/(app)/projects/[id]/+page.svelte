<script lang="ts">
	import { goto } from '$app/navigation';
	import { ProjectAPI } from '$lib/api/projects';
	import type { Project } from '$lib/types/project';
	import { notifications } from '$lib/stores/notifications';
	import { Button } from '$lib/components/ui/button';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import Icon from '$lib/components/ui/Icon.svelte';

	// 从页面数据获取初始数据
	let { data } = $props();

	// 响应式状态
	let project = $state(data.project);
	let members = $state(data.members);
	let activities = $state(data.activities);
	let loading = $state(false);

	// 获取扫描状态颜色
	const getScanStatusColor = (status?: string) => {
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

	// 格式化日期
	const formatDate = (dateString: string) => {
		return new Date(dateString).toLocaleDateString('zh-CN', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	};

	// 开始扫描
	const handleStartScan = async () => {
		loading = true;
		try {
			// TODO: 实现开始扫描API
			notifications.add({
				type: 'info',
				message: '扫描功能正在开发中'
			});
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '启动扫描失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		} finally {
			loading = false;
		}
	};

	// 暂停扫描
	const handlePauseScan = async () => {
		loading = true;
		try {
			// TODO: 实现暂停扫描API
			notifications.add({
				type: 'info',
				message: '暂停扫描功能正在开发中'
			});
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '暂停扫描失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		} finally {
			loading = false;
		}
	};

	// 导出项目数据
	const handleExport = async (format: 'json' | 'csv' | 'xlsx' = 'json') => {
		try {
			const downloadUrl = await ProjectAPI.exportProject(project.id, format);
			const link = document.createElement('a');
			link.href = downloadUrl;
			link.download = `${project.name}_export.${format}`;
			document.body.appendChild(link);
			link.click();
			document.body.removeChild(link);
			URL.revokeObjectURL(downloadUrl);

			notifications.add({
				type: 'success',
				message: `项目数据导出成功 (${format.toUpperCase()})`
			});
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '导出失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		}
	};

	// 删除项目
	const handleDelete = async () => {
		if (!confirm(`确定要删除项目 "${project.name}" 吗？此操作不可逆。`)) return;

		try {
			await ProjectAPI.deleteProject(project.id);
			notifications.add({
				type: 'success',
				message: '项目删除成功'
			});
			await goto('/projects');
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '删除项目失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		}
	};
</script>

<svelte:head>
	<title>{project.name} - 项目详情 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6">
	<!-- 页面标题和操作 -->
	<div class="mb-6">
		<div class="flex items-center gap-4 mb-4">
			<Button variant="ghost" href="/projects" class="flex items-center gap-2">
				<ArrowLeft class="h-4 w-4" />
				返回项目列表
			</Button>
		</div>

		<div class="flex items-start justify-between">
			<div class="flex-1">
				<div class="flex items-center gap-3 mb-2">
					<h1 class="text-3xl font-bold text-gray-900">{project.name}</h1>
					{#if project.scan_status}
						<Badge class={getScanStatusColor(project.scan_status)}>
							{project.scan_status}
						</Badge>
					{/if}
					{#if project.is_private}
						<Badge variant="outline">私有</Badge>
					{/if}
				</div>

				{#if project.description}
					<p class="text-gray-600">{project.description}</p>
				{/if}

				<div class="flex items-center gap-4 mt-3 text-sm text-gray-500">
					<span>创建于 {formatDate(project.created_at)}</span>
					<span>最后更新 {formatDate(project.updated_at)}</span>
				</div>
			</div>

			<!-- 操作按钮 -->
			<div class="flex items-center gap-2">
				{#if project.scan_status === 'running'}
					<Button
						variant="outline"
						onclick={handlePauseScan}
						disabled={loading}
						class="flex items-center gap-2"
					>
						<Icon name="pause" class="h-4 w-4" />
						暂停扫描
					</Button>
				{:else}
					<Button onclick={handleStartScan} disabled={loading} class="flex items-center gap-2">
						<Icon name="play" class="h-4 w-4" />
						开始扫描
					</Button>
				{/if}

				<Button
					variant="outline"
					href="/projects/{project.id}/edit"
					class="flex items-center gap-2"
				>
					<Edit class="h-4 w-4" />
					编辑
				</Button>

				<div class="relative">
					<Button variant="outline" class="flex items-center gap-2">
						<Icon name="more-horizontal" class="h-4 w-4" />
						更多
					</Button>
					<!-- TODO: 实现下拉菜单 -->
				</div>
			</div>
		</div>
	</div>

	<!-- 项目统计卡片 -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">总资产数</CardTitle>
				<Icon name="folder" class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				<div class="text-2xl font-bold">{project.assets_count || 0}</div>
				<p class="text-xs text-muted-foreground">已发现的资产</p>
			</CardContent>
		</Card>

		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">发现漏洞</CardTitle>
				<Icon name="shield" class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				<div class="text-2xl font-bold text-red-600">{project.vulnerabilities_count || 0}</div>
				<p class="text-xs text-muted-foreground">需要关注的漏洞</p>
			</CardContent>
		</Card>

		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">运行任务</CardTitle>
				<Icon name="activity" class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				<div class="text-2xl font-bold">{project.tasks_count || 0}</div>
				<p class="text-xs text-muted-foreground">正在执行的任务</p>
			</CardContent>
		</Card>

		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">项目成员</CardTitle>
				<Icon name="user" class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				<div class="text-2xl font-bold">{members.length}</div>
				<p class="text-xs text-muted-foreground">参与项目的人员</p>
			</CardContent>
		</Card>
	</div>

	<!-- 主要内容区域 -->
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- 左侧主要内容 -->
		<div class="lg:col-span-2 space-y-6">
			<!-- 项目信息 -->
			<Card>
				<CardHeader>
					<CardTitle>项目信息</CardTitle>
					<CardDescription>项目的基本配置和设置</CardDescription>
				</CardHeader>
				<CardContent class="space-y-4">
					{#if project.target}
						<div>
							<div class="text-sm font-medium text-gray-500">扫描目标</div>
							<p class="mt-1 font-mono text-blue-600">{project.target}</p>
						</div>
					{/if}

					<div>
						<div class="text-sm font-medium text-gray-500">项目颜色</div>
						<div class="mt-1 flex items-center gap-2">
							<div class="w-4 h-4 rounded-full bg-{project.color || 'blue'}-500"></div>
							<span class="capitalize">{project.color || 'blue'}</span>
						</div>
					</div>

					<div>
						<div class="text-sm font-medium text-gray-500">可见性</div>
						<p class="mt-1">{project.is_private ? '私有项目' : '公开项目'}</p>
					</div>

					<div>
						<div class="text-sm font-medium text-gray-500">创建者</div>
						<p class="mt-1">{project.created_by || '未知'}</p>
					</div>
				</CardContent>
			</Card>

			<!-- 快速操作 -->
			<Card>
				<CardHeader>
					<CardTitle>快速操作</CardTitle>
					<CardDescription>常用的项目管理操作</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
						<Button
							variant="outline"
							href="/projects/{project.id}/assets"
							class="flex items-center gap-2 h-auto py-4 flex-col"
						>
							<Icon name="folder" class="h-6 w-6" />
							<span class="text-sm">资产管理</span>
						</Button>

						<Button
							variant="outline"
							href="/projects/{project.id}/vulnerabilities"
							class="flex items-center gap-2 h-auto py-4 flex-col"
						>
							<Icon name="shield" class="h-6 w-6" />
							<span class="text-sm">漏洞报告</span>
						</Button>

						<Button
							variant="outline"
							href="/projects/{project.id}/tasks"
							class="flex items-center gap-2 h-auto py-4 flex-col"
						>
							<Icon name="activity" class="h-6 w-6" />
							<span class="text-sm">任务管理</span>
						</Button>

						<Button
							variant="outline"
							onclick={() => handleExport('json')}
							class="flex items-center gap-2 h-auto py-4 flex-col"
						>
							<Icon name="download" class="h-6 w-6" />
							<span class="text-sm">导出数据</span>
						</Button>
					</div>
				</CardContent>
			</Card>
		</div>

		<!-- 右侧边栏 -->
		<div class="space-y-6">
			<!-- 项目成员 -->
			<Card>
				<CardHeader>
					<CardTitle class="flex items-center justify-between">
						项目成员
						<Button variant="outline" size="sm" href="/projects/{project.id}/members">管理</Button>
					</CardTitle>
				</CardHeader>
				<CardContent>
					{#if members.length === 0}
						<p class="text-gray-500 text-sm">暂无成员</p>
					{:else}
						<div class="space-y-3">
							{#each members.slice(0, 5) as member}
								<div class="flex items-center gap-3">
									<div
										class="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center text-white text-sm font-medium"
									>
										{member.name?.charAt(0) || 'U'}
									</div>
									<div class="flex-1">
										<p class="text-sm font-medium">{member.name || '未知用户'}</p>
										<p class="text-xs text-gray-500">{member.role || 'member'}</p>
									</div>
								</div>
							{/each}

							{#if members.length > 5}
								<p class="text-xs text-gray-500">还有 {members.length - 5} 个成员...</p>
							{/if}
						</div>
					{/if}
				</CardContent>
			</Card>

			<!-- 最近活动 -->
			<Card>
				<CardHeader>
					<CardTitle>最近活动</CardTitle>
				</CardHeader>
				<CardContent>
					{#if activities.data && activities.data.length === 0}
						<p class="text-gray-500 text-sm">暂无活动记录</p>
					{:else if activities.data}
						<div class="space-y-3">
							{#each activities.data.slice(0, 5) as activity}
								<div class="flex gap-3">
									<div class="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></div>
									<div class="flex-1">
										<p class="text-sm">{activity.description || activity.action}</p>
										<p class="text-xs text-gray-500">{formatDate(activity.created_at)}</p>
									</div>
								</div>
							{/each}

							{#if activities.data.length > 5}
								<Button
									variant="outline"
									size="sm"
									href="/projects/{project.id}/activities"
									class="w-full"
								>
									查看更多活动
								</Button>
							{/if}
						</div>
					{/if}
				</CardContent>
			</Card>

			<!-- 危险操作 -->
			<Card>
				<CardHeader>
					<CardTitle class="text-red-600">危险操作</CardTitle>
					<CardDescription>这些操作不可逆，请谨慎使用</CardDescription>
				</CardHeader>
				<CardContent>
					<Button
						variant="outline"
						onclick={handleDelete}
						class="w-full text-red-600 hover:text-red-800 border-red-200 hover:border-red-300"
					>
						删除项目
					</Button>
				</CardContent>
			</Card>
		</div>
	</div>
</div>

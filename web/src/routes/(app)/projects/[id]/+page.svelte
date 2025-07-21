<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import { ProjectAPI } from '$lib/api/projects';
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
	import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
	import Icon from '@iconify/svelte';
	import { onMount } from 'svelte';

	let { data }: { data: PageData } = $props();
	let project = $state(data.project);
	let members = $state(data.members);
	let activities = $state(data.activities);
	let loading = $state(false);

	// 格式化日期
	const formatDate = (dateString: string) => {
		return new Date(dateString).toLocaleDateString('zh-CN', {
			year: 'numeric',
			month: 'long',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	};

	// 获取状态颜色
	const getStatusColor = (status?: string) => {
		switch (status) {
			case 'active':
				return 'bg-green-100 text-green-800';
			case 'paused':
				return 'bg-yellow-100 text-yellow-800';
			case 'completed':
				return 'bg-blue-100 text-blue-800';
			case 'failed':
				return 'bg-red-100 text-red-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	};

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

	// 删除项目
	const handleDelete = async () => {
		if (!confirm(`确定要删除项目 "${project.name}" 吗？此操作不可恢复。`)) return;

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

	// 更新项目状态
	const updateProjectStatus = async (newStatus: string) => {
		try {
			const updated = await ProjectAPI.updateProject(project.id, { status: newStatus });
			project = updated;
			notifications.add({
				type: 'success',
				message: '项目状态更新成功'
			});
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '更新项目状态失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		}
	};

	// 复制项目
	const duplicateProject = async () => {
		const newName = prompt('请输入新项目名称:', `${project.name} - 副本`);
		if (!newName) return;

		try {
			const newProject = await ProjectAPI.duplicateProject(project.id, newName);
			notifications.add({
				type: 'success',
				message: '项目复制成功'
			});
			await goto(`/projects/${newProject.id}`);
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '复制项目失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		}
	};
</script>

<svelte:head>
	<title>{project.name} - 项目详情</title>
</svelte:head>

<div class="container mx-auto px-4 py-6 max-w-7xl">
	<!-- 页面头部 -->
	<div class="mb-6">
		<div class="flex items-center justify-between">
			<div>
				<div class="flex items-center gap-3 mb-2">
					<Icon icon="tabler:folder" class="h-8 w-8 text-blue-600" />
					<h1 class="text-3xl font-bold text-gray-900">{project.name}</h1>
					{#if project.is_private}
						<Badge variant="outline" class="bg-yellow-50 text-yellow-800">
							<Icon icon="tabler:lock" class="h-3 w-3 mr-1" />
							私有
						</Badge>
					{/if}
				</div>
				{#if project.description}
					<p class="text-gray-600 max-w-2xl">{project.description}</p>
				{/if}
			</div>

			<div class="flex items-center gap-2">
				<Button variant="outline" onclick={duplicateProject}>
					<Icon icon="tabler:copy" class="h-4 w-4 mr-1" />
					复制
				</Button>
				<Button variant="outline" href="/projects/{project.id}/edit">
					<Icon icon="tabler:edit" class="h-4 w-4 mr-1" />
					编辑
				</Button>
				<Button variant="destructive" onclick={handleDelete}>
					<Icon icon="tabler:trash" class="h-4 w-4 mr-1" />
					删除
				</Button>
			</div>
		</div>
	</div>

	<!-- 项目统计卡片 -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
		<Card>
			<CardHeader class="pb-3">
				<CardDescription class="flex items-center gap-2">
					<Icon icon="tabler:target" class="h-4 w-4" />
					目标
				</CardDescription>
				<CardTitle class="text-2xl">{project.target || 'N/A'}</CardTitle>
			</CardHeader>
		</Card>

		<Card>
			<CardHeader class="pb-3">
				<CardDescription class="flex items-center gap-2">
					<Icon icon="tabler:server" class="h-4 w-4" />
					资产数量
				</CardDescription>
				<CardTitle class="text-2xl">{project.assets_count || 0}</CardTitle>
			</CardHeader>
		</Card>

		<Card>
			<CardHeader class="pb-3">
				<CardDescription class="flex items-center gap-2">
					<Icon icon="tabler:bug" class="h-4 w-4" />
					漏洞数量
				</CardDescription>
				<CardTitle class="text-2xl text-red-600">{project.vulnerabilities_count || 0}</CardTitle>
			</CardHeader>
		</Card>

		<Card>
			<CardHeader class="pb-3">
				<CardDescription class="flex items-center gap-2">
					<Icon icon="tabler:activity" class="h-4 w-4" />
					扫描状态
				</CardDescription>
				<CardTitle>
					<Badge class={getScanStatusColor(project.scan_status)}>
						{project.scan_status || '未开始'}
					</Badge>
				</CardTitle>
			</CardHeader>
		</Card>
	</div>

	<!-- 内容区域 -->
	<Tabs value="overview" class="w-full">
		<TabsList class="grid w-full grid-cols-4">
			<TabsTrigger value="overview">项目概览</TabsTrigger>
			<TabsTrigger value="members">项目成员</TabsTrigger>
			<TabsTrigger value="assets">项目资产</TabsTrigger>
			<TabsTrigger value="activities">项目活动</TabsTrigger>
		</TabsList>

		<TabsContent value="overview" class="space-y-6">
			<!-- 项目信息 -->
			<Card>
				<CardHeader>
					<CardTitle class="text-lg">项目信息</CardTitle>
				</CardHeader>
				<CardContent class="space-y-4">
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<span class="text-sm font-medium text-gray-500">创建时间</span>
							<p>{formatDate(project.created_at)}</p>
						</div>
						<div>
							<span class="text-sm font-medium text-gray-500">更新时间</span>
							<p>{formatDate(project.updated_at)}</p>
						</div>
						<div>
							<span class="text-sm font-medium text-gray-500">项目状态</span>
							<Badge class={getStatusColor(project.status)}>
								{project.status || 'active'}
							</Badge>
						</div>
						<div>
							<span class="text-sm font-medium text-gray-500">扫描类型</span>
							<p>{project.scan_type || '未指定'}</p>
						</div>
					</div>
					{#if project.tags?.length > 0}
						<div>
							<label class="text-sm font-medium text-gray-500">标签</label>
							<div class="flex gap-2 mt-1">
								{#each project.tags as tag}
									<Badge variant="secondary">{tag}</Badge>
								{/each}
							</div>
						</div>
					{/if}
				</CardContent>
			</Card>

			<!-- 项目描述 -->
			{#if project.description}
				<Card>
					<CardHeader>
						<CardTitle class="text-lg">项目描述</CardTitle>
					</CardHeader>
					<CardContent>
						<p class="text-gray-700">{project.description}</p>
					</CardContent>
				</Card>
			{/if}
		</TabsContent>

		<TabsContent value="members" class="space-y-6">
			<!-- 项目成员 -->
			<Card>
				<CardHeader>
					<CardTitle class="text-lg">项目成员</CardTitle>
					<CardDescription>管理项目成员和权限</CardDescription>
				</CardHeader>
				<CardContent>
					{#if members.length > 0}
						<div class="space-y-3">
							{#each members as member}
								<div class="flex items-center justify-between p-3 border rounded-lg">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
											<Icon icon="tabler:user" class="h-4 w-4 text-blue-600" />
										</div>
										<div>
											<p class="font-medium">{member.username}</p>
											<p class="text-sm text-gray-500">{member.role}</p>
										</div>
									</div>
									<Badge variant="outline">{member.permission}</Badge>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-gray-500 text-center py-4">暂无项目成员</p>
					{/if}
				</CardContent>
			</Card>
		</TabsContent>

		<TabsContent value="assets" class="space-y-6">
			<!-- 项目资产 -->
			<Card>
				<CardHeader>
					<CardTitle class="text-lg">项目资产</CardTitle>
					<CardDescription>项目相关的所有资产列表</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="text-center py-12">
						<Icon icon="tabler:server" class="h-12 w-12 text-gray-400 mx-auto mb-4" />
						<p class="text-gray-500">资产功能正在开发中</p>
						<Button href="/assets" class="mt-4">前往资产管理</Button>
					</div>
				</CardContent>
			</Card>
		</TabsContent>

		<TabsContent value="activities" class="space-y-6">
			<!-- 项目活动 -->
			<Card>
				<CardHeader>
					<CardTitle class="text-lg">项目活动</CardTitle>
					<CardDescription>项目相关的操作记录</CardDescription>
				</CardHeader>
				<CardContent>
					{#if activities.data.length > 0}
						<div class="space-y-3">
							{#each activities.data as activity}
								<div class="flex items-start gap-3 p-3 border rounded-lg">
									<div class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0">
										<Icon icon="tabler:activity" class="h-4 w-4 text-blue-600" />
									</div>
									<div class="flex-1">
										<p class="font-medium">{activity.action}</p>
										<p class="text-sm text-gray-500">{activity.description}</p>
										<p class="text-xs text-gray-400">{formatDate(activity.created_at)}</p>
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-gray-500 text-center py-4">暂无活动记录</p>
					{/if}
				</CardContent>
			</Card>
		</TabsContent>
	</Tabs>
</div>
<script lang="ts">
	import type { Project } from '$lib/types/project';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';

	import { Button } from '$lib/components/ui/button';
	import { MoreHorizontal, Target, Shield, Activity } from 'lucide-svelte';

	interface Props {
		project: Project;
		onEdit?: (project: Project) => void;
		onDelete?: (project: Project) => void;
		onDuplicate?: (project: Project) => void;
		onExport?: (project: Project) => void;
	}

	let { project, onEdit, onDelete, onDuplicate, onExport }: Props = $props();

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
		return new Date(dateString).toLocaleDateString('zh-CN');
	};

	// 获取项目颜色类
	const getProjectColorClass = (color?: string) => {
		const colorMap: Record<string, string> = {
			blue: 'bg-blue-500',
			green: 'bg-green-500',
			red: 'bg-red-500',
			yellow: 'bg-yellow-500',
			purple: 'bg-purple-500',
			pink: 'bg-pink-500',
			indigo: 'bg-indigo-500',
			gray: 'bg-gray-500'
		};
		return colorMap[color || 'blue'] || 'bg-blue-500';
	};

	// 显示更多菜单
	let showMenu = $state(false);
</script>

<Card class="hover:shadow-lg transition-shadow duration-200 group">
	<CardHeader>
		<div class="flex items-start justify-between">
			<div class="flex-1">
				<!-- 项目标题和颜色指示器 -->
				<div class="flex items-center gap-3 mb-2">
					<div class="w-3 h-3 rounded-full {getProjectColorClass(project.color)}"></div>
					<CardTitle class="text-lg group-hover:text-blue-600 transition-colors">
						<a href="/projects/{project.id}" class="block">
							{project.name}
						</a>
					</CardTitle>
				</div>

				<!-- 项目描述 -->
				{#if project.description}
					<CardDescription class="line-clamp-2 mb-3">
						{project.description}
					</CardDescription>
				{/if}

				<!-- 目标地址 -->
				{#if project.target}
					<div class="flex items-center gap-2 mb-2 text-sm text-gray-600">
						<Target class="h-4 w-4" />
						<span class="font-mono text-blue-600">{project.target}</span>
					</div>
				{/if}
			</div>

			<!-- 状态和操作 -->
			<div class="flex items-center gap-2">
				{#if project.scan_status}
					<span class={`px-2 py-1 text-xs rounded-full ${getScanStatusColor(project.scan_status)}`}>
						{project.scan_status}
					</span>
				{/if}

				{#if project.is_private}
					<span class="px-2 py-1 text-xs rounded-full border border-gray-300 text-gray-600"
						>私有</span
					>
				{/if}

				<!-- 更多操作按钮 -->
				<div class="relative">
					<Button
						variant="ghost"
						size="sm"
						class="h-8 w-8 p-0 opacity-0 group-hover:opacity-100 transition-opacity"
						onclick={() => (showMenu = !showMenu)}
					>
						<MoreHorizontal class="h-4 w-4" />
					</Button>

					<!-- 下拉菜单 -->
					{#if showMenu}
						<div
							class="absolute right-0 top-full mt-1 w-48 bg-white rounded-md shadow-lg border z-10"
						>
							<div class="py-1">
								{#if onEdit}
									<button
										onclick={() => {
											onEdit?.(project);
											showMenu = false;
										}}
										class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
									>
										编辑项目
									</button>
								{/if}

								{#if onDuplicate}
									<button
										onclick={() => {
											onDuplicate?.(project);
											showMenu = false;
										}}
										class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
									>
										复制项目
									</button>
								{/if}

								{#if onExport}
									<button
										onclick={() => {
											onExport?.(project);
											showMenu = false;
										}}
										class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
									>
										导出数据
									</button>
								{/if}

								<hr class="my-1" />

								{#if onDelete}
									<button
										onclick={() => {
											onDelete?.(project);
											showMenu = false;
										}}
										class="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50"
									>
										删除项目
									</button>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</CardHeader>

	<CardContent>
		<div class="space-y-4">
			<!-- 项目统计 -->
			<div class="grid grid-cols-3 gap-4 text-center">
				<div class="flex flex-col items-center">
					<Target class="h-5 w-5 text-gray-400 mb-1" />
					<div class="text-lg font-semibold">{project.assets_count || 0}</div>
					<div class="text-xs text-gray-500">资产</div>
				</div>

				<div class="flex flex-col items-center">
					<Shield class="h-5 w-5 text-red-400 mb-1" />
					<div class="text-lg font-semibold text-red-600">{project.vulnerabilities_count || 0}</div>
					<div class="text-xs text-gray-500">漏洞</div>
				</div>

				<div class="flex flex-col items-center">
					<Activity class="h-5 w-5 text-blue-400 mb-1" />
					<div class="text-lg font-semibold">{project.tasks_count || 0}</div>
					<div class="text-xs text-gray-500">任务</div>
				</div>
			</div>

			<!-- 项目元信息 -->
			<div class="flex items-center justify-between text-sm text-gray-500 border-t pt-3">
				<div>创建于 {formatDate(project.created_at)}</div>
				{#if project.created_by}
					<div>by {project.created_by}</div>
				{/if}
			</div>

			<!-- 操作按钮 -->
			<div class="flex items-center gap-2 pt-2">
				<Button size="sm" href="/projects/{project.id}" class="flex-1">查看详情</Button>

				{#if project.scan_status === 'running'}
					<Button variant="outline" size="sm" class="text-yellow-600">运行中</Button>
				{:else}
					<Button variant="outline" size="sm" class="text-green-600">启动扫描</Button>
				{/if}
			</div>
		</div>
	</CardContent>
</Card>

<!-- 点击外部关闭菜单 -->
{#if showMenu}
	<div class="fixed inset-0 z-0" onclick={() => (showMenu = false)}></div>
{/if}

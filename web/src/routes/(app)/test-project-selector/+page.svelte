<script lang="ts">
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import ProjectSelector from '$lib/components/ui/project-selector/ProjectSelector.svelte';
	import type { Project } from '$lib/types/project';

	// 状态
	let selectedProject = $state<Project | null>(null);
	let selectedProjectId = $state('');
	let selectedProjectName = $state('');

	// 处理项目选择
	const handleProjectSelect = (project: Project | null) => {
		selectedProject = project;
		selectedProjectId = project?.id || '';
		selectedProjectName = project?.name || '';
		console.log('选择的项目:', project);
	};

	// 重置选择
	const resetSelection = () => {
		selectedProject = null;
		selectedProjectId = '';
		selectedProjectName = '';
	};
</script>

<svelte:head>
	<title>项目选择器测试 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6 max-w-2xl">
	<Card>
		<CardHeader>
			<CardTitle>项目选择器功能测试</CardTitle>
		</CardHeader>
		<CardContent class="space-y-6">
			<!-- 项目选择器 -->
			<div>
				<ProjectSelector 
					bind:selectedProjectId={selectedProjectId}
					bind:selectedProjectName={selectedProjectName}
					placeholder="搜索项目名称、ID或标签..."
					onProjectSelect={handleProjectSelect}
				/>
			</div>

			<!-- 选择结果显示 -->
			{#if selectedProject}
				<div class="p-4 bg-green-50 border border-green-200 rounded-lg">
					<h3 class="font-semibold text-green-800 mb-2">选择结果</h3>
					<div class="space-y-1 text-sm">
						<p><strong>项目ID:</strong> {selectedProject.id}</p>
						<p><strong>项目名称:</strong> {selectedProject.name}</p>
						{#if selectedProject.description}
							<p><strong>描述:</strong> {selectedProject.description}</p>
						{/if}
						{#if selectedProject.tag}
							<p><strong>标签:</strong> {selectedProject.tag}</p>
						{/if}
						{#if selectedProject.created}
							<p><strong>创建时间:</strong> {new Date(selectedProject.created).toLocaleString()}</p>
						{/if}
					</div>
				</div>
			{:else}
				<div class="p-4 bg-gray-50 border border-gray-200 rounded-lg">
					<p class="text-gray-600 text-sm">未选择项目</p>
				</div>
			{/if}

			<!-- 操作按钮 -->
			<div class="flex gap-3">
				<Button onclick={resetSelection} variant="outline">
					重置选择
				</Button>
				<Button 
					onclick={() => console.log('当前选择:', { selectedProjectId, selectedProjectName, selectedProject })}
					variant="outline"
				>
					打印到控制台
				</Button>
			</div>

			<!-- 使用说明 -->
			<div class="p-4 bg-blue-50 border border-blue-200 rounded-lg">
				<h3 class="font-semibold text-blue-800 mb-2">功能说明</h3>
				<ul class="text-sm text-blue-700 space-y-1">
					<li>• 支持通过项目名称、ID或标签进行搜索</li>
					<li>• 搜索结果实时更新，支持模糊匹配</li>
					<li>• 显示项目完整信息：名称、ID、描述、标签</li>
					<li>• 支持清除选择功能</li>
					<li>• 选择后显示详细的项目信息卡片</li>
				</ul>
			</div>
		</CardContent>
	</Card>
</div>
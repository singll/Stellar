<script lang="ts">
	import SearchableProjectSelector from '$lib/components/ui/searchable-project-selector/SearchableProjectSelector.svelte';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import type { Project } from '$lib/types/project';

	let selectedProjectId = $state('');
	let selectedProjectName = $state('');
	let selectedProject: Project | null = $state(null);

	function handleProjectSelect(project: Project | null) {
		selectedProject = project;
		console.log('Selected project:', project);
	}
</script>

<svelte:head>
	<title>测试可搜索项目选择器 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6 max-w-4xl">
	<h1 class="text-3xl font-bold text-gray-900 mb-8">测试可搜索项目选择器</h1>

	<div class="grid gap-6">
		<Card>
			<CardHeader>
				<CardTitle>基础功能测试</CardTitle>
			</CardHeader>
			<CardContent class="space-y-6">
				<SearchableProjectSelector 
					bind:selectedProjectId
					bind:selectedProjectName
					placeholder="搜索项目名称、ID或标签..."
					onProjectSelect={handleProjectSelect}
					class="max-w-md"
				/>

				<div class="p-4 bg-gray-50 rounded-md">
					<h3 class="font-medium mb-2">当前选择状态：</h3>
					<pre class="text-sm">{JSON.stringify({
						selectedProjectId,
						selectedProjectName,
						selectedProject
					}, null, 2)}</pre>
				</div>
			</CardContent>
		</Card>

		<Card>
			<CardHeader>
				<CardTitle>必填字段测试</CardTitle>
			</CardHeader>
			<CardContent>
				<SearchableProjectSelector 
					placeholder="必填项目选择..."
					required={true}
					class="max-w-md"
				/>
			</CardContent>
		</Card>

		<Card>
			<CardHeader>
				<CardTitle>禁用状态测试</CardTitle>
			</CardHeader>
			<CardContent>
				<SearchableProjectSelector 
					placeholder="禁用状态..."
					disabled={true}
					class="max-w-md"
				/>
			</CardContent>
		</Card>
	</div>
</div>
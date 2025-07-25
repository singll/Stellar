<script lang="ts">
	import SearchableProjectSelector from '$lib/components/ui/searchable-project-selector/SearchableProjectSelector.svelte';
	import { Button } from '$lib/components/ui/button';
	import type { Project } from '$lib/types/project';

	let selectedProjectId = $state('');
	let selectedProjectName = $state('');

	function handleProjectSelect(project: Project | null) {
		console.log('项目选择回调:', project);
	}

	function resetSelection() {
		selectedProjectId = '';
		selectedProjectName = '';
	}
</script>

<svelte:head>
	<title>搜索模式测试 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6 max-w-2xl">
	<h1 class="text-2xl font-bold mb-6">搜索模式修复测试</h1>

	<div class="space-y-6">
		<div class="p-4 bg-blue-50 border border-blue-200 rounded-md">
			<h3 class="font-medium mb-2">测试场景：</h3>
			<ul class="text-sm text-blue-800 space-y-1">
				<li>1. 点击输入框 → 应该显示所有项目（不触发搜索过滤）</li>
				<li>2. 输入文字 → 开始搜索过滤</li>
				<li>3. 选择项目 → 退出搜索模式，显示项目名称</li>
				<li>4. 修改已选择项目的名称 → 重新进入搜索模式</li>
			</ul>
		</div>

		<SearchableProjectSelector 
			bind:selectedProjectId
			bind:selectedProjectName
			placeholder="点击或输入搜索项目..."
			onProjectSelect={handleProjectSelect}
		/>

		<div class="flex gap-2">
			<Button onclick={resetSelection} variant="outline">
				重置选择
			</Button>
		</div>

		<div class="p-4 bg-gray-50 border rounded-md">
			<h3 class="font-medium mb-2">当前状态：</h3>
			<div class="text-sm text-gray-600 space-y-1">
				<p>选中ID: <code class="bg-white px-1 rounded">{selectedProjectId || '无'}</code></p>
				<p>选中名称: <code class="bg-white px-1 rounded">{selectedProjectName || '无'}</code></p>
			</div>
		</div>
	</div>
</div>
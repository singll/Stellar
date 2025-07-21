<script lang="ts">
	import { onMount } from 'svelte';
	import { ProjectAPI } from '$lib/api/projects';
	
	let projects = [];
	let loading = true;
	let error = null;
	let rawData = null;
	
	onMount(async () => {
		try {
			console.log('🔍 开始测试API调用...');
			const response = await ProjectAPI.getProjects({ page: 1, limit: 10 });
			console.log('📦 API响应:', response);
			
			projects = response.data || [];
			rawData = response;
			
			console.log('✅ 数据验证:', {
				projectsLength: projects.length,
				firstProject: projects[0] || null,
				responseStructure: Object.keys(response)
			});
			
		} catch (err) {
			console.error('❌ API调用失败:', err);
			error = err.message;
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>数据调试测试 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6">
	<h1 class="text-3xl font-bold mb-6">API数据调试测试</h1>
	
	{#if loading}
		<div class="flex justify-center items-center h-64">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
			<span class="ml-2">加载中...</span>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
			<strong>错误: </strong> {error}
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
			<div class="bg-white border rounded-lg p-4">
				<h2 class="text-xl font-semibold mb-4">API响应</h2>
				<pre class="bg-gray-100 p-2 rounded text-xs overflow-auto max-h-64">{JSON.stringify(rawData, null, 2)}</pre>
			</div>
			
			<div class="bg-white border rounded-lg p-4">
				<h2 class="text-xl font-semibold mb-4">项目列表</h2>
				<div class="text-sm">
					项目数量: {projects.length}
					{#each projects as project (project.id)}
						<div class="mt-2 p-2 border rounded">
							<div class="font-semibold">{project.name}</div>
							<div class="text-gray-600 text-xs">{project.description || '无描述'}</div>
							<div class="text-gray-500 text-xs">ID: {project.id}</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{/if}
</div>
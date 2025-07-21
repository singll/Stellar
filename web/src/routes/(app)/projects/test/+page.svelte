<script lang="ts">
	import { onMount } from 'svelte';
	import { ProjectAPI } from '$lib/api/projects';
	import { notifications } from '$lib/stores/notifications';

	let projects = $state([]);
	let loading = $state(true);
	let error = $state(null);

	onMount(async () => {
		try {
			loading = true;
			const response = await ProjectAPI.getProjects({ page: 1, limit: 10 });
			projects = response.data;
			console.log('项目列表数据:', response);
		} catch (err) {
			error = err instanceof Error ? err.message : '未知错误';
			console.error('获取项目列表失败:', err);
			notifications.add({
				type: 'error',
				message: '获取项目列表失败: ' + error
			});
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>项目列表测试 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6">
	<h1 class="text-2xl font-bold mb-4">项目列表测试</h1>
	
	{#if loading}
		<div class="text-center py-8">加载中...</div>
	{:else if error}
		<div class="text-red-500 bg-red-50 p-4 rounded">
			错误: {error}
		</div>
	{:else}
		<div class="space-y-4">
			<p>找到 {projects.length} 个项目</p>
			{#each projects as project}
				<div class="border p-4 rounded">
					<h3 class="font-bold text-lg">{project.name}</h3>
					<p class="text-gray-600">{project.description || '无描述'}</p>
					<p class="text-sm text-gray-500">创建时间: {project.created_at}</p>
				</div>
			{/each}
		</div>
	{/if}
</div>
<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	
	let message = $state('页面加载中...');
	
	onMount(() => {
		if (browser) {
			message = '页面加载成功！';
		}
	});
</script>

<svelte:head>
	<title>简单测试 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-6 py-8">
	<h1 class="text-3xl font-bold mb-8">简单测试页面</h1>
	
	<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">
		<p><strong>状态:</strong> {message}</p>
		<p><strong>浏览器环境:</strong> {browser ? '是' : '否'}</p>
		<p><strong>时间:</strong> {new Date().toLocaleString()}</p>
	</div>
	
	<div class="mt-8">
		<a href="/login" class="text-blue-600 hover:text-blue-800 underline">
			前往登录页面
		</a>
	</div>
</div> 
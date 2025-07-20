<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';

	let authState = $state(auth.state);
	let sessionStatus: any = $state(null);
	let isInitialized = $state(false);

	onMount(async () => {
		if (browser) {
			await auth.initialize();
			isInitialized = true;
		}
	});

	async function checkSessionStatus() {
		if (authState.token) {
			sessionStatus = await auth.getSessionStatus();
		}
	}

	async function verifySession() {
		const result = await auth.verifySession();
		console.log('会话验证结果:', result);
	}

	function clearAuth() {
		auth.clearState();
	}

	async function refreshSession() {
		const result = await auth.refreshSession();
		console.log('会话刷新结果:', result);
	}
</script>

<svelte:head>
	<title>认证状态测试 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-6 py-8">
	<h1 class="text-3xl font-bold mb-8">认证状态测试页面</h1>

	<div class="grid md:grid-cols-2 gap-8">
		<!-- 认证状态 -->
		<Card>
			<CardHeader>
				<CardTitle>认证状态</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					<div>
						<strong>初始化状态:</strong> {isInitialized ? '✅ 已初始化' : '⏳ 初始化中...'}
					</div>
					<div>
						<strong>认证状态:</strong> {authState.isAuthenticated ? '✅ 已认证' : '❌ 未认证'}
					</div>
					<div>
						<strong>Token:</strong> {authState.token ? `✅ ${authState.token.substring(0, 20)}...` : '❌ 无Token'}
					</div>
					<div>
						<strong>用户:</strong> {authState.user ? `✅ ${authState.user.username}` : '❌ 无用户信息'}
					</div>
					<div>
						<strong>localStorage:</strong> 
						{#if browser}
							{localStorage.getItem('auth_state') ? '✅ 有数据' : '❌ 无数据'}
						{:else}
							⏳ 服务端
						{/if}
					</div>
				</div>
			</CardContent>
		</Card>

		<!-- 操作按钮 -->
		<Card>
			<CardHeader>
				<CardTitle>操作</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					<Button onclick={verifySession} disabled={!authState.token}>
						验证会话
					</Button>
					<Button onclick={checkSessionStatus} disabled={!authState.token}>
						获取会话状态
					</Button>
					<Button onclick={refreshSession} disabled={!authState.token}>
						刷新会话
					</Button>
					<Button onclick={clearAuth} variant="outline">
						清理认证状态
					</Button>
				</div>
			</CardContent>
		</Card>
	</div>

	<!-- 会话状态信息 -->
	{#if sessionStatus}
		<Card class="mt-8">
			<CardHeader>
				<CardTitle>会话状态详情</CardTitle>
			</CardHeader>
			<CardContent>
				<pre class="bg-gray-100 p-4 rounded text-sm overflow-auto">
{JSON.stringify(sessionStatus, null, 2)}
				</pre>
			</CardContent>
		</Card>
	{/if}

	<!-- localStorage 内容 -->
	{#if browser}
		<Card class="mt-8">
			<CardHeader>
				<CardTitle>localStorage 内容</CardTitle>
			</CardHeader>
			<CardContent>
				<pre class="bg-gray-100 p-4 rounded text-sm overflow-auto">
{localStorage.getItem('auth_state') || '无数据'}
				</pre>
			</CardContent>
		</Card>
	{/if}
</div> 
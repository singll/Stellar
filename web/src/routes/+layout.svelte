<script lang="ts">
	import '../app.css';
	import NotificationContainer from '$lib/components/notification/NotificationContainer.svelte';
	import type { Snippet } from 'svelte';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { auth } from '$lib/stores/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();
	let isCheckingAuth = $state(true);

	// 公开路由列表
	const publicRoutes = [
		'/login',
		'/register',
		'/forgot-password',
		'/reset-password',
		'/verify-email',
		'/test-auth',
		'/test-basic',
		'/test-minimal',
		'/test-simple'
	];

	// 检查是否是公开路由
	function isPublicRoute(path: string): boolean {
		return publicRoutes.some((route) => path.startsWith(route));
	}

	// 客户端认证检查
	async function checkAuth() {
		if (!browser) return;

		try {
			// 初始化认证状态
			await auth.initialize();
			
			const { isAuthenticated, token } = auth.state;
			const currentPath = $page.url.pathname;

			console.log('客户端认证检查:', { 
				currentPath, 
				isAuthenticated, 
				hasToken: !!token,
				isPublicRoute: isPublicRoute(currentPath) 
			});

			// 如果已认证且访问登录页，重定向到dashboard
			if (isAuthenticated && token && currentPath === '/login') {
				console.log('已认证用户访问登录页，重定向到dashboard');
				await goto('/dashboard');
				return;
			}

			// 如果未认证且不是公开路由，需要验证会话
			if (!isAuthenticated && !isPublicRoute(currentPath)) {
				console.log('未认证，尝试验证会话...');
				
				// 尝试从localStorage恢复状态并验证会话
				const storedState = localStorage.getItem('auth_state');
				if (storedState) {
					try {
						const parsedState = JSON.parse(storedState);
						if (parsedState.token && parsedState.user) {
							console.log('发现存储的认证状态，验证会话');
							const isValid = await auth.verifySession();
							if (isValid) {
								console.log('会话验证成功，继续访问');
								return;
							}
						}
					} catch (error) {
						console.error('解析存储的认证状态失败:', error);
					}
				}

				console.log('会话无效，重定向到登录页');
				// 会话无效，重定向到登录页
				const redirectUrl = encodeURIComponent(currentPath);
				await goto(`/login?redirect=${redirectUrl}`);
			}
		} finally {
			isCheckingAuth = false;
		}
	}

	onMount(() => {
		checkAuth();
	});
</script>

{#if isCheckingAuth}
	<!-- 全局认证检查加载界面 -->
	<div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100 dark:from-slate-900 dark:via-blue-900 dark:to-indigo-900 flex items-center justify-center">
		<div class="text-center">
			<div class="inline-flex items-center justify-center w-20 h-20 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl shadow-lg mb-4">
				<svg class="animate-spin h-10 w-10 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
					<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
					<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
				</svg>
			</div>
			<h2 class="text-2xl font-semibold text-slate-900 dark:text-slate-100 mb-2">正在验证登录状态</h2>
			<p class="text-slate-600 dark:text-slate-400">请稍候...</p>
		</div>
	</div>
{:else}
	<div class="min-h-screen bg-background font-sans antialiased">
		{@render children()}
	</div>
{/if}

<NotificationContainer />

<style>
	:global(html) {
		height: 100%;
	}
	:global(body) {
		height: 100%;
		margin: 0;
	}
</style>

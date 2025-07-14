<!-- 
应用主布局 - 现代化设计
使用现代化图标和优雅的视觉设计
-->
<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { cn } from '$lib/utils';
	import { auth } from '$lib/stores/auth';
	import { themeStore, themeActions, isDarkMode } from '$lib/stores/theme';
	import { Button } from '$lib/components/ui/button';
	import NotificationContainer from '$lib/components/ui/notifications/NotificationContainer.svelte';
	
	// 现代化图标
	import Icon from '@iconify/svelte';

	let { children } = $props();

	// 导航菜单配置 - 使用现代化图标
	const navigationItems = [
		{
			name: '仪表盘',
			href: '/dashboard',
			icon: 'tabler:home',
			description: '系统概览和统计',
			color: 'text-blue-500'
		},
		{
			name: '项目管理',
			href: '/projects',
			icon: 'tabler:folder',
			description: '管理和组织项目',
			color: 'text-emerald-500'
		},
		{
			name: '资产管理',
			href: '/assets',
			icon: 'tabler:target',
			description: '管理和监控资产',
			color: 'text-purple-500'
		},
		{
			name: '任务管理',
			href: '/tasks',
			icon: 'tabler:checklist',
			description: '扫描任务管理',
			color: 'text-orange-500'
		},
		{
			name: '节点管理',
			href: '/nodes',
			icon: 'tabler:device-desktop',
			description: '计算节点管理',
			color: 'text-cyan-500'
		},
		{
			name: '插件管理',
			href: '/plugins',
			icon: 'tabler:puzzle',
			description: '扫描插件管理',
			color: 'text-pink-500'
		},
		{
			name: '页面监控',
			href: '/monitoring',
			icon: 'tabler:eye',
			description: '页面变化监控',
			color: 'text-indigo-500'
		},
		{
			name: '敏感信息检测',
			href: '/sensitive',
			icon: 'tabler:shield',
			description: '敏感信息泄露检测',
			color: 'text-red-500'
		},
		{
			name: 'POC管理',
			href: '/pocs',
			icon: 'tabler:tool',
			description: '漏洞验证POC管理',
			color: 'text-yellow-500'
		},
		{
			name: '漏洞扫描',
			href: '/vulnerability/tasks',
			icon: 'tabler:search',
			description: '漏洞扫描任务管理',
			color: 'text-violet-500'
		},
		{
			name: '实时监控',
			href: '/vulnerability/monitor',
			icon: 'tabler:activity',
			description: '扫描任务实时监控',
			color: 'text-teal-500'
		}
	];

	// 侧边栏状态
	let sidebarCollapsed = $state(false);
	let isMobile = $state(false);

	// 检查是否为当前路径
	function isCurrentPath(href: string): boolean {
		if (href === '/dashboard') {
			return $page.url.pathname === '/dashboard';
		}
		return $page.url.pathname.startsWith(href);
	}

	// 切换侧边栏
	function toggleSidebar() {
		sidebarCollapsed = !sidebarCollapsed;
	}

	// 处理导航点击
	function handleNavigation(href: string) {
		goto(href);
		// 在移动端点击后收起侧边栏
		if (isMobile) {
			sidebarCollapsed = true;
		}
	}

	// 登出处理
	function handleLogout() {
		auth.logout();
		goto('/login');
	}

	// 响应式检测
	function checkMobile() {
		if (typeof window !== 'undefined') {
			isMobile = window.innerWidth < 768;
			if (isMobile) {
				sidebarCollapsed = true;
			}
		}
	}

	// 监听窗口大小变化
	if (typeof window !== 'undefined') {
		checkMobile();
		window.addEventListener('resize', checkMobile);
	}
</script>

<svelte:window on:resize={checkMobile} />

<div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50 dark:from-slate-900 dark:via-blue-900 dark:to-indigo-900 flex">
	<!-- 侧边导航栏 -->
	<aside
		class="fixed inset-y-0 left-0 z-50 flex flex-col bg-white/95 dark:bg-slate-900/95 backdrop-blur-xl border-r border-slate-200/50 dark:border-slate-700/50 shadow-xl transition-all duration-300 {sidebarCollapsed
			? 'w-16'
			: 'w-72'} {isMobile && sidebarCollapsed ? '-translate-x-full' : 'translate-x-0'}"
	>
		<!-- 侧边栏头部 -->
		<div class="flex items-center justify-between p-6 border-b border-slate-200/50 dark:border-slate-700/50">
			{#if !sidebarCollapsed}
				<div class="flex items-center space-x-3">
					<div
						class="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center text-white font-bold shadow-lg"
					>
						<Icon icon="tabler:shield" width={20} />
					</div>
					<div>
						<span class="text-xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">Stellar</span>
						<div class="text-xs text-slate-500 dark:text-slate-400">安全资产管理平台</div>
					</div>
				</div>
			{:else}
				<div
					class="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center text-white font-bold mx-auto shadow-lg"
				>
					<Icon icon="tabler:shield" width={20} />
				</div>
			{/if}

			{#if !isMobile}
				<button
					onclick={toggleSidebar}
					class="p-2 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-all duration-200"
				>
					{#if sidebarCollapsed}
						<Icon icon="tabler:chevron-right" width={16} />
					{:else}
						<Icon icon="tabler:chevron-left" width={16} />
					{/if}
				</button>
			{/if}
		</div>

		<!-- 导航菜单 -->
		<nav class="flex-1 px-3 py-6 space-y-2 overflow-y-auto">
			{#each navigationItems as item}
				<button
					type="button"
					onclick={() => handleNavigation(item.href)}
					class="w-full group flex items-center px-3 py-3 text-sm font-medium rounded-xl transition-all duration-200 {isCurrentPath(
						item.href
					)
						? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-lg transform scale-[1.02]'
						: 'text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-white hover:transform hover:scale-[1.01]'}"
				>
					<div class="flex items-center justify-center w-5 h-5 mr-3">
						<Icon 
							icon={item.icon}
							width={18} 
							class={isCurrentPath(item.href) ? 'text-white' : item.color}
						/>
					</div>
					{#if !sidebarCollapsed}
						<div class="flex-1 text-left">
							<div class="font-medium">{item.name}</div>
							<div class="text-xs opacity-70 mt-0.5">{item.description}</div>
						</div>
					{/if}
				</button>
			{/each}
		</nav>

		<!-- 侧边栏底部 -->
		<div class="border-t border-slate-200/50 dark:border-slate-700/50 p-4">
			{#if !sidebarCollapsed}
				<div class="space-y-2">
					<button
						onclick={() => goto('/settings')}
						class="w-full flex items-center px-3 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-white rounded-lg transition-all duration-200"
					>
						<Icon icon="tabler:settings" width={16} class="mr-3 text-slate-400" />
						设置
					</button>
					<button
						onclick={handleLogout}
						class="w-full flex items-center px-3 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-white rounded-lg transition-all duration-200"
					>
						<Icon icon="tabler:logout" width={16} class="mr-3 text-slate-400" />
						登出
					</button>
				</div>
			{:else}
				<div class="space-y-2">
					<button
						onclick={() => goto('/settings')}
						class="w-full p-2 text-slate-600 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-all duration-200"
						title="设置"
					>
						<Icon icon="tabler:settings" width={16} />
					</button>
					<button
						onclick={handleLogout}
						class="w-full p-2 text-slate-600 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-all duration-200"
						title="登出"
					>
						<Icon icon="tabler:logout" width={16} />
					</button>
				</div>
			{/if}
		</div>
	</aside>

	<!-- 主内容区域 -->
	<div class="flex-1 flex flex-col {sidebarCollapsed ? 'ml-16' : 'ml-72'} {isMobile ? 'ml-0' : ''} transition-all duration-300">
		<!-- 顶部导航栏 -->
		<header
			class="sticky top-0 z-40 w-full border-b border-slate-200/50 dark:border-slate-700/50 bg-white/95 dark:bg-slate-900/95 backdrop-blur-xl shadow-sm"
		>
			<div class="flex h-16 items-center justify-between px-6">
				<!-- 左侧：移动端菜单按钮 + 面包屑 -->
				<div class="flex items-center space-x-4">
					{#if isMobile}
						<button
							onclick={toggleSidebar}
							class="p-2 rounded-lg text-slate-600 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-all duration-200"
						>
							<Icon icon="tabler:menu-2" width={20} />
						</button>
					{/if}

					<!-- 面包屑导航 -->
					<nav class="flex items-center space-x-2 text-sm">
						<a href="/dashboard" class="text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 transition-colors font-medium">
							首页
						</a>
						{#if $page.url.pathname !== '/dashboard'}
							<span class="text-slate-300 dark:text-slate-600">/</span>
							{#if $page.url.pathname.startsWith('/projects')}
								<span class="text-slate-900 dark:text-slate-100 font-medium">项目管理</span>
							{:else if $page.url.pathname.startsWith('/assets')}
								<span class="text-slate-900 dark:text-slate-100 font-medium">资产管理</span>
							{:else if $page.url.pathname.startsWith('/tasks')}
								<span class="text-slate-900 dark:text-slate-100 font-medium">任务管理</span>
							{:else if $page.url.pathname.startsWith('/nodes')}
								<span class="text-slate-900 dark:text-slate-100 font-medium">节点管理</span>
							{:else if $page.url.pathname.startsWith('/settings')}
								<span class="text-slate-900 dark:text-slate-100 font-medium">设置</span>
							{/if}
						{/if}
					</nav>
				</div>

				<!-- 右侧：用户信息和操作 -->
				<div class="flex items-center space-x-3">
					<!-- 主题切换按钮 -->
					<button
						class="p-2 rounded-lg text-slate-600 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-all duration-200"
						title={$isDarkMode ? '切换到亮色模式' : '切换到暗色模式'}
						onclick={themeActions.toggleMode}
					>
						{#if $isDarkMode}
							<Icon icon="tabler:sun" width={18} />
						{:else}
							<Icon icon="tabler:moon" width={18} />
						{/if}
					</button>

					<!-- 通知图标 -->
					<button 
						class="relative p-2 rounded-lg text-slate-600 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-all duration-200" 
						title="通知"
					>
						<Icon icon="tabler:bell" width={18} />
						<span class="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
					</button>

					<!-- 帮助图标 -->
					<button 
						class="p-2 rounded-lg text-slate-600 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-all duration-200" 
						title="帮助"
					>
						<Icon icon="tabler:help-circle" width={18} />
					</button>

					<!-- 用户菜单 -->
					<div class="flex items-center space-x-3 pl-3 border-l border-slate-200 dark:border-slate-700">
						<div
							class="w-9 h-9 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center text-white shadow-lg"
						>
							<Icon icon="tabler:user" width={18} />
						</div>
						{#if !isMobile}
							<div class="text-sm">
								<div class="font-medium text-slate-900 dark:text-slate-100">管理员</div>
								<div class="text-slate-500 dark:text-slate-400">admin@stellar.com</div>
							</div>
						{/if}
					</div>
				</div>
			</div>
		</header>

		<!-- 主要内容 -->
		<main class="flex-1 overflow-y-auto bg-transparent">
			<div class="container mx-auto px-6 py-8">
				{@render children()}
			</div>
		</main>

		<!-- 底部状态栏 -->
		<footer class="border-t border-slate-200/50 dark:border-slate-700/50 bg-white/95 dark:bg-slate-900/95 backdrop-blur-xl px-6 py-3 shadow-sm">
			<div class="flex items-center justify-between text-xs">
				<div class="flex items-center space-x-4 text-slate-500 dark:text-slate-400">
					<span class="flex items-center space-x-2">
						<span class="font-medium text-slate-700 dark:text-slate-300">Stellar</span>
						<span class="px-2 py-0.5 bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 rounded-full text-xs font-medium">v1.0.0</span>
					</span>
					<span class="flex items-center space-x-2">
						<div class="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></div>
						<span>系统正常运行</span>
					</span>
				</div>
				<div class="text-slate-500 dark:text-slate-400">
					<span>最后更新: {new Date().toLocaleString('zh-CN')}</span>
				</div>
			</div>
		</footer>
	</div>
</div>

<!-- 移动端遮罩层 -->
{#if isMobile && !sidebarCollapsed}
	<div
		class="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm"
		onclick={() => (sidebarCollapsed = true)}
		onkeydown={(e) => {
			if (e.key === 'Escape') sidebarCollapsed = true;
		}}
		role="button"
		tabindex="0"
		aria-label="关闭侧边栏"
	></div>
{/if}

<!-- 通知容器 -->
<NotificationContainer />

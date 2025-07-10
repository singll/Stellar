<!-- 
应用主布局 - 改进版本
使用侧边导航 + 顶部栏的布局，提升用户体验
-->
<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { cn } from '$lib/utils';
	import { auth } from '$lib/stores/auth';
	import { themeStore, themeActions, isDarkMode } from '$lib/stores/theme';
	import { Button } from '$lib/components/ui/button';
	import NotificationContainer from '$lib/components/ui/notifications/NotificationContainer.svelte';

	let { children } = $props();

	// 导航菜单配置
	const navigationItems = [
		{
			name: '仪表盘',
			href: '/dashboard',
			icon: '📊',
			description: '系统概览和统计'
		},
		{
			name: '项目管理',
			href: '/projects',
			icon: '📁',
			description: '管理和组织项目'
		},
		{
			name: '资产管理',
			href: '/assets',
			icon: '🎯',
			description: '管理和监控资产'
		},
		{
			name: '任务管理',
			href: '/tasks',
			icon: '📋',
			description: '扫描任务管理'
		},
		{
			name: '节点管理',
			href: '/nodes',
			icon: '🖥️',
			description: '计算节点管理'
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

<div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50 flex">
	<!-- 侧边导航栏 -->
	<aside
		class="fixed inset-y-0 left-0 z-50 flex flex-col glass border-r border-white/20 shadow-medium transition-all duration-300 {sidebarCollapsed
			? 'w-16'
			: 'w-64'} {isMobile && sidebarCollapsed ? '-translate-x-full' : 'translate-x-0'}"
	>
		<!-- 侧边栏头部 -->
		<div class="flex items-center justify-between p-4 border-b border-white/10">
			{#if !sidebarCollapsed}
				<div class="flex items-center space-x-2">
					<div
						class="w-8 h-8 gradient-bg rounded-xl flex items-center justify-center text-white font-bold shadow-soft"
					>
						S
					</div>
					<span class="text-xl font-bold text-slate-800">Stellar</span>
				</div>
			{:else}
				<div
					class="w-8 h-8 gradient-bg rounded-xl flex items-center justify-center text-white font-bold mx-auto shadow-soft"
				>
					S
				</div>
			{/if}

			{#if !isMobile}
				<button 
					onclick={toggleSidebar} 
					class="modern-btn-ghost text-slate-600 hover:text-slate-800 hover:bg-white/50 p-1.5"
				>
					{sidebarCollapsed ? '→' : '←'}
				</button>
			{/if}
		</div>

		<!-- 导航菜单 -->
		<nav class="flex-1 px-2 py-4 space-y-1 overflow-y-auto">
			{#each navigationItems as item}
				<button
					type="button"
					onclick={() => handleNavigation(item.href)}
					class="w-full group flex items-center px-3 py-2.5 text-sm font-medium rounded-xl transition-all duration-200 hover:scale-[1.02] {isCurrentPath(
						item.href
					)
						? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-soft'
						: 'text-slate-700 hover:bg-white/50 hover:text-slate-900'}"
				>
					<span class="text-lg mr-3">{item.icon}</span>
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
		<div class="border-t border-white/10 p-4">
			{#if !sidebarCollapsed}
				<div class="space-y-2">
					<button
						onclick={() => goto('/settings')}
						class="modern-btn-ghost w-full justify-start text-slate-600 hover:text-slate-800 hover:bg-white/50"
					>
						<span class="mr-2">⚙️</span>
						设置
					</button>
					<button onclick={handleLogout} class="modern-btn-ghost w-full justify-start text-slate-600 hover:text-slate-800 hover:bg-white/50">
						<span class="mr-2">🚪</span>
						登出
					</button>
				</div>
			{:else}
				<div class="space-y-2">
					<button
						onclick={() => goto('/settings')}
						class="modern-btn-ghost w-full p-2 text-slate-600 hover:text-slate-800 hover:bg-white/50"
						title="设置"
					>
						⚙️
					</button>
					<button onclick={handleLogout} class="modern-btn-ghost w-full p-2 text-slate-600 hover:text-slate-800 hover:bg-white/50" title="登出">
						🚪
					</button>
				</div>
			{/if}
		</div>
	</aside>

	<!-- 主内容区域 -->
	<div class="flex-1 flex flex-col {sidebarCollapsed ? 'ml-16' : 'ml-64'} {isMobile ? 'ml-0' : ''}">
		<!-- 顶部导航栏 -->
		<header
			class="sticky top-0 z-40 w-full border-b border-white/10 bg-white/90 backdrop-blur-md shadow-soft"
		>
			<div class="flex h-16 items-center justify-between px-4">
				<!-- 左侧：移动端菜单按钮 + 面包屑 -->
				<div class="flex items-center space-x-4">
					{#if isMobile}
						<button onclick={toggleSidebar} class="modern-btn-ghost p-2 text-slate-600 hover:text-slate-800">☰</button>
					{/if}

					<!-- 面包屑导航 -->
					<nav class="flex items-center space-x-2 text-sm">
						<a href="/dashboard" class="text-slate-500 hover:text-slate-700 transition-colors"> 首页 </a>
						{#if $page.url.pathname !== '/dashboard'}
							<span class="text-slate-300">/</span>
							{#if $page.url.pathname.startsWith('/projects')}
								<span class="text-slate-900 font-medium">项目管理</span>
							{:else if $page.url.pathname.startsWith('/assets')}
								<span class="text-slate-900 font-medium">资产管理</span>
							{:else if $page.url.pathname.startsWith('/tasks')}
								<span class="text-slate-900 font-medium">任务管理</span>
							{:else if $page.url.pathname.startsWith('/nodes')}
								<span class="text-slate-900 font-medium">节点管理</span>
							{:else if $page.url.pathname.startsWith('/settings')}
								<span class="text-slate-900 font-medium">设置</span>
							{/if}
						{/if}
					</nav>
				</div>

				<!-- 右侧：用户信息和操作 -->
				<div class="flex items-center space-x-4">
					<!-- 主题切换按钮 -->
					<button
						class="modern-btn-ghost p-2 text-slate-600 hover:text-slate-800"
						title={$isDarkMode ? '切换到亮色模式' : '切换到暗色模式'}
						onclick={themeActions.toggleMode}
					>
						{$isDarkMode ? '☀️' : '🌙'}
					</button>

					<!-- 通知图标 -->
					<button class="modern-btn-ghost p-2 text-slate-600 hover:text-slate-800" title="通知">🔔</button>

					<!-- 帮助图标 -->
					<button class="modern-btn-ghost p-2 text-slate-600 hover:text-slate-800" title="帮助">❓</button>

					<!-- 用户菜单 -->
					<div class="flex items-center space-x-2">
						<div class="w-8 h-8 bg-gradient-to-br from-blue-400 to-purple-500 rounded-full flex items-center justify-center text-white shadow-soft">👤</div>
						{#if !isMobile}
							<div class="text-sm">
								<div class="font-medium text-slate-900">管理员</div>
								<div class="text-slate-500">admin@stellar.com</div>
							</div>
						{/if}
					</div>
				</div>
			</div>
		</header>

		<!-- 主要内容 -->
		<main class="flex-1 overflow-y-auto">
			<div class="container mx-auto px-4 py-6">
				{@render children()}
			</div>
		</main>

		<!-- 底部状态栏 -->
		<footer class="border-t border-white/10 bg-white/90 backdrop-blur-sm px-4 py-2 shadow-soft">
			<div class="flex items-center justify-between text-xs text-slate-500">
				<div class="flex items-center space-x-4">
					<span class="font-medium">Stellar v1.0.0</span>
					<span class="flex items-center space-x-1">
						<span class="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></span>
						<span>系统正常</span>
					</span>
				</div>
				<div class="flex items-center space-x-4">
					<span>最后更新: {new Date().toLocaleString('zh-CN')}</span>
				</div>
			</div>
		</footer>
	</div>
</div>

<!-- 移动端遮罩层 -->
{#if isMobile && !sidebarCollapsed}
	<div
		class="fixed inset-0 z-40 bg-black bg-opacity-50"
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

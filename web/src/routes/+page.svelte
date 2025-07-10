<!--
系统首页/欢迎页面
为新用户和访客提供系统介绍和快速入门
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import Button from '$lib/components/ui/Button.svelte';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';

	// 检查是否已登录
	$effect(() => {
		if ($auth.isAuthenticated) {
			goto('/dashboard');
		}
	});

	// 功能特性
	const features = [
		{
			icon: '🎯',
			title: '资产管理',
			description: '统一管理和监控域名、IP、端口等数字资产，实现资产全生命周期跟踪。'
		},
		{
			icon: '🔍',
			title: '安全扫描',
			description: '支持子域名枚举、端口扫描、漏洞检测等多种安全检测能力。'
		},
		{
			icon: '📋',
			title: '任务调度',
			description: '分布式任务调度引擎，支持大规模并发扫描和智能任务分发。'
		},
		{
			icon: '🖥️',
			title: '节点管理',
			description: '多节点集群管理，支持弹性扩容和负载均衡，提升扫描效率。'
		},
		{
			icon: '📊',
			title: '数据可视化',
			description: '丰富的图表和报表，让安全数据可视化，便于分析和决策。'
		},
		{
			icon: '🔧',
			title: '插件扩展',
			description: '开放的插件架构，支持自定义扫描器和安全工具集成。'
		}
	];

	// 快速入门步骤
	const quickStartSteps = [
		{
			step: 1,
			title: '登录系统',
			description: '使用管理员账户登录到 Stellar 安全平台。',
			action: '立即登录',
			href: '/login'
		},
		{
			step: 2,
			title: '创建项目',
			description: '创建您的第一个安全扫描项目，组织和管理目标资产。',
			action: '创建项目',
			href: '/projects/create'
		},
		{
			step: 3,
			title: '添加资产',
			description: '导入或手动添加需要扫描的域名、IP 地址等目标资产。',
			action: '管理资产',
			href: '/assets'
		},
		{
			step: 4,
			title: '启动扫描',
			description: '配置扫描任务参数，启动安全扫描并监控执行状态。',
			action: '创建任务',
			href: '/tasks/create'
		}
	];
</script>

<svelte:head>
	<title>Stellar - 分布式安全资产管理与漏洞扫描平台</title>
	<meta
		name="description"
		content="Stellar 是一个现代化的分布式安全资产管理与漏洞扫描平台，提供全面的安全检测和资产管理能力。"
	/>
</svelte:head>

<div class="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
	<!-- 头部导航 -->
	<header class="relative z-10">
		<nav class="container mx-auto px-6 py-4">
			<div class="flex items-center justify-between">
				<div class="flex items-center space-x-3">
					<div
						class="w-10 h-10 bg-blue-600 rounded-xl flex items-center justify-center text-white font-bold text-lg"
					>
						S
					</div>
					<div>
						<h1 class="text-2xl font-bold text-gray-900">Stellar</h1>
						<p class="text-xs text-gray-600">安全资产管理平台</p>
					</div>
				</div>

				<div class="flex items-center space-x-4">
					<Button variant="outline" onclick={() => goto('/login')}>登录</Button>
					<Button variant="primary" onclick={() => goto('/register')}>注册</Button>
				</div>
			</div>
		</nav>
	</header>

	<!-- 主要内容 -->
	<main>
		<!-- Hero 区域 -->
		<section class="relative py-20 px-6">
			<div class="container mx-auto text-center">
				<div class="max-w-4xl mx-auto">
					<h1 class="text-5xl md:text-6xl font-bold text-gray-900 mb-6">
						分布式安全资产管理
						<span
							class="text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-purple-600"
						>
							与漏洞扫描平台
						</span>
					</h1>

					<p class="text-xl text-gray-600 mb-10 leading-relaxed">
						Stellar 为企业提供全面的数字资产发现、安全扫描和风险管理能力。
						基于分布式架构，支持大规模资产扫描和智能威胁检测。
					</p>

					<div class="flex flex-col sm:flex-row gap-4 justify-center">
						<Button
							variant="primary"
							size="lg"
							class="text-lg px-8 py-4"
							onclick={() => goto('/login')}
						>
							🚀 开始使用
						</Button>
						<Button
							variant="outline"
							size="lg"
							class="text-lg px-8 py-4"
							onclick={() => goto('/dashboard')}
						>
							📊 查看演示
						</Button>
					</div>
				</div>
			</div>

			<!-- 背景装饰 -->
			<div class="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none">
				<div
					class="absolute top-10 left-10 w-20 h-20 bg-blue-200 rounded-full opacity-50 animate-pulse"
				></div>
				<div
					class="absolute bottom-10 right-10 w-32 h-32 bg-purple-200 rounded-full opacity-50 animate-pulse delay-1000"
				></div>
				<div
					class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-40 h-40 bg-blue-100 rounded-full opacity-30 animate-ping"
				></div>
			</div>
		</section>

		<!-- 功能特性 -->
		<section class="py-20 px-6 bg-white">
			<div class="container mx-auto">
				<div class="text-center mb-16">
					<h2 class="text-4xl font-bold text-gray-900 mb-4">核心功能特性</h2>
					<p class="text-xl text-gray-600">全面的安全检测和资产管理能力</p>
				</div>

				<div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
					{#each features as feature}
						<Card class="hover:shadow-lg transition-shadow duration-300">
							<CardHeader>
								<div
									class="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center text-2xl mb-4"
								>
									{feature.icon}
								</div>
								<CardTitle class="text-xl">{feature.title}</CardTitle>
							</CardHeader>
							<CardContent>
								<CardDescription class="text-gray-600 leading-relaxed">
									{feature.description}
								</CardDescription>
							</CardContent>
						</Card>
					{/each}
				</div>
			</div>
		</section>

		<!-- 快速入门 -->
		<section class="py-20 px-6 bg-gray-50">
			<div class="container mx-auto">
				<div class="text-center mb-16">
					<h2 class="text-4xl font-bold text-gray-900 mb-4">快速入门</h2>
					<p class="text-xl text-gray-600">四个步骤，快速开始安全扫描</p>
				</div>

				<div class="max-w-4xl mx-auto">
					<div class="grid md:grid-cols-2 gap-8">
						{#each quickStartSteps as step}
							<Card class="relative hover:shadow-lg transition-shadow duration-300">
								<CardHeader>
									<div class="flex items-center space-x-4">
										<div
											class="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center text-white font-bold"
										>
											{step.step}
										</div>
										<CardTitle class="text-xl">{step.title}</CardTitle>
									</div>
								</CardHeader>
								<CardContent>
									<CardDescription class="text-gray-600 mb-4 leading-relaxed">
										{step.description}
									</CardDescription>
									<Button variant="outline" onclick={() => goto(step.href)}>
										{step.action}
									</Button>
								</CardContent>
							</Card>
						{/each}
					</div>
				</div>
			</div>
		</section>

		<!-- 技术特点 -->
		<section class="py-20 px-6 bg-white">
			<div class="container mx-auto">
				<div class="max-w-6xl mx-auto">
					<div class="grid lg:grid-cols-2 gap-12 items-center">
						<!-- 左侧：技术特点 -->
						<div>
							<h2 class="text-4xl font-bold text-gray-900 mb-6">技术架构优势</h2>
							<div class="space-y-6">
								<div class="flex items-start space-x-4">
									<div
										class="w-8 h-8 bg-green-100 rounded-lg flex items-center justify-center flex-shrink-0"
									>
										⚡
									</div>
									<div>
										<h3 class="text-lg font-semibold text-gray-900 mb-2">高性能并发</h3>
										<p class="text-gray-600">
											基于 Go 语言构建，支持大规模并发扫描，单节点可处理数万个目标。
										</p>
									</div>
								</div>

								<div class="flex items-start space-x-4">
									<div
										class="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center flex-shrink-0"
									>
										🏗️
									</div>
									<div>
										<h3 class="text-lg font-semibold text-gray-900 mb-2">分布式架构</h3>
										<p class="text-gray-600">
											支持多节点集群部署，自动负载均衡和任务分发，轻松扩展计算能力。
										</p>
									</div>
								</div>

								<div class="flex items-start space-x-4">
									<div
										class="w-8 h-8 bg-purple-100 rounded-lg flex items-center justify-center flex-shrink-0"
									>
										🔧
									</div>
									<div>
										<h3 class="text-lg font-semibold text-gray-900 mb-2">插件生态</h3>
										<p class="text-gray-600">
											开放的插件架构，支持自定义扫描器和第三方安全工具集成。
										</p>
									</div>
								</div>

								<div class="flex items-start space-x-4">
									<div
										class="w-8 h-8 bg-yellow-100 rounded-lg flex items-center justify-center flex-shrink-0"
									>
										📱
									</div>
									<div>
										<h3 class="text-lg font-semibold text-gray-900 mb-2">现代化界面</h3>
										<p class="text-gray-600">
											基于 Svelte 5 构建的响应式用户界面，提供优秀的用户体验。
										</p>
									</div>
								</div>
							</div>
						</div>

						<!-- 右侧：技术栈展示 -->
						<div class="bg-gray-50 rounded-2xl p-8">
							<h3 class="text-2xl font-bold text-gray-900 mb-6">技术栈</h3>
							<div class="space-y-4">
								<div class="bg-white rounded-lg p-4 shadow-sm">
									<div class="flex items-center justify-between">
										<span class="font-medium text-gray-900">后端</span>
										<span class="text-blue-600">Go + Gin + MongoDB</span>
									</div>
								</div>
								<div class="bg-white rounded-lg p-4 shadow-sm">
									<div class="flex items-center justify-between">
										<span class="font-medium text-gray-900">前端</span>
										<span class="text-green-600">Svelte 5 + TypeScript</span>
									</div>
								</div>
								<div class="bg-white rounded-lg p-4 shadow-sm">
									<div class="flex items-center justify-between">
										<span class="font-medium text-gray-900">数据库</span>
										<span class="text-yellow-600">MongoDB + Redis</span>
									</div>
								</div>
								<div class="bg-white rounded-lg p-4 shadow-sm">
									<div class="flex items-center justify-between">
										<span class="font-medium text-gray-900">部署</span>
										<span class="text-purple-600">Docker + Kubernetes</span>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</section>

		<!-- CTA 区域 -->
		<section class="py-20 px-6 bg-gradient-to-r from-blue-600 to-purple-600">
			<div class="container mx-auto text-center">
				<div class="max-w-3xl mx-auto">
					<h2 class="text-4xl font-bold text-white mb-6">开始您的安全扫描之旅</h2>
					<p class="text-xl text-blue-100 mb-10">
						立即注册 Stellar 账户，体验强大的分布式安全扫描能力
					</p>
					<div class="flex flex-col sm:flex-row gap-4 justify-center">
						<Button
							variant="secondary"
							size="lg"
							class="text-lg px-8 py-4 bg-white text-blue-600 hover:bg-gray-50"
							onclick={() => goto('/register')}
						>
							🎯 免费注册
						</Button>
						<Button
							variant="outline"
							size="lg"
							class="text-lg px-8 py-4 border-white text-white hover:bg-white hover:text-blue-600"
							onclick={() => goto('/login')}
						>
							🚀 立即登录
						</Button>
					</div>
				</div>
			</div>
		</section>
	</main>

	<!-- 底部信息 -->
	<footer class="bg-gray-900 text-white py-12">
		<div class="container mx-auto px-6">
			<div class="text-center">
				<div class="flex items-center justify-center space-x-3 mb-4">
					<div
						class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold"
					>
						S
					</div>
					<span class="text-xl font-bold">Stellar</span>
				</div>
				<p class="text-gray-400 mb-4">分布式安全资产管理与漏洞扫描平台</p>
				<div
					class="flex flex-col sm:flex-row gap-4 justify-center items-center text-sm text-gray-500"
				>
					<span>© 2024 Stellar. All rights reserved.</span>
					<span class="hidden sm:block">•</span>
					<span>版本 v1.0.0</span>
					<span class="hidden sm:block">•</span>
					<span>基于开源技术构建</span>
				</div>
			</div>
		</div>
	</footer>
</div>

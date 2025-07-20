<!--
  @component
  Settings page - Main settings dashboard
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import Icon from '@iconify/svelte';

	// 设置选项配置
	const settingsOptions = [
		{
			id: 'themes',
			title: '主题设置',
			description: '自定义应用外观和主题',
			icon: 'tabler:palette',
			color: 'text-purple-500',
			bgColor: 'bg-purple-50 dark:bg-purple-900/20',
			href: '/settings/themes'
		},
		{
			id: 'profile',
			title: '个人资料',
			description: '管理个人信息和账户设置',
			icon: 'tabler:user',
			color: 'text-blue-500',
			bgColor: 'bg-blue-50 dark:bg-blue-900/20',
			href: '/settings/profile'
		},
		{
			id: 'security',
			title: '安全设置',
			description: '密码、双因素认证等安全选项',
			icon: 'tabler:shield-lock',
			color: 'text-red-500',
			bgColor: 'bg-red-50 dark:bg-red-900/20',
			href: '/settings/security'
		},
		{
			id: 'notifications',
			title: '通知设置',
			description: '配置邮件和系统通知',
			icon: 'tabler:bell',
			color: 'text-orange-500',
			bgColor: 'bg-orange-50 dark:bg-orange-900/20',
			href: '/settings/notifications'
		},
		{
			id: 'system',
			title: '系统设置',
			description: '系统配置和高级选项',
			icon: 'tabler:settings',
			color: 'text-green-500',
			bgColor: 'bg-green-50 dark:bg-green-900/20',
			href: '/settings/system'
		},
		{
			id: 'api',
			title: 'API 设置',
			description: 'API 密钥和集成配置',
			icon: 'tabler:key',
			color: 'text-cyan-500',
			bgColor: 'bg-cyan-50 dark:bg-cyan-900/20',
			href: '/settings/api'
		}
	];

	function handleSettingClick(href: string) {
		goto(href);
	}
</script>

<div class="container py-10">
	<!-- 页面标题 -->
	<div class="mb-8">
		<h1 class="text-3xl font-bold tracking-tight text-slate-900 dark:text-slate-100">
			系统设置
		</h1>
		<p class="text-slate-600 dark:text-slate-400 mt-2">
			管理您的账户设置、主题偏好和系统配置
		</p>
	</div>

	<!-- 设置选项网格 -->
	<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
		{#each settingsOptions as option}
			<Card 
				class="cursor-pointer transition-all duration-200 hover:shadow-lg hover:scale-[1.02] border-slate-200 dark:border-slate-700"
				onclick={() => handleSettingClick(option.href)}
			>
				<CardHeader class="pb-4">
					<div class="flex items-center space-x-3">
						<div class="w-10 h-10 rounded-lg {option.bgColor} flex items-center justify-center">
							<Icon 
								icon={option.icon} 
								width={20} 
								class={option.color}
							/>
						</div>
						<div>
							<CardTitle class="text-lg text-slate-900 dark:text-slate-100">
								{option.title}
							</CardTitle>
							<CardDescription class="text-slate-600 dark:text-slate-400">
								{option.description}
							</CardDescription>
						</div>
					</div>
				</CardHeader>
				<CardContent class="pt-0">
					<Button 
						variant="ghost" 
						class="w-full justify-start text-slate-600 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300"
					>
						管理设置
						<Icon icon="tabler:arrow-right" width={16} class="ml-2" />
					</Button>
				</CardContent>
			</Card>
		{/each}
	</div>

	<!-- 快速操作区域 -->
	<div class="mt-12">
		<h2 class="text-xl font-semibold text-slate-900 dark:text-slate-100 mb-4">
			快速操作
		</h2>
		<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
			<Button 
				variant="outline" 
				class="h-auto p-4 flex flex-col items-center space-y-2"
				onclick={() => goto('/dashboard')}
			>
				<Icon icon="tabler:home" width={20} class="text-slate-600 dark:text-slate-400" />
				<span class="text-sm">返回仪表盘</span>
			</Button>
			
			<Button 
				variant="outline" 
				class="h-auto p-4 flex flex-col items-center space-y-2"
				onclick={() => window.open('/api/docs', '_blank')}
			>
				<Icon icon="tabler:book" width={20} class="text-slate-600 dark:text-slate-400" />
				<span class="text-sm">API 文档</span>
			</Button>
			
			<Button 
				variant="outline" 
				class="h-auto p-4 flex flex-col items-center space-y-2"
				onclick={() => window.open('/help', '_blank')}
			>
				<Icon icon="tabler:help" width={20} class="text-slate-600 dark:text-slate-400" />
				<span class="text-sm">帮助中心</span>
			</Button>
			
			<Button 
				variant="outline" 
				class="h-auto p-4 flex flex-col items-center space-y-2"
				onclick={() => window.open('/about', '_blank')}
			>
				<Icon icon="tabler:info-circle" width={20} class="text-slate-600 dark:text-slate-400" />
				<span class="text-sm">关于系统</span>
			</Button>
		</div>
	</div>
</div> 
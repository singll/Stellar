<!-- 登录页面 - 现代化设计 -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index';
	import {
		Card,
		CardContent,
		CardDescription,
		CardFooter,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index';
	import { Input } from '$lib/components/ui/input/index';
	import { Label } from '$lib/components/ui/label/index';
	import { authApi } from '$lib/api/auth';
	import { auth } from '$lib/stores/auth';
	import { notifications } from '$lib/stores/notifications';
	import { createForm } from 'felte';
	import { validator } from '@felte/validator-zod';
	import { z } from 'zod';
	
	// 现代化图标
	import Icon from '@iconify/svelte';

	const schema = z.object({
		username: z.string().min(1, '用户名不能为空'),
		password: z.string().min(1, '密码不能为空')
	});

	type LoginForm = z.infer<typeof schema>;

	let isLoading = $state(false);
	let showPassword = $state(false);

	const { form, errors, isSubmitting } = createForm<LoginForm>({
		onSubmit: async (values) => {
			isLoading = true;
			try {
				const credentials = {
					username: values.username,
					password: values.password
				};
				await auth.login(credentials);
				await goto('/dashboard');
			} catch (error) {
				notifications.add({
					type: 'error',
					message: error instanceof Error ? error.message : '发生未知错误，请稍后再试'
				});
			} finally {
				isLoading = false;
			}
		},
		extend: validator({ schema })
	});
</script>

<svelte:head>
	<title>登录 - Stellar 安全资产管理平台</title>
</svelte:head>

<div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100 dark:from-slate-900 dark:via-blue-900 dark:to-indigo-900 flex items-center justify-center p-4">
	<!-- 背景装饰 -->
	<div class="absolute inset-0 overflow-hidden">
		<div class="absolute -top-40 -right-40 w-80 h-80 bg-blue-400/20 rounded-full blur-3xl"></div>
		<div class="absolute -bottom-40 -left-40 w-80 h-80 bg-purple-400/20 rounded-full blur-3xl"></div>
		<div class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-indigo-400/10 rounded-full blur-3xl"></div>
	</div>

	<!-- 登录卡片 -->
	<div class="relative z-10 w-full max-w-md">
		<!-- 品牌标识 -->
		<div class="text-center mb-8">
			<div class="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl shadow-lg mb-4">
				<Icon icon="tabler:shield" width={32} class="text-white" />
			</div>
			<h1 class="text-3xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
				Stellar
			</h1>
			<p class="text-slate-600 dark:text-slate-400 mt-1">
				安全资产管理平台
			</p>
		</div>

		<!-- 登录表单 -->
		<Card class="bg-white/95 dark:bg-slate-900/95 backdrop-blur-xl border-slate-200/50 dark:border-slate-700/50 shadow-2xl">
			<CardHeader class="space-y-2 text-center pb-6">
				<CardTitle class="text-2xl font-bold text-slate-900 dark:text-slate-100">
					欢迎回来
				</CardTitle>
				<CardDescription class="text-slate-600 dark:text-slate-400">
					请输入您的凭据以访问您的账户
				</CardDescription>
			</CardHeader>
			
			<CardContent class="space-y-6">
				<form use:form class="space-y-5">
					<!-- 用户名输入 -->
					<div class="space-y-2">
						<Label for="username" class="text-sm font-medium text-slate-700 dark:text-slate-300">
							用户名或邮箱
						</Label>
						<div class="relative">
							<div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
								<Icon icon="tabler:mail" width={18} class="text-slate-400" />
							</div>
							<Input
								id="username"
								name="username"
								type="text"
								placeholder="请输入用户名或邮箱"
								disabled={isLoading || $isSubmitting}
								class="pl-10 h-11 bg-white/50 dark:bg-slate-800/50 border-slate-300 dark:border-slate-600 focus:border-blue-500 dark:focus:border-blue-400 focus:ring-blue-500/20 transition-all duration-200"
							/>
						</div>
						{#if $errors.username}
							<div class="flex items-center gap-2 text-sm text-red-600 dark:text-red-400">
								<Icon icon="tabler:alert-circle" width={14} />
								{$errors.username}
							</div>
						{/if}
					</div>

					<!-- 密码输入 -->
					<div class="space-y-2">
						<Label for="password" class="text-sm font-medium text-slate-700 dark:text-slate-300">
							密码
						</Label>
						<div class="relative">
							<div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
								<Icon icon="tabler:lock" width={18} class="text-slate-400" />
							</div>
							<Input
								id="password"
								name="password"
								type={showPassword ? 'text' : 'password'}
								placeholder="请输入密码"
								disabled={isLoading || $isSubmitting}
								class="pl-10 pr-10 h-11 bg-white/50 dark:bg-slate-800/50 border-slate-300 dark:border-slate-600 focus:border-blue-500 dark:focus:border-blue-400 focus:ring-blue-500/20 transition-all duration-200"
							/>
							<button
								type="button"
								onclick={() => showPassword = !showPassword}
								class="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors"
							>
								{#if showPassword}
									<Icon icon="tabler:eye-off" width={18} />
								{:else}
									<Icon icon="tabler:eye" width={18} />
								{/if}
							</button>
						</div>
						{#if $errors.password}
							<div class="flex items-center gap-2 text-sm text-red-600 dark:text-red-400">
								<Icon icon="tabler:alert-circle" width={14} />
								{$errors.password}
							</div>
						{/if}
					</div>

					<!-- 记住我和忘记密码 -->
					<div class="flex items-center justify-between">
						<label class="flex items-center space-x-2 text-sm">
							<input 
								type="checkbox" 
								class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500/20"
							>
							<span class="text-slate-600 dark:text-slate-400">记住我</span>
						</label>
						<a 
							href="/forgot-password" 
							class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 font-medium transition-colors"
						>
							忘记密码？
						</a>
					</div>

					<!-- 登录按钮 -->
					<Button 
						type="submit" 
						class="w-full h-11 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-medium shadow-lg hover:shadow-xl transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
						disabled={isLoading || $isSubmitting}
					>
						{#if isLoading || $isSubmitting}
							<div class="flex items-center justify-center space-x-2">
								<div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
								<span>登录中...</span>
							</div>
						{:else}
							<div class="flex items-center justify-center space-x-2">
								<span>登录</span>
								<Icon icon="tabler:arrow-right" width={16} />
							</div>
						{/if}
					</Button>
				</form>

				<!-- 分割线 -->
				<div class="relative">
					<div class="absolute inset-0 flex items-center">
						<div class="w-full border-t border-slate-200 dark:border-slate-700"></div>
					</div>
					<div class="relative flex justify-center text-sm">
						<span class="px-4 bg-white dark:bg-slate-900 text-slate-500 dark:text-slate-400">
							或者
						</span>
					</div>
				</div>

				<!-- 第三方登录 -->
				<div class="grid grid-cols-2 gap-3">
					<Button
						variant="outline"
						class="h-11 border-slate-300 dark:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
						disabled={isLoading}
					>
						<div class="flex items-center space-x-2">
							<div class="w-5 h-5 bg-blue-500 rounded"></div>
							<span>GitHub</span>
						</div>
					</Button>
					<Button
						variant="outline"
						class="h-11 border-slate-300 dark:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
						disabled={isLoading}
					>
						<div class="flex items-center space-x-2">
							<div class="w-5 h-5 bg-red-500 rounded"></div>
							<span>Google</span>
						</div>
					</Button>
				</div>
			</CardContent>

			<CardFooter class="pt-6">
				<div class="w-full text-center">
					<p class="text-sm text-slate-600 dark:text-slate-400">
						还没有账号？
						<a 
							href="/register" 
							class="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 font-medium transition-colors"
						>
							立即注册
						</a>
					</p>
				</div>
			</CardFooter>
		</Card>

		<!-- 版权信息 -->
		<div class="text-center mt-8 text-sm text-slate-500 dark:text-slate-400">
			<p>© 2024 Stellar. 保留所有权利。</p>
			<div class="flex items-center justify-center space-x-4 mt-2">
				<a href="/privacy" class="hover:text-slate-700 dark:hover:text-slate-300 transition-colors">
					隐私政策
				</a>
				<span>•</span>
				<a href="/terms" class="hover:text-slate-700 dark:hover:text-slate-300 transition-colors">
					服务条款
				</a>
				<span>•</span>
				<a href="/support" class="hover:text-slate-700 dark:hover:text-slate-300 transition-colors">
					技术支持
				</a>
			</div>
		</div>
	</div>
</div>

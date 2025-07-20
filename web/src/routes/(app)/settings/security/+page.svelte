<!--
  @component
  Security Settings page - Password change and security options
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { Badge } from '$lib/components/ui/badge';
	import Icon from '@iconify/svelte';
	import { authApi } from '$lib/api/auth';
	import { auth } from '$lib/stores/auth';
	import { notifications } from '$lib/stores/notifications';

	// 表单数据
	let oldPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let showPassword = $state(false);
	let showNewPassword = $state(false);
	let showConfirmPassword = $state(false);

	// 获取当前用户信息
	let user = $state(auth.state.user);

	// 密码强度检查
	let passwordStrength = $state({
		score: 0,
		feedback: '',
		label: '弱',
		color: 'text-red-500'
	});

	// 检查密码强度
	function checkPasswordStrength(password: string) {
		let score = 0;
		let feedback = [];

		if (password.length >= 8) score += 1;
		else feedback.push('至少8个字符');

		if (/[a-z]/.test(password)) score += 1;
		else feedback.push('包含小写字母');

		if (/[A-Z]/.test(password)) score += 1;
		else feedback.push('包含大写字母');

		if (/[0-9]/.test(password)) score += 1;
		else feedback.push('包含数字');

		if (/[^A-Za-z0-9]/.test(password)) score += 1;
		else feedback.push('包含特殊字符');

		let label, color;
		if (score <= 2) {
			label = '弱';
			color = 'text-red-500';
		} else if (score <= 3) {
			label = '中等';
			color = 'text-yellow-500';
		} else if (score <= 4) {
			label = '强';
			color = 'text-blue-500';
		} else {
			label = '很强';
			color = 'text-green-500';
		}

		return {
			score,
			feedback: feedback.join('、'),
			label,
			color
		};
	}

	// 监听新密码变化
	$effect(() => {
		if (newPassword) {
			passwordStrength = checkPasswordStrength(newPassword);
		}
	});

	// 验证表单
	function validateForm() {
		if (!oldPassword) {
			notifications.add({ type: 'error', message: '请输入当前密码' });
			return false;
		}
		if (!newPassword) {
			notifications.add({ type: 'error', message: '请输入新密码' });
			return false;
		}
		if (newPassword.length < 6) {
			notifications.add({ type: 'error', message: '新密码长度不能少于6位' });
			return false;
		}
		if (newPassword === oldPassword) {
			notifications.add({ type: 'error', message: '新密码不能与当前密码相同' });
			return false;
		}
		if (newPassword !== confirmPassword) {
			notifications.add({ type: 'error', message: '两次输入的新密码不一致' });
			return false;
		}
		return true;
	}

	// 修改密码
	async function handleChangePassword() {
		if (!validateForm()) return;

		isLoading = true;
		try {
			await authApi.changePassword({
				oldPassword,
				newPassword
			});

			notifications.add({ type: 'success', message: '密码修改成功' });
			
			// 清空表单
			oldPassword = '';
			newPassword = '';
			confirmPassword = '';
		} catch (error: any) {
			console.error('修改密码失败:', error);
			notifications.add({ type: 'error', message: error.message || '修改密码失败' });
		} finally {
			isLoading = false;
		}
	}

	// 返回设置页面
	function goBack() {
		goto('/settings');
	}
</script>

<div class="container py-10">
	<!-- 页面标题 -->
	<div class="mb-8">
		<div class="flex items-center space-x-4">
			<Button variant="ghost" size="sm" onclick={goBack} class="p-2">
				<Icon icon="tabler:arrow-left" width={20} />
			</Button>
			<div>
				<h1 class="text-3xl font-bold tracking-tight text-slate-900 dark:text-slate-100">
					安全设置
				</h1>
				<p class="text-slate-600 dark:text-slate-400 mt-2">
					管理您的密码和安全选项
				</p>
			</div>
		</div>
	</div>

	<div class="grid gap-8 lg:grid-cols-3">
		<!-- 主要内容 -->
		<div class="lg:col-span-2 space-y-6">
			<!-- 修改密码卡片 -->
			<Card class="border-slate-200 dark:border-slate-700">
				<CardHeader>
					<div class="flex items-center space-x-3">
						<div class="w-10 h-10 rounded-lg bg-red-50 dark:bg-red-900/20 flex items-center justify-center">
							<Icon icon="tabler:lock" width={20} class="text-red-500" />
						</div>
						<div>
							<CardTitle class="text-lg text-slate-900 dark:text-slate-100">
								修改密码
							</CardTitle>
							<CardDescription class="text-slate-600 dark:text-slate-400">
								定期更换密码可以提高账户安全性
							</CardDescription>
						</div>
					</div>
				</CardHeader>
				<CardContent class="space-y-6">
					<!-- 当前密码 -->
					<div class="space-y-2">
						<Label for="oldPassword" class="text-sm font-medium">
							当前密码
						</Label>
						<div class="relative">
							<Input
								id="oldPassword"
								type={showPassword ? 'text' : 'password'}
								placeholder="请输入当前密码"
								bind:value={oldPassword}
								class="pr-10"
							/>
							<Button
								type="button"
								variant="ghost"
								size="sm"
								class="absolute right-0 top-0 h-full px-3 hover:bg-transparent"
								onclick={() => (showPassword = !showPassword)}
							>
								<Icon
									icon={showPassword ? 'tabler:eye-off' : 'tabler:eye'}
									width={16}
									class="text-slate-500"
								/>
							</Button>
						</div>
					</div>

					<!-- 新密码 -->
					<div class="space-y-2">
						<Label for="newPassword" class="text-sm font-medium">
							新密码
						</Label>
						<div class="relative">
							<Input
								id="newPassword"
								type={showNewPassword ? 'text' : 'password'}
								placeholder="请输入新密码"
								bind:value={newPassword}
								class="pr-10"
							/>
							<Button
								type="button"
								variant="ghost"
								size="sm"
								class="absolute right-0 top-0 h-full px-3 hover:bg-transparent"
								onclick={() => (showNewPassword = !showNewPassword)}
							>
								<Icon
									icon={showNewPassword ? 'tabler:eye-off' : 'tabler:eye'}
									width={16}
									class="text-slate-500"
								/>
							</Button>
						</div>
						
						<!-- 密码强度指示器 -->
						{#if newPassword}
							<div class="space-y-2">
								<div class="flex items-center justify-between text-sm">
									<span class="text-slate-600 dark:text-slate-400">密码强度</span>
									<span class={passwordStrength.color}>{passwordStrength.label}</span>
								</div>
								<div class="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2">
									<div
										class="h-2 rounded-full transition-all duration-300 {passwordStrength.score <= 2 ? 'bg-red-500' : passwordStrength.score <= 3 ? 'bg-yellow-500' : passwordStrength.score <= 4 ? 'bg-blue-500' : 'bg-green-500'}"
										style="width: {passwordStrength.score * 20}%"
									></div>
								</div>
								{#if passwordStrength.feedback}
									<p class="text-xs text-slate-500 dark:text-slate-400">
										建议：{passwordStrength.feedback}
									</p>
								{/if}
							</div>
						{/if}
					</div>

					<!-- 确认新密码 -->
					<div class="space-y-2">
						<Label for="confirmPassword" class="text-sm font-medium">
							确认新密码
						</Label>
						<div class="relative">
							<Input
								id="confirmPassword"
								type={showConfirmPassword ? 'text' : 'password'}
								placeholder="请再次输入新密码"
								bind:value={confirmPassword}
								class="pr-10"
							/>
							<Button
								type="button"
								variant="ghost"
								size="sm"
								class="absolute right-0 top-0 h-full px-3 hover:bg-transparent"
								onclick={() => (showConfirmPassword = !showConfirmPassword)}
							>
								<Icon
									icon={showConfirmPassword ? 'tabler:eye-off' : 'tabler:eye'}
									width={16}
									class="text-slate-500"
								/>
							</Button>
						</div>
						
						<!-- 密码匹配提示 -->
						{#if confirmPassword && newPassword !== confirmPassword}
							<p class="text-xs text-red-500">两次输入的密码不一致</p>
						{:else if confirmPassword && newPassword === confirmPassword}
							<p class="text-xs text-green-500">密码匹配</p>
						{/if}
					</div>

					<!-- 提交按钮 -->
					<Button
						onclick={handleChangePassword}
						disabled={isLoading || !oldPassword || !newPassword || !confirmPassword || newPassword !== confirmPassword}
						class="w-full"
					>
						{#if isLoading}
							<Icon icon="tabler:loader-2" width={16} class="animate-spin mr-2" />
							修改中...
						{:else}
							<Icon icon="tabler:check" width={16} class="mr-2" />
							修改密码
						{/if}
					</Button>
				</CardContent>
			</Card>

			<!-- 安全建议 -->
			<Card class="border-slate-200 dark:border-slate-700">
				<CardHeader>
					<CardTitle class="text-lg text-slate-900 dark:text-slate-100">
						安全建议
					</CardTitle>
				</CardHeader>
				<CardContent class="space-y-4">
					<div class="flex items-start space-x-3">
						<Icon icon="tabler:check-circle" width={20} class="text-green-500 mt-0.5" />
						<div>
							<h4 class="font-medium text-slate-900 dark:text-slate-100">使用强密码</h4>
							<p class="text-sm text-slate-600 dark:text-slate-400">
								密码应包含大小写字母、数字和特殊字符，长度至少8位
							</p>
						</div>
					</div>
					<div class="flex items-start space-x-3">
						<Icon icon="tabler:check-circle" width={20} class="text-green-500 mt-0.5" />
						<div>
							<h4 class="font-medium text-slate-900 dark:text-slate-100">定期更换密码</h4>
							<p class="text-sm text-slate-600 dark:text-slate-400">
								建议每3-6个月更换一次密码，提高账户安全性
							</p>
						</div>
					</div>
					<div class="flex items-start space-x-3">
						<Icon icon="tabler:check-circle" width={20} class="text-green-500 mt-0.5" />
						<div>
							<h4 class="font-medium text-slate-900 dark:text-slate-100">不要重复使用密码</h4>
							<p class="text-sm text-slate-600 dark:text-slate-400">
								避免在其他网站使用相同的密码，防止密码泄露
							</p>
						</div>
					</div>
				</CardContent>
			</Card>
		</div>

		<!-- 侧边栏 -->
		<div class="space-y-6">
			<!-- 账户状态 -->
			<Card class="border-slate-200 dark:border-slate-700">
				<CardHeader>
					<CardTitle class="text-lg text-slate-900 dark:text-slate-100">
						账户状态
					</CardTitle>
				</CardHeader>
				<CardContent class="space-y-4">
					<div class="flex items-center justify-between">
						<span class="text-sm text-slate-600 dark:text-slate-400">登录状态</span>
						<Badge variant="success">已登录</Badge>
					</div>
					<div class="flex items-center justify-between">
						<span class="text-sm text-slate-600 dark:text-slate-400">用户名</span>
						<span class="text-sm font-medium text-slate-900 dark:text-slate-100">
							{user?.username || '未知'}
						</span>
					</div>
					<div class="flex items-center justify-between">
						<span class="text-sm text-slate-600 dark:text-slate-400">邮箱</span>
						<span class="text-sm font-medium text-slate-900 dark:text-slate-100">
							{user?.email || '未知'}
						</span>
					</div>
				</CardContent>
			</Card>

			<!-- 快速操作 -->
			<Card class="border-slate-200 dark:border-slate-700">
				<CardHeader>
					<CardTitle class="text-lg text-slate-900 dark:text-slate-100">
						快速操作
					</CardTitle>
				</CardHeader>
				<CardContent class="space-y-3">
					<Button
						variant="outline"
						class="w-full justify-start"
						onclick={() => goto('/settings')}
					>
						<Icon icon="tabler:settings" width={16} class="mr-2" />
						返回设置
					</Button>
					<Button
						variant="outline"
						class="w-full justify-start"
						onclick={() => goto('/dashboard')}
					>
						<Icon icon="tabler:home" width={16} class="mr-2" />
						返回仪表盘
					</Button>
				</CardContent>
			</Card>
		</div>
	</div>
</div> 
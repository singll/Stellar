<!-- 登录页面（Felte+Zod重构） -->
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

	const schema = z.object({
		username: z.string().min(1, '用户名不能为空'),
		password: z.string().min(1, '密码不能为空')
	});

	type LoginForm = z.infer<typeof schema>;

	let isLoading = $state(false);

	const { form, errors, isSubmitting } = createForm<LoginForm>({
		onSubmit: async (values) => {
			isLoading = true;
			try {
				const response = await authApi.login({
					username: values.username,
					password: values.password
				});
				if (response.code === 200 && response.data) {
					// 存储token到localStorage，兼容SSR环境
					if (typeof window !== 'undefined') {
						localStorage.setItem('token', response.data.token);
					}
					auth.login(response.data);
					await goto('/dashboard');
				} else {
					notifications.add({
						type: 'error',
						message: response.message || '登录失败，请检查您的凭据'
					});
				}
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

<Card class="w-full max-w-sm">
	<CardHeader class="space-y-1 text-center">
		<CardTitle class="text-2xl font-bold tracking-tight">登录</CardTitle>
		<CardDescription>欢迎回来！请输入您的凭据。</CardDescription>
	</CardHeader>
	<CardContent class="grid gap-4">
		<form use:form class="grid gap-4">
			<div class="grid gap-2">
				<Label for="username">用户名</Label>
				<Input
					id="username"
					name="username"
					type="text"
					placeholder="请输入用户名"
					disabled={isLoading || $isSubmitting}
				/>
				{#if $errors.username}
					<div class="text-sm text-red-500">{$errors.username}</div>
				{/if}
			</div>
			<div class="grid gap-2">
				<Label for="password">密码</Label>
				<Input
					id="password"
					name="password"
					type="password"
					placeholder="请输入密码"
					disabled={isLoading || $isSubmitting}
				/>
				{#if $errors.password}
					<div class="text-sm text-red-500">{$errors.password}</div>
				{/if}
			</div>
			<Button type="submit" class="w-full" disabled={isLoading || $isSubmitting}>
				{#if isLoading || $isSubmitting}
					<svg
						class="mr-2 h-4 w-4 animate-spin"
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
					>
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"
						></circle>
						<path
							class="opacity-75"
							fill="currentColor"
							d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
						></path>
					</svg>
					处理中...
				{:else}
					登录
				{/if}
			</Button>
		</form>
	</CardContent>
	<CardFooter>
		<div class="w-full text-center text-sm">
			还没有账号？
			<a href="/register" class="underline"> 注册 </a>
		</div>
	</CardFooter>
</Card>

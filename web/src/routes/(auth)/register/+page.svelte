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

	const schema = z
		.object({
			username: z.string().min(2, '用户名至少2位'),
			email: z.string().email('请输入有效邮箱'),
			password: z.string().min(6, '密码至少6位'),
			confirmPassword: z.string(),
			agreeToTerms: z.literal(true, { errorMap: () => ({ message: '请同意服务协议' }) })
		})
		.refine((data) => data.password === data.confirmPassword, {
			message: '两次输入的密码不一致',
			path: ['confirmPassword']
		});

	type RegisterForm = z.infer<typeof schema>;

	let isLoading = $state(false);

	const { form, errors, isSubmitting } = createForm<RegisterForm>({
		onSubmit: async (values) => {
			isLoading = true;
			try {
				const { username, email, password } = values;
				const response = await authApi.register({ username, email, password });
				if (response.code === 200 && response.data) {
					// 注册成功后自动登录
					if (typeof window !== 'undefined') {
						localStorage.setItem('token', response.data.token);
					}
					auth.login(response.data);
					notifications.add({ type: 'success', message: '注册成功，已自动登录' });
					await goto('/');
				} else {
					notifications.add({ type: 'error', message: response.message || '注册失败' });
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
		<CardTitle class="text-2xl font-bold tracking-tight">注册</CardTitle>
		<CardDescription>创建新账号，开启安全之旅。</CardDescription>
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
				<Label for="email">邮箱</Label>
				<Input
					id="email"
					name="email"
					type="email"
					placeholder="请输入邮箱"
					disabled={isLoading || $isSubmitting}
				/>
				{#if $errors.email}
					<div class="text-sm text-red-500">{$errors.email}</div>
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
			<div class="grid gap-2">
				<Label for="confirmPassword">确认密码</Label>
				<Input
					id="confirmPassword"
					name="confirmPassword"
					type="password"
					placeholder="请再次输入密码"
					disabled={isLoading || $isSubmitting}
				/>
				{#if $errors.confirmPassword}
					<div class="text-sm text-red-500">{$errors.confirmPassword}</div>
				{/if}
			</div>
			<div class="flex items-center gap-2">
				<input
					id="agreeToTerms"
					name="agreeToTerms"
					type="checkbox"
					disabled={isLoading || $isSubmitting}
				/>
				<Label for="agreeToTerms"
					>我已阅读并同意 <a href="/terms" class="underline">服务协议</a></Label
				>
			</div>
			{#if $errors.agreeToTerms}
				<div class="text-sm text-red-500">{$errors.agreeToTerms}</div>
			{/if}
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
					注册中...
				{:else}
					注册
				{/if}
			</Button>
		</form>
	</CardContent>
	<CardFooter>
		<div class="w-full text-center text-sm">
			已有账号？
			<a href="/login" class="underline"> 登录 </a>
		</div>
	</CardFooter>
</Card>

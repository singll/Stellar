<script lang="ts">
	import { goto } from '$app/navigation';
	import { ProjectAPI } from '$lib/api/projects';
	import type { CreateProjectRequest } from '$lib/types/project';
	import { PROJECT_COLORS } from '$lib/types/project';
	import { notifications } from '$lib/stores/notifications';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import Icon from '$lib/components/ui/Icon.svelte';

	// 表单状态
	let formData: CreateProjectRequest = $state({
		name: '',
		description: '',
		target: '',
		color: 'blue',
		is_private: false
	});

	let loading = $state(false);
	let errors = $state<Record<string, string>>({});

	// 表单验证
	const validateForm = (): boolean => {
		const newErrors: Record<string, string> = {};

		if (!formData.name.trim()) {
			newErrors.name = '项目名称不能为空';
		} else if (formData.name.length < 2) {
			newErrors.name = '项目名称至少需要2个字符';
		} else if (formData.name.length > 50) {
			newErrors.name = '项目名称不能超过50个字符';
		}

		if (formData.description && formData.description.length > 500) {
			newErrors.description = '项目描述不能超过500个字符';
		}

		if (formData.target) {
			// 验证目标格式（域名或IP）
			const targetPattern =
				/^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$|^(\d{1,3}\.){3}\d{1,3}$|^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}\/.*$/;
			if (!targetPattern.test(formData.target.trim())) {
				newErrors.target = '请输入有效的域名或IP地址';
			}
		}

		errors = newErrors;
		return Object.keys(newErrors).length === 0;
	};

	// 提交表单
	const handleSubmit = async (event: Event) => {
		event.preventDefault();

		if (!validateForm()) {
			return;
		}

		loading = true;

		try {
			const response = await ProjectAPI.createProject({
				...formData,
				name: formData.name.trim(),
				description: formData.description?.trim() || undefined,
				target: formData.target?.trim() || undefined
			});

			notifications.add({
				type: 'success',
				message: '项目创建成功'
			});

			// 跳转到项目详情页
			await goto(`/projects/${response.id}`);
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '创建项目失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		} finally {
			loading = false;
		}
	};

	// 测试目标连通性
	const handleTestTarget = async () => {
		if (!formData.target?.trim()) {
			notifications.add({
				type: 'warning',
				message: '请先输入目标地址'
			});
			return;
		}

		notifications.add({
			type: 'info',
			message: '目标连通性测试功能正在开发中'
		});
	};

	// 颜色选择器
	const colorOptions = PROJECT_COLORS.map((color) => ({
		value: color,
		label: color,
		class: `bg-${color}-500`
	}));

	// 实时验证
	$effect(() => {
		if (formData.name) {
			validateForm();
		}
	});
</script>

<svelte:head>
	<title>创建项目 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6 max-w-2xl">
	<!-- 页面标题 -->
	<div class="mb-6">
		<div class="flex items-center gap-4 mb-4">
			<Button variant="ghost" onclick={() => goto('/projects')} class="flex items-center gap-2">
				<Icon name="chevron-left" class="h-4 w-4" />
				返回项目列表
			</Button>
		</div>

		<h1 class="text-3xl font-bold text-gray-900">创建新项目</h1>
		<p class="text-gray-600 mt-1">设置项目基本信息并开始安全扫描</p>
	</div>

	<!-- 创建表单 -->
	<Card>
		<CardHeader>
			<CardTitle>项目信息</CardTitle>
			<CardDescription>填写项目的基本信息。标有 * 的字段为必填项。</CardDescription>
		</CardHeader>

		<CardContent>
			<form onsubmit={handleSubmit} class="space-y-6">
				<!-- 项目名称 -->
				<div class="space-y-2">
					<Label for="name" class="text-sm font-medium">
						项目名称 <span class="text-red-500">*</span>
					</Label>
					<Input
						id="name"
						type="text"
						bind:value={formData.name}
						placeholder="输入项目名称"
						class={errors.name ? 'border-red-500 focus:border-red-500' : ''}
						disabled={loading}
						required
					/>
					{#if errors.name}
						<p class="text-sm text-red-600">{errors.name}</p>
					{/if}
				</div>

				<!-- 项目描述 -->
				<div class="space-y-2">
					<Label for="description" class="text-sm font-medium">项目描述</Label>
					<textarea
						id="description"
						bind:value={formData.description}
						placeholder="描述项目的目的、范围或特殊要求..."
						rows={4}
						class={`w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${errors.description ? 'border-red-500 focus:border-red-500' : ''}`}
						disabled={loading}
					></textarea>
					{#if errors.description}
						<p class="text-sm text-red-600">{errors.description}</p>
					{/if}
					<p class="text-xs text-gray-500">
						{formData.description?.length || 0}/500
					</p>
				</div>

				<!-- 目标地址 -->
				<div class="space-y-2">
					<Label for="target" class="text-sm font-medium">目标地址</Label>
					<div class="flex gap-2">
						<Input
							id="target"
							type="text"
							bind:value={formData.target}
							placeholder="example.com 或 192.168.1.1"
							class={`flex-1 ${errors.target ? 'border-red-500 focus:border-red-500' : ''}`}
							disabled={loading}
						/>
						<Button
							type="button"
							variant="outline"
							onclick={handleTestTarget}
							disabled={loading || !formData.target?.trim()}
							class="flex items-center gap-2"
						>
							<Icon name="check" class="h-4 w-4" />
							测试
						</Button>
					</div>
					{#if errors.target}
						<p class="text-sm text-red-600">{errors.target}</p>
					{:else}
						<p class="text-xs text-gray-500">输入要扫描的目标域名或IP地址</p>
					{/if}
				</div>

				<!-- 项目颜色 -->
				<div class="space-y-2">
					<Label for="color" class="text-sm font-medium">项目颜色</Label>
					<select
						id="color"
						bind:value={formData.color}
						disabled={loading}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					>
						{#each colorOptions as option}
							<option value={option.value}>
								{option.label}
							</option>
						{/each}
					</select>
					<p class="text-xs text-gray-500">选择一个颜色来区分不同的项目</p>
				</div>

				<!-- 项目可见性 -->
				<div class="space-y-2">
					<div class="flex items-center space-x-2">
						<Checkbox id="is_private" bind:checked={formData.is_private} disabled={loading} />
						<Label for="is_private" class="text-sm font-medium">私有项目</Label>
					</div>
					<p class="text-xs text-gray-500 ml-6">私有项目只有项目成员才能查看和访问</p>
				</div>

				<!-- 提交按钮 -->
				<div class="flex items-center gap-4 pt-4 border-t">
					<button
						type="submit"
						disabled={loading || !formData.name.trim()}
						class="inline-flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						{#if loading}
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
							创建中...
						{:else}
							<Icon name="check" class="h-4 w-4" />
							创建项目
						{/if}
					</button>

					<button
						type="button"
						onclick={() => goto('/projects')}
						disabled={loading}
						class="inline-flex items-center gap-2 border border-gray-300 bg-white text-gray-700 px-4 py-2 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						取消
					</button>
				</div>
			</form>
		</CardContent>
	</Card>

	<!-- 帮助信息 -->
	<Card class="mt-6">
		<CardHeader>
			<CardTitle class="text-lg">创建提示</CardTitle>
		</CardHeader>
		<CardContent class="space-y-3 text-sm text-gray-600">
			<div>
				<strong>项目名称:</strong> 建议使用有意义的名称，便于团队成员识别
			</div>
			<div>
				<strong>目标地址:</strong> 支持域名（example.com）、IP地址（192.168.1.1）或子域名（sub.example.com）
			</div>
			<div>
				<strong>项目颜色:</strong> 用于在界面上快速区分不同项目
			</div>
			<div>
				<strong>私有项目:</strong> 只有被邀请的成员才能查看项目内容
			</div>
		</CardContent>
	</Card>
</div>

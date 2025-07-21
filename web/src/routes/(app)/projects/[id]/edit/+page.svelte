<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { ProjectAPI } from '$lib/api/projects';
	import type { Project, UpdateProjectRequest } from '$lib/types/project';
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
	import Icon from '@iconify/svelte';
	import { onMount } from 'svelte';

	let { data } = $props();
	let project = $state(data.project);
	let loading = $state(false);
	let errors = $state<Record<string, string>>({});

	// 表单数据
	let formData: UpdateProjectRequest = $state({
		name: '',
		description: '',
		target: '',
		color: '',
		is_private: false,
		status: 'active',
		tags: []
	});

	// 初始化表单数据
	$effect(() => {
		formData = {
			name: project.name || '',
			description: project.description || '',
			target: project.target || '',
			color: project.color || '',
			is_private: project.is_private || false,
			status: project.status || 'active',
			tags: project.tags || []
		};
	});

	// 表单验证
	const validateForm = (): boolean => {
		const newErrors: Record<string, string> = {};

		if (!formData.name?.trim()) {
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
			const targetPattern = /^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$|^(\d{1,3}\.){3}\d{1,3}$|^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}\/.*$/;
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
			const updated = await ProjectAPI.updateProject(project.id, {
				...formData,
				name: formData.name?.trim(),
				description: formData.description?.trim() || undefined,
				target: formData.target?.trim() || undefined
			});

			project = updated;
			notifications.add({
				type: 'success',
				message: '项目更新成功'
			});

			await goto(`/projects/${project.id}`);
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '更新项目失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		} finally {
			loading = false;
		}
	};

	// 删除项目
	const handleDelete = async () => {
		if (!confirm(`确定要删除项目 "${project.name}" 吗？此操作不可恢复。`)) return;

		try {
			await ProjectAPI.deleteProject(project.id);
			notifications.add({
				type: 'success',
				message: '项目删除成功'
			});
			await goto('/projects');
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '删除项目失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		}
	};

	// 颜色选项
	const colorOptions = PROJECT_COLORS.map((color) => ({
		value: color,
		label: color
	}));

	// 状态选项
	const statusOptions = [
		{ value: 'active', label: '活跃' },
		{ value: 'paused', label: '暂停' },
		{ value: 'completed', label: '已完成' },
		{ value: 'archived', label: '已归档' }
	];

	// 标签管理
	let newTag = $state('');

	const addTag = () => {
		if (newTag.trim() && !formData.tags?.includes(newTag.trim())) {
			formData.tags = [...(formData.tags || []), newTag.trim()];
			newTag = '';
		}
	};

	const removeTag = (tagToRemove: string) => {
		formData.tags = formData.tags?.filter(tag => tag !== tagToRemove);
	};

	// 实时验证
	$effect(() => {
		if (formData.name) {
			validateForm();
		}
	});
</script>

<svelte:head>
	<title>编辑项目 - {project.name} - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6 max-w-2xl">
	<!-- 页面标题 -->
	<div class="mb-6">
		<div class="flex items-center gap-4 mb-4">
			<Button variant="ghost" onclick={() => goto(`/projects/${project.id}`)} class="flex items-center gap-2">
				<Icon icon="tabler:chevron-left" class="h-4 w-4" />
				返回项目
			</Button>
		</div>

		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-3xl font-bold text-gray-900">编辑项目</h1>
				<p class="text-gray-600 mt-1">修改 "{project.name}" 项目信息</p>
			</div>
			<Button variant="destructive" onclick={handleDelete}>
				<Icon icon="tabler:trash" class="h-4 w-4 mr-1" />
				删除
			</Button>
		</div>
	</div>

	<!-- 编辑表单 -->
	<Card>
		<CardHeader>
			<CardTitle>项目信息</CardTitle>
			<CardDescription>更新项目的基本信息</CardDescription>
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
					<p class="text-xs text-gray-500">{(formData.description?.length || 0)}/500</p>
				</div>

				<!-- 目标地址 -->
				<div class="space-y-2">
					<Label for="target" class="text-sm font-medium">目标地址</Label>
					<Input
						id="target"
						type="text"
						bind:value={formData.target}
						placeholder="example.com 或 192.168.1.1"
						class={errors.target ? 'border-red-500 focus:border-red-500' : ''}
						disabled={loading}
					/>
					{#if errors.target}
						<p class="text-sm text-red-600">{errors.target}</p>
					{:else}
						<p class="text-xs text-gray-500">输入要扫描的目标域名或IP地址</p>
					{/if}
				</div>

				<!-- 项目状态 -->
				<div class="space-y-2">
					<Label for="status" class="text-sm font-medium">项目状态</Label>
					<select
						id="status"
						bind:value={formData.status}
						disabled={loading}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					>
						{#each statusOptions as option}
							<option value={option.value}>{option.label}</option>
						{/each}
					</select>
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
							<option value={option.value}>{option.label}</option>
						{/each}
					</select>
				</div>

				<!-- 项目标签 -->
				<div class="space-y-2">
					<Label for="tags" class="text-sm font-medium">项目标签</Label>
					<div class="flex gap-2">
						<Input
							id="tags"
							type="text"
							bind:value={newTag}
							placeholder="输入标签后按回车添加"
							class="flex-1"
							disabled={loading}
							onkeydown={(e) => {
								if (e.key === 'Enter') {
									e.preventDefault();
									addTag();
								}
							}}
						/>
						<Button type="button" onclick={addTag} disabled={loading || !newTag.trim()}>
							添加
						</Button>
					</div>
					<div class="flex flex-wrap gap-2 mt-2">
						{#each formData.tags || [] as tag}
							<span class="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-800 rounded-md text-sm">
								{tag}
								<button
									type="button"
									onclick={() => removeTag(tag)}
									class="text-blue-600 hover:text-blue-800"
								>
									<Icon icon="tabler:x" class="h-3 w-3" />
								</button>
							</span>
						{/each}
					</div>
				</div>

				<!-- 项目可见性 -->
				<div class="space-y-2">
					<div class="flex items-center space-x-2">
						<Checkbox
							id="is_private"
							bind:checked={formData.is_private}
							disabled={loading}
						/>
						<Label for="is_private" class="text-sm font-medium">私有项目</Label>
					</div>
					<p class="text-xs text-gray-500 ml-6">私有项目只有项目成员才能查看和访问</p>
				</div>

				<!-- 提交按钮 -->
				<div class="flex items-center gap-4 pt-4 border-t">
					<button
						type="submit"
						disabled={loading || !formData.name?.trim()}
						class="inline-flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						{#if loading}
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
							更新中...
						{:else}
							<Icon icon="tabler:check" class="h-4 w-4" />
							保存更改
						{/if}
					</button>

					<button
						type="button"
						onclick={() => goto(`/projects/${project.id}`)}
						disabled={loading}
						class="inline-flex items-center gap-2 border border-gray-300 bg-white text-gray-700 px-4 py-2 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						取消
					</button>
				</div>
			</form>
		</CardContent>
	</Card>
</div>
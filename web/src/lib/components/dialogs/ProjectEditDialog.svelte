<!-- 项目编辑对话框 -->
<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { DialogContent, DialogHeader, DialogTitle, DialogFooter } from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import Icon from '$lib/components/ui/Icon.svelte';
	import type { Project, UpdateProjectRequest, PROJECT_COLORS } from '$lib/types/project';
	
	interface Props {
		open: boolean;
		project: Project | null;
		loading?: boolean;
		onSave: (data: UpdateProjectRequest) => Promise<void>;
		onCancel: () => void;
	}

	let {
		open = $bindable(),
		project,
		loading = false,
		onSave,
		onCancel
	}: Props = $props();

	// 表单数据
	let formData = $state({
		name: '',
		description: '',
		target: '',
		color: 'blue' as const,
		is_private: false
	});

	// 表单验证错误
	let errors = $state({
		name: '',
		target: ''
	});

	// 可用的颜色选项
	const colorOptions = [
		{ value: 'blue', label: '蓝色', class: 'bg-blue-500' },
		{ value: 'green', label: '绿色', class: 'bg-green-500' },
		{ value: 'red', label: '红色', class: 'bg-red-500' },
		{ value: 'yellow', label: '黄色', class: 'bg-yellow-500' },
		{ value: 'purple', label: '紫色', class: 'bg-purple-500' },
		{ value: 'pink', label: '粉色', class: 'bg-pink-500' },
		{ value: 'indigo', label: '靛蓝', class: 'bg-indigo-500' },
		{ value: 'gray', label: '灰色', class: 'bg-gray-500' }
	];

	// 监听项目数据变化，自动填充表单
	$effect(() => {
		if (project && open) {
			formData = {
				name: project.name || '',
				description: project.description || '',
				target: project.target || '',
				color: project.color as any || 'blue',
				is_private: project.is_private || false
			};
			// 清除错误
			errors = { name: '', target: '' };
		}
	});

	// 表单验证
	function validateForm(): boolean {
		let isValid = true;
		errors = { name: '', target: '' };

		if (!formData.name.trim()) {
			errors.name = '项目名称不能为空';
			isValid = false;
		} else if (formData.name.length > 100) {
			errors.name = '项目名称不能超过100个字符';
			isValid = false;
		}

		if (formData.target && formData.target.length > 500) {
			errors.target = '目标不能超过500个字符';
			isValid = false;
		}

		return isValid;
	}

	// 提交表单
	async function handleSave() {
		if (!validateForm()) {
			return;
		}

		try {
			await onSave({
				name: formData.name.trim(),
				description: formData.description.trim() || undefined,
				target: formData.target.trim() || undefined,
				color: formData.color,
				is_private: formData.is_private
			});
			open = false;
		} catch (error) {
			// 错误处理已在父组件中处理
		}
	}

	// 取消
	function handleCancel() {
		open = false;
		onCancel();
	}

	// 监听键盘事件
	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
			handleSave();
		}
	}
</script>

<DialogPrimitive.Root bind:open>
	<DialogContent class="sm:max-w-lg" onkeydown={handleKeydown}>
		<DialogHeader>
			<div class="flex items-center gap-3">
				<div class="flex items-center justify-center w-10 h-10 rounded-full bg-blue-100">
					<Icon name="edit" class="h-5 w-5 text-blue-600" />
				</div>
				<DialogTitle>
					{project ? '编辑项目' : '新建项目'}
				</DialogTitle>
			</div>
		</DialogHeader>

		<div class="space-y-6 py-4">
			<!-- 项目名称 -->
			<div class="space-y-2">
				<Label for="project-name" class="text-sm font-medium">
					项目名称 <span class="text-red-500">*</span>
				</Label>
				<Input
					id="project-name"
					bind:value={formData.name}
					placeholder="请输入项目名称"
					class={errors.name ? 'border-red-500' : ''}
					disabled={loading}
				/>
				{#if errors.name}
					<p class="text-sm text-red-500">{errors.name}</p>
				{/if}
			</div>

			<!-- 项目描述 -->
			<div class="space-y-2">
				<Label for="project-description" class="text-sm font-medium">
					项目描述
				</Label>
				<Textarea
					id="project-description"
					bind:value={formData.description}
					placeholder="请输入项目描述（可选）"
					rows={3}
					disabled={loading}
				/>
			</div>

			<!-- 扫描目标 -->
			<div class="space-y-2">
				<Label for="project-target" class="text-sm font-medium">
					扫描目标
				</Label>
				<Input
					id="project-target"
					bind:value={formData.target}
					placeholder="请输入扫描目标，如域名、IP等"
					class={errors.target ? 'border-red-500' : ''}
					disabled={loading}
				/>
				{#if errors.target}
					<p class="text-sm text-red-500">{errors.target}</p>
				{/if}
			</div>

			<!-- 项目颜色 -->
			<div class="space-y-2">
				<Label class="text-sm font-medium">项目颜色</Label>
				<div class="flex flex-wrap gap-2">
					{#each colorOptions as color}
						<button
							type="button"
							class="flex items-center gap-2 px-3 py-2 text-sm border rounded-md transition-colors hover:bg-gray-50 {formData.color === color.value ? 'border-blue-500 bg-blue-50' : 'border-gray-300'}"
							onclick={() => formData.color = color.value as any}
							disabled={loading}
						>
							<div class="w-4 h-4 rounded-full {color.class}"></div>
							{color.label}
						</button>
					{/each}
				</div>
			</div>

			<!-- 隐私设置 -->
			<div class="space-y-2">
				<div class="flex items-center gap-2">
					<input
						id="project-private"
						type="checkbox"
						bind:checked={formData.is_private}
						class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
						disabled={loading}
					/>
					<Label for="project-private" class="text-sm font-medium cursor-pointer">
						私有项目
					</Label>
				</div>
				<p class="text-xs text-gray-500">私有项目仅对项目成员可见</p>
			</div>
		</div>

		<DialogFooter>
			<div class="flex justify-end gap-3 w-full">
				<Button
					variant="outline"
					onclick={handleCancel}
					disabled={loading}
				>
					取消
				</Button>
				<Button
					onclick={handleSave}
					disabled={loading}
					class="min-w-[80px]"
				>
					{#if loading}
						<div class="flex items-center gap-2">
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
							保存中...
						</div>
					{:else}
						<div class="flex items-center gap-2">
							<Icon name="check" class="h-4 w-4" />
							保存
						</div>
					{/if}
				</Button>
			</div>
			
			<!-- 快捷键提示 -->
			<div class="text-xs text-gray-500 text-center pt-2 border-t border-gray-100">
				按 Ctrl+Enter 快速保存
			</div>
		</DialogFooter>
	</DialogContent>
</DialogPrimitive.Root>
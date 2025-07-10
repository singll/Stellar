<!--
创建任务页面
提供创建新任务的表单界面
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { taskActions } from '$lib/stores/tasks';
	import { projectStore, projectActions } from '$lib/stores/projects';
	import type { TaskType, TaskPriority, CreateTaskRequest } from '$lib/types/task';

	import TaskTypeSelector from '$lib/components/tasks/TaskTypeSelector.svelte';
	import TaskConfigEditor from '$lib/components/tasks/TaskConfigEditor.svelte';
	import ProjectSelector from '$lib/components/projects/ProjectSelector.svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import TagInput from '$lib/components/ui/TagInput.svelte';
	import FormField from '$lib/components/ui/FormField.svelte';

	// 表单状态
	let name = $state('');
	let description = $state('');
	let type = $state<TaskType>('subdomain_enum');
	let priority = $state<TaskPriority>('normal');
	let projectId = $state('');
	let config = $state<Record<string, any>>({});
	let timeout = $state(3600);
	let maxRetries = $state(3);
	let tags = $state<string[]>([]);
	let scheduledAt = $state('');

	// 表单验证和提交状态
	let errors = $state<Record<string, string>>({});
	let isSubmitting = $state(false);
	let isValidating = $state(false);

	// Store 订阅
	let projects = $state();
	projectStore.subscribe((value) => {
		projects = value;
	});

	// 任务类型选项
	const taskTypes = [
		{ value: 'subdomain_enum', label: '子域名枚举', icon: 'fas fa-globe' },
		{ value: 'port_scan', label: '端口扫描', icon: 'fas fa-network-wired' },
		{ value: 'vuln_scan', label: '漏洞扫描', icon: 'fas fa-bug' },
		{ value: 'asset_discovery', label: '资产发现', icon: 'fas fa-search' },
		{ value: 'dir_scan', label: '目录扫描', icon: 'fas fa-folder' },
		{ value: 'web_crawl', label: 'Web爬虫', icon: 'fas fa-spider' }
	];

	// 优先级选项
	const priorityOptions = [
		{ value: 'low', label: '低' },
		{ value: 'normal', label: '正常' },
		{ value: 'high', label: '高' },
		{ value: 'critical', label: '紧急' }
	];

	onMount(async () => {
		// 加载项目列表
		await projectActions.loadProjects();

		// 如果只有一个项目，自动选择
		if (projects?.projects?.length === 1) {
			projectId = projects.projects[0].id;
		}

		// 初始化默认配置
		updateTaskConfig();
	});

	// 更新任务配置
	async function updateTaskConfig() {
		try {
			// 获取任务类型的配置模板
			const template = await taskActions.getTaskConfigTemplate(type);
			if (template) {
				config = template;
			}
		} catch (error) {
			console.error('获取配置模板失败:', error);
		}
	}

	// 验证表单
	async function validateForm(): Promise<boolean> {
		errors = {};

		// 基础字段验证
		if (!name.trim()) {
			errors.name = '任务名称不能为空';
		}

		if (!projectId) {
			errors.projectId = '请选择项目';
		}

		if (timeout <= 0) {
			errors.timeout = '超时时间必须大于0';
		}

		if (maxRetries < 0) {
			errors.maxRetries = '重试次数不能为负数';
		}

		// 验证任务配置
		if (Object.keys(config).length > 0) {
			isValidating = true;
			try {
				const validation = await taskActions.validateTaskConfig(type, config);
				if (!validation.valid) {
					validation.errors.forEach((error) => {
						errors[`config.${error.field}`] = error.message;
					});
				}
			} catch (error) {
				errors.config = '配置验证失败';
			} finally {
				isValidating = false;
			}
		}

		return Object.keys(errors).length === 0;
	}

	// 提交表单
	async function handleSubmit() {
		if (isSubmitting) return;

		const isValid = await validateForm();
		if (!isValid) return;

		isSubmitting = true;

		try {
			const taskData: CreateTaskRequest = {
				name: name.trim(),
				description: description.trim() || undefined,
				type,
				priority,
				projectId,
				config,
				timeout,
				maxRetries,
				tags,
				scheduledAt: scheduledAt || undefined
			};

			const createdTask = await taskActions.createTask(taskData);

			if (createdTask) {
				// 跳转到任务详情页
				goto(`/tasks/${createdTask.id}`);
			}
		} catch (error) {
			console.error('创建任务失败:', error);
		} finally {
			isSubmitting = false;
		}
	}

	// 从模板创建
	function createFromTemplate() {
		goto('/tasks/templates');
	}

	// 返回列表
	function goBack() {
		goto('/tasks');
	}

	// 预览执行计划
	async function previewExecutionPlan() {
		const taskData: CreateTaskRequest = {
			name: name.trim(),
			description: description.trim() || undefined,
			type,
			priority,
			projectId,
			config,
			timeout,
			maxRetries,
			tags
		};

		try {
			const plan = await taskActions.generateTaskExecutionPlan(taskData);
			// TODO: 显示执行计划预览对话框
			console.log('执行计划:', plan);
		} catch (error) {
			console.error('生成执行计划失败:', error);
		}
	}

	// 当任务类型改变时更新配置
	$effect(() => {
		updateTaskConfig();
	});
</script>

<svelte:head>
	<title>创建任务 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6 max-w-4xl">
	<!-- 页面标题 -->
	<div class="flex items-center justify-between mb-6">
		<div class="flex items-center gap-4">
			<Button variant="ghost" onclick={goBack}>
				<i class="fas fa-arrow-left mr-2"></i>
				返回
			</Button>
			<div>
				<h1 class="text-2xl font-bold text-gray-900 dark:text-white">创建任务</h1>
				<p class="text-gray-600 dark:text-gray-400 mt-1">配置并创建新的安全扫描任务</p>
			</div>
		</div>
		<Button variant="outline" onclick={createFromTemplate}>
			<i class="fas fa-template mr-2"></i>
			从模板创建
		</Button>
	</div>

	<form
		onsubmit={(e) => {
			e.preventDefault();
			handleSubmit();
		}}
		class="space-y-6"
	>
		<!-- 基础信息 -->
		<div
			class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
		>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">基础信息</h2>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<FormField label="任务名称" required error={errors.name}>
					<Input bind:value={name} placeholder="输入任务名称" disabled={isSubmitting} />
				</FormField>

				<FormField label="项目" required error={errors.projectId}>
					<ProjectSelector
						bind:value={projectId}
						projects={projects?.projects || []}
						disabled={isSubmitting}
					/>
				</FormField>
			</div>

			<FormField label="描述" error={errors.description}>
				<Textarea
					bind:value={description}
					placeholder="输入任务描述（可选）"
					rows={3}
					disabled={isSubmitting}
				/>
			</FormField>
		</div>

		<!-- 任务类型和配置 -->
		<div
			class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
		>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">任务类型和配置</h2>

			<FormField label="任务类型" required>
				<TaskTypeSelector bind:value={type} options={taskTypes} disabled={isSubmitting} />
			</FormField>

			{#if isValidating}
				<div class="flex items-center justify-center py-4">
					<LoadingSpinner />
					<span class="ml-2 text-gray-600 dark:text-gray-400">验证配置...</span>
				</div>
			{:else}
				<FormField label="任务配置" error={errors.config}>
					<TaskConfigEditor bind:config taskType={type} disabled={isSubmitting} {errors} />
				</FormField>
			{/if}
		</div>

		<!-- 执行选项 -->
		<div
			class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
		>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">执行选项</h2>

			<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
				<FormField label="优先级" error={errors.priority}>
					<Select bind:value={priority} options={priorityOptions} disabled={isSubmitting} />
				</FormField>

				<FormField label="超时时间（秒）" error={errors.timeout}>
					<Input type="number" bind:value={timeout} min="1" step="1" disabled={isSubmitting} />
				</FormField>

				<FormField label="最大重试次数" error={errors.maxRetries}>
					<Input type="number" bind:value={maxRetries} min="0" step="1" disabled={isSubmitting} />
				</FormField>
			</div>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<FormField label="标签" error={errors.tags}>
					<TagInput bind:tags placeholder="添加标签" disabled={isSubmitting} />
				</FormField>

				<FormField label="计划执行时间（可选）" error={errors.scheduledAt}>
					<Input type="datetime-local" bind:value={scheduledAt} disabled={isSubmitting} />
				</FormField>
			</div>
		</div>

		<!-- 操作按钮 -->
		<div
			class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
		>
			<div class="flex flex-col sm:flex-row gap-4 justify-between">
				<div class="flex gap-2">
					<Button variant="outline" onclick={previewExecutionPlan} disabled={isSubmitting}>
						<i class="fas fa-eye mr-2"></i>
						预览执行计划
					</Button>
				</div>

				<div class="flex gap-2">
					<Button variant="outline" onclick={goBack} disabled={isSubmitting}>取消</Button>
					<Button type="submit" disabled={isSubmitting || isValidating}>
						{#if isSubmitting}
							<LoadingSpinner size="sm" />
							<span class="ml-2">创建中...</span>
						{:else}
							<i class="fas fa-plus mr-2"></i>
							创建任务
						{/if}
					</Button>
				</div>
			</div>
		</div>
	</form>
</div>

<!-- 创建端口扫描任务页面 -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { toastStore } from '$lib/stores/toast';
	import { portScanStore } from '$lib/stores/portscan';
	import { portScanApi } from '$lib/api/portscan';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import FormField from '$lib/components/ui/FormField.svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import TagInput from '$lib/components/ui/TagInput.svelte';
	import type { PortScanFormData } from '$lib/types/portscan';

	// 表单数据
	let formData = $state<PortScanFormData>({
		name: '',
		description: '',
		target: '',
		ports: '',
		preset: 'custom',
		scanMethod: 'tcp',
		advanced: {
			maxWorkers: 50,
			timeout: 30,
			enableBanner: false,
			enableSSL: false,
			enableService: true,
			rateLimit: 100
		},
		projectId: undefined
	});

	// 表单状态
	let isSubmitting = $state(false);
	let showAdvanced = $state(false);
	let validationErrors = $state<Record<string, string>>({});

	// 端口预设选项
	const portPresets = portScanApi.getPresetPorts();
	const presetOptions = [
		{ value: 'custom', label: '自定义' },
		...Object.entries(portPresets).map(([key, preset]) => ({
			value: key,
			label: preset.name
		}))
	];

	// 扫描方法选项
	const scanMethodOptions = [
		{ value: 'tcp', label: 'TCP' },
		{ value: 'udp', label: 'UDP' },
		{ value: 'both', label: 'TCP + UDP' }
	];

	// 项目选项（模拟数据）
	const projectOptions = [
		{ value: '', label: '不关联项目' },
		{ value: 'proj1', label: '项目 1' },
		{ value: 'proj2', label: '项目 2' }
	];

	// 处理预设选择
	function handlePresetChange(preset: string) {
		if (preset !== 'custom' && portPresets[preset]) {
			formData.ports = portPresets[preset].ports;
		}
	}

	// 验证表单
	function validateForm(): boolean {
		const errors: Record<string, string> = {};

		// 验证任务名称
		if (!formData.name.trim()) {
			errors.name = '任务名称不能为空';
		}

		// 验证目标
		if (!formData.target.trim()) {
			errors.target = '扫描目标不能为空';
		} else {
			// 简单的IP/域名验证
			const targetRegex =
				/^([a-zA-Z0-9.-]+\.[a-zA-Z]{2,}|(\d{1,3}\.){3}\d{1,3}|(\d{1,3}\.){3}\d{1,3}\/\d{1,2})$/;
			if (!targetRegex.test(formData.target.trim())) {
				errors.target = '请输入有效的域名或IP地址';
			}
		}

		// 验证端口配置
		if (!formData.ports.trim()) {
			errors.ports = '端口配置不能为空';
		} else {
			const validation = portScanApi.validatePorts(formData.ports);
			if (!validation.valid) {
				errors.ports = validation.message || '端口配置格式错误';
			}
		}

		// 验证高级设置
		if (formData.advanced.maxWorkers < 1 || formData.advanced.maxWorkers > 1000) {
			errors.maxWorkers = '并发数必须在1-1000之间';
		}

		if (formData.advanced.timeout < 1 || formData.advanced.timeout > 300) {
			errors.timeout = '超时时间必须在1-300秒之间';
		}

		if (formData.advanced.rateLimit < 1 || formData.advanced.rateLimit > 1000) {
			errors.rateLimit = '速率限制必须在1-1000之间';
		}

		validationErrors = errors;
		return Object.keys(errors).length === 0;
	}

	// 提交表单
	async function handleSubmit() {
		if (!validateForm()) {
			toastStore.error('请检查表单输入');
			return;
		}

		isSubmitting = true;

		try {
			// 构建任务数据
			const taskData = {
				name: formData.name.trim(),
				description: formData.description.trim(),
				target: formData.target.trim(),
				ports: formData.ports.trim(),
				scanMethod: formData.scanMethod,
				maxWorkers: formData.advanced.maxWorkers,
				timeout: formData.advanced.timeout,
				enableBanner: formData.advanced.enableBanner,
				enableSSL: formData.advanced.enableSSL,
				enableService: formData.advanced.enableService,
				rateLimit: formData.advanced.rateLimit,
				projectId: formData.projectId || undefined
			};

			// 创建任务
			const response = await portScanStore.actions.createTask(taskData);

			toastStore.success('任务创建成功');
			await goto(`/portscan/${response.taskId}`);
		} catch (error) {
			console.error('创建任务失败:', error);
			toastStore.error('创建任务失败: ' + (error as Error).message);
		} finally {
			isSubmitting = false;
		}
	}

	// 重置表单
	function resetForm() {
		formData = {
			name: '',
			description: '',
			target: '',
			ports: '',
			preset: 'custom',
			scanMethod: 'tcp',
			advanced: {
				maxWorkers: 50,
				timeout: 30,
				enableBanner: false,
				enableSSL: false,
				enableService: true,
				rateLimit: 100
			},
			projectId: undefined
		};
		validationErrors = {};
	}

	// 获取端口数量估计
	let portCount = $derived(
		(() => {
			if (!formData.ports.trim()) return 0;
			const validation = portScanApi.validatePorts(formData.ports);
			return validation.valid ? validation.count || 0 : 0;
		})()
	);

	// 获取预估扫描时间
	let estimatedTime = $derived(
		(() => {
			if (portCount === 0) return '0分钟';
			const timePerPort = formData.scanMethod === 'both' ? 2 : 1; // 双协议扫描时间翻倍
			const totalTime = (portCount * timePerPort) / formData.advanced.maxWorkers;

			if (totalTime < 60) {
				return `${Math.ceil(totalTime)}秒`;
			} else if (totalTime < 3600) {
				return `${Math.ceil(totalTime / 60)}分钟`;
			} else {
				return `${Math.ceil(totalTime / 3600)}小时`;
			}
		})()
	);
</script>

<svelte:head>
	<title>创建端口扫描任务 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6">
	<!-- 页面标题 -->
	<div class="flex justify-between items-center mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">创建端口扫描任务</h1>
			<p class="text-gray-600 mt-1">配置并启动新的端口扫描任务</p>
		</div>
		<Button variant="outline" onclick={() => goto('/portscan')}>返回列表</Button>
	</div>

	<!-- 表单 -->
	<div class="max-w-4xl mx-auto">
		<form
			onsubmit={(e) => {
				e.preventDefault();
				handleSubmit();
			}}
			class="space-y-6"
		>
			<!-- 基本信息 -->
			<div class="bg-white rounded-lg shadow-sm border p-6">
				<h2 class="text-lg font-semibold text-gray-900 mb-4">基本信息</h2>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<FormField label="任务名称" required error={validationErrors.name}>
						<Input bind:value={formData.name} placeholder="输入任务名称" disabled={isSubmitting} />
					</FormField>

					<FormField label="关联项目" error={validationErrors.projectId}>
						<Select
							bind:value={formData.projectId}
							options={projectOptions}
							disabled={isSubmitting}
						/>
					</FormField>
				</div>

				<FormField label="任务描述" error={validationErrors.description}>
					<Textarea
						bind:value={formData.description}
						placeholder="输入任务描述（可选）"
						rows={3}
						disabled={isSubmitting}
					/>
				</FormField>
			</div>

			<!-- 扫描配置 -->
			<div class="bg-white rounded-lg shadow-sm border p-6">
				<h2 class="text-lg font-semibold text-gray-900 mb-4">扫描配置</h2>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<FormField label="扫描目标" required error={validationErrors.target}>
						<Input
							bind:value={formData.target}
							placeholder="example.com 或 192.168.1.1 或 192.168.1.0/24"
							disabled={isSubmitting}
						/>
						<div class="text-sm text-gray-500 mt-1">支持域名、IP地址或CIDR网段</div>
					</FormField>

					<FormField label="扫描方法" error={validationErrors.scanMethod}>
						<Select
							bind:value={formData.scanMethod}
							options={scanMethodOptions}
							disabled={isSubmitting}
						/>
					</FormField>
				</div>

				<FormField label="端口预设" error={validationErrors.preset}>
					<Select
						bind:value={formData.preset}
						options={presetOptions}
						onselect={handlePresetChange}
						disabled={isSubmitting}
					/>
					{#if formData.preset !== 'custom' && portPresets[formData.preset]}
						<div class="text-sm text-gray-500 mt-1">
							{portPresets[formData.preset].description}
						</div>
					{/if}
				</FormField>

				<FormField label="端口配置" required error={validationErrors.ports}>
					<Textarea
						bind:value={formData.ports}
						placeholder="80,443,1-1000,8080-8090"
						rows={3}
						disabled={isSubmitting}
					/>
					<div class="text-sm text-gray-500 mt-1">
						{#if portCount > 0}
							预计扫描 {portCount} 个端口，估计耗时 {estimatedTime}
						{:else}
							支持单个端口、范围或逗号分隔的组合
						{/if}
					</div>
				</FormField>
			</div>

			<!-- 高级设置 -->
			<div class="bg-white rounded-lg shadow-sm border p-6">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-gray-900">高级设置</h2>
					<Button type="button" variant="outline" onclick={() => (showAdvanced = !showAdvanced)}>
						{showAdvanced ? '隐藏' : '显示'}高级选项
					</Button>
				</div>

				{#if showAdvanced}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<FormField label="并发数" error={validationErrors.maxWorkers}>
							<Input
								type="number"
								bind:value={formData.advanced.maxWorkers}
								min={1}
								max={1000}
								disabled={isSubmitting}
							/>
							<div class="text-sm text-gray-500 mt-1">同时扫描的端口数量（1-1000）</div>
						</FormField>

						<FormField label="超时时间（秒）" error={validationErrors.timeout}>
							<Input
								type="number"
								bind:value={formData.advanced.timeout}
								min={1}
								max={300}
								disabled={isSubmitting}
							/>
							<div class="text-sm text-gray-500 mt-1">单个端口连接超时时间（1-300秒）</div>
						</FormField>

						<FormField label="速率限制" error={validationErrors.rateLimit}>
							<Input
								type="number"
								bind:value={formData.advanced.rateLimit}
								min={1}
								max={1000}
								disabled={isSubmitting}
							/>
							<div class="text-sm text-gray-500 mt-1">每秒最大请求数（1-1000）</div>
						</FormField>

						<div class="space-y-3">
							<div class="flex items-center gap-2">
								<input
									id="enable-service"
									type="checkbox"
									bind:checked={formData.advanced.enableService}
									disabled={isSubmitting}
									class="rounded border-gray-300"
								/>
								<label for="enable-service" class="text-sm text-gray-700">启用服务识别</label>
							</div>

							<div class="flex items-center gap-2">
								<input
									id="enable-banner"
									type="checkbox"
									bind:checked={formData.advanced.enableBanner}
									disabled={isSubmitting}
									class="rounded border-gray-300"
								/>
								<label for="enable-banner" class="text-sm text-gray-700">启用Banner抓取</label>
							</div>

							<div class="flex items-center gap-2">
								<input
									id="enable-ssl"
									type="checkbox"
									bind:checked={formData.advanced.enableSSL}
									disabled={isSubmitting}
									class="rounded border-gray-300"
								/>
								<label for="enable-ssl" class="text-sm text-gray-700">启用SSL检测</label>
							</div>
						</div>
					</div>
				{/if}
			</div>

			<!-- 操作按钮 -->
			<div class="flex justify-end gap-3">
				<Button type="button" variant="outline" onclick={resetForm} disabled={isSubmitting}>
					重置
				</Button>
				<Button type="submit" disabled={isSubmitting}>
					{#if isSubmitting}
						<LoadingSpinner size="sm" />
						创建中...
					{:else}
						创建任务
					{/if}
				</Button>
			</div>
		</form>
	</div>
</div>

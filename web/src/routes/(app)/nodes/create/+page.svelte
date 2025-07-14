<!--
节点创建页面
手动添加新节点的表单页面
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { nodeAPI } from '$lib/api/nodes';
	import { NodeRole } from '$lib/types/node';
	import type { NodeRegistrationRequest, NodeConfig, NodeRoleType } from '$lib/types/node';

	// 组件导入
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import FormField from '$lib/components/ui/FormField.svelte';
	import TagInput from '$lib/components/ui/TagInput.svelte';

	// 表单数据
	let formData = $state<NodeRegistrationRequest>({
		name: '',
		ip: '',
		port: 8090,
		role: NodeRole.WORKER,
		tags: [],
		config: {
			maxConcurrentTasks: 10,
			maxMemoryUsage: 2048,
			maxCpuUsage: 80,
			heartbeatInterval: 30,
			taskTimeout: 300,
			enabledTaskTypes: ['subdomain_enum', 'port_scan', 'vuln_scan'],
			logLevel: 'info',
			autoUpdate: true
		}
	});

	// 状态变量
	let loading = $state(false);
	let error = $state<string | null>(null);
	let success = $state<string | null>(null);

	// 角色选项
	const roleOptions = [
		{ value: NodeRole.MASTER, label: '主节点', description: '管理和协调其他节点' },
		{ value: NodeRole.WORKER, label: '工作节点', description: '执行具体的扫描任务' },
		{ value: NodeRole.SLAVE, label: '从节点', description: '备份和辅助节点' }
	];

	// 日志级别选项
	const logLevelOptions = [
		{ value: 'debug', label: 'Debug' },
		{ value: 'info', label: 'Info' },
		{ value: 'warn', label: 'Warn' },
		{ value: 'error', label: 'Error' }
	];

	// 可用任务类型
	const availableTaskTypes = [
		{ value: 'subdomain_enum', label: '子域名枚举' },
		{ value: 'port_scan', label: '端口扫描' },
		{ value: 'vuln_scan', label: '漏洞扫描' },
		{ value: 'web_crawl', label: '网页爬虫' },
		{ value: 'asset_discovery', label: '资产发现' },
		{ value: 'sensitive_info', label: '敏感信息检测' }
	];

	// 表单验证
	function validateForm(): string | null {
		if (!formData.name.trim()) {
			return '请输入节点名称';
		}

		if (!formData.ip.trim()) {
			return '请输入IP地址';
		}

		// 简单的IP地址格式验证
		const ipRegex =
			/^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
		if (!ipRegex.test(formData.ip)) {
			return '请输入有效的IP地址';
		}

		if (formData.port < 1 || formData.port > 65535) {
			return '端口号必须在1-65535之间';
		}

		if (formData.config.maxConcurrentTasks < 1 || formData.config.maxConcurrentTasks > 100) {
			return '最大并发任务数必须在1-100之间';
		}

		if (formData.config.maxMemoryUsage < 512 || formData.config.maxMemoryUsage > 32768) {
			return '最大内存使用量必须在512MB-32GB之间';
		}

		if (formData.config.maxCpuUsage < 10 || formData.config.maxCpuUsage > 100) {
			return 'CPU使用率限制必须在10%-100%之间';
		}

		if (formData.config.heartbeatInterval < 5 || formData.config.heartbeatInterval > 300) {
			return '心跳间隔必须在5-300秒之间';
		}

		if (formData.config.taskTimeout < 60 || formData.config.taskTimeout > 3600) {
			return '任务超时时间必须在60-3600秒之间';
		}

		return null;
	}

	// 提交表单
	async function submitForm() {
		const validationError = validateForm();
		if (validationError) {
			error = validationError;
			return;
		}

		try {
			loading = true;
			error = null;
			success = null;

			const response = await nodeAPI.createNode(formData);
			success = `节点创建成功！节点ID: ${response.nodeId}`;

			// 3秒后跳转到节点列表
			setTimeout(() => {
				goto('/nodes');
			}, 3000);
		} catch (err) {
			error = err instanceof Error ? err.message : '创建节点失败';
		} finally {
			loading = false;
		}
	}

	// 重置表单
	function resetForm() {
		formData = {
			name: '',
			ip: '',
			port: 8090,
			role: NodeRole.WORKER,
			tags: [],
			config: {
				maxConcurrentTasks: 10,
				maxMemoryUsage: 2048,
				maxCpuUsage: 80,
				heartbeatInterval: 30,
				taskTimeout: 300,
				enabledTaskTypes: ['subdomain_enum', 'port_scan', 'vuln_scan'],
				logLevel: 'info',
				autoUpdate: true
			}
		};
		error = null;
		success = null;
	}

	// 处理任务类型变化
	function handleTaskTypeChange(taskType: string, enabled: boolean) {
		if (enabled) {
			if (!formData.config.enabledTaskTypes.includes(taskType)) {
				formData.config.enabledTaskTypes = [...formData.config.enabledTaskTypes, taskType];
			}
		} else {
			formData.config.enabledTaskTypes = formData.config.enabledTaskTypes.filter(
				(t) => t !== taskType
			);
		}
	}
</script>

<svelte:head>
	<title>添加节点 - Stellar</title>
</svelte:head>

<div class="p-6 max-w-4xl mx-auto">
	<!-- 页面标题 -->
	<div class="flex items-center justify-between mb-6">
		<div class="flex items-center space-x-4">
			<Button variant="outline" onclick={() => goto('/nodes')}>← 返回</Button>
			<div>
				<h1 class="text-2xl font-bold text-gray-900">添加节点</h1>
				<p class="text-gray-600">手动添加新的计算节点</p>
			</div>
		</div>
	</div>

	<!-- 成功提示 -->
	{#if success}
		<div class="bg-green-50 border border-green-200 rounded-lg p-4 mb-6">
			<div class="flex items-center space-x-2">
				<span class="text-green-800">✅</span>
				<span class="text-green-800">{success}</span>
			</div>
		</div>
	{/if}

	<!-- 错误提示 -->
	{#if error}
		<div class="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
			<div class="flex items-center space-x-2">
				<span class="text-red-800">❌</span>
				<span class="text-red-800">{error}</span>
				<Button variant="outline" size="sm" onclick={() => (error = null)}>关闭</Button>
			</div>
		</div>
	{/if}

	<!-- 表单 -->
	<form
		onsubmit={(e) => {
			e.preventDefault();
			submitForm();
		}}
		class="space-y-8"
	>
		<!-- 基本信息 -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
			<h3 class="text-lg font-semibold text-gray-900 mb-6">基本信息</h3>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<FormField label="节点名称" required>
					<Input bind:value={formData.name} placeholder="输入节点名称" required />
				</FormField>

				<FormField label="节点角色" required>
					<Select bind:value={formData.role} options={roleOptions} />
				</FormField>

				<FormField label="IP地址" required>
					<Input bind:value={formData.ip} placeholder="192.168.1.100" required />
				</FormField>

				<FormField label="端口号" required>
					<Input type="number" bind:value={formData.port} min={1} max={65535} required />
				</FormField>
			</div>

			<div class="mt-6">
				<FormField label="标签">
					<TagInput bind:tags={formData.tags} placeholder="添加标签..." />
				</FormField>
			</div>
		</div>

		<!-- 性能配置 -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
			<h3 class="text-lg font-semibold text-gray-900 mb-6">性能配置</h3>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<FormField label="最大并发任务数" required>
					<Input
						type="number"
						bind:value={formData.config.maxConcurrentTasks}
						min={1}
						max={100}
						required
					/>
				</FormField>

				<FormField label="最大内存使用量 (MB)" required>
					<Input
						type="number"
						bind:value={formData.config.maxMemoryUsage}
						min={512}
						max={32768}
						step={256}
						required
					/>
				</FormField>

				<FormField label="最大CPU使用率 (%)" required>
					<Input
						type="number"
						bind:value={formData.config.maxCpuUsage}
						min={10}
						max={100}
						required
					/>
				</FormField>

				<FormField label="心跳间隔 (秒)" required>
					<Input
						type="number"
						bind:value={formData.config.heartbeatInterval}
						min={5}
						max={300}
						required
					/>
				</FormField>

				<FormField label="任务超时时间 (秒)" required>
					<Input
						type="number"
						bind:value={formData.config.taskTimeout}
						min={60}
						max={3600}
						required
					/>
				</FormField>

				<FormField label="日志级别" required>
					<Select bind:value={formData.config.logLevel} options={logLevelOptions} />
				</FormField>
			</div>
		</div>

		<!-- 任务类型配置 -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
			<h3 class="text-lg font-semibold text-gray-900 mb-6">启用的任务类型</h3>

			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each availableTaskTypes as taskType}
					<label class="flex items-center space-x-3 p-3 border rounded-lg hover:bg-gray-50">
						<input
							type="checkbox"
							checked={formData.config.enabledTaskTypes.includes(taskType.value)}
							onchange={(e) =>
								handleTaskTypeChange(taskType.value, (e.target as HTMLInputElement).checked)}
							class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
						/>
						<span class="text-sm font-medium text-gray-900">{taskType.label}</span>
					</label>
				{/each}
			</div>
		</div>

		<!-- 其他选项 -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
			<h3 class="text-lg font-semibold text-gray-900 mb-6">其他选项</h3>

			<div class="space-y-4">
				<label class="flex items-center space-x-3">
					<input
						type="checkbox"
						bind:checked={formData.config.autoUpdate}
						class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
					/>
					<span class="text-sm font-medium text-gray-900">自动更新</span>
					<span class="text-sm text-gray-500">启用后节点将自动更新到最新版本</span>
				</label>
			</div>
		</div>

		<!-- 操作按钮 -->
		<div class="flex items-center justify-end space-x-4">
			<Button type="button" variant="outline" onclick={resetForm} disabled={loading}>重置</Button>
			<Button type="button" variant="outline" onclick={() => goto('/nodes')} disabled={loading}>
				取消
			</Button>
			<Button type="submit" variant="default" disabled={loading}>
				{loading ? '创建中...' : '创建节点'}
			</Button>
		</div>
	</form>
</div>

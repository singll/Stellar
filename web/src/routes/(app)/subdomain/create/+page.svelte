<!-- 创建子域名枚举任务页面 -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { toastStore } from '$lib/stores/toast';
	import { subdomainStore } from '$lib/stores/subdomain';
	import { subdomainApi } from '$lib/api/subdomain';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import FormField from '$lib/components/ui/FormField.svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import TagInput from '$lib/components/ui/TagInput.svelte';
	import type { SubdomainFormData } from '$lib/types/subdomain';

	// 表单数据
	let formData = $state<SubdomainFormData>({
		name: '',
		description: '',
		target: '',
		methodPreset: 'standard',
		enumMethods: ['dns_brute', 'cert_transparency'],
		wordlistType: 'preset',
		wordlistPreset: 'common',
		customWordlist: [],
		dnsPreset: 'public',
		customDnsServers: [],
		advanced: {
			maxWorkers: 50,
			timeout: 30,
			maxRetries: 3,
			rateLimit: 10,
			enableWildcard: true,
			enableDoH: false,
			enableRecursive: false,
			maxDepth: 2,
			verifySubdomains: true,
			enableCache: true,
			cacheTimeout: 300
		},
		searchEngines: {
			enableGoogle: true,
			enableBing: true,
			censysApiKey: '',
			shodanApiKey: ''
		},
		projectId: undefined
	});

	// 表单状态
	let isSubmitting = $state(false);
	let showAdvanced = $state(false);
	let showSearchEngines = $state(false);
	let validationErrors = $state<Record<string, string>>({});

	// 获取预设配置
	const methodPresets = subdomainApi.getPresetMethods();
	const dnsPresets = subdomainApi.getPresetDNSServers();
	const wordlistPresets = subdomainApi.getPresetWordlists();

	// 预设选项
	const methodPresetOptions = [
		...Object.entries(methodPresets).map(([key, preset]) => ({
			value: key,
			label: preset.name
		}))
	];

	const dnsPresetOptions = [
		...Object.entries(dnsPresets).map(([key, preset]) => ({
			value: key,
			label: preset.name
		}))
	];

	const wordlistPresetOptions = [
		...Object.entries(wordlistPresets).map(([key, preset]) => ({
			value: key,
			label: preset.name
		}))
	];

	// 枚举方法选项
	const enumMethodOptions = [
		{ value: 'dns_brute', label: 'DNS暴力破解' },
		{ value: 'cert_transparency', label: '证书透明度' },
		{ value: 'search_engine', label: '搜索引擎' },
		{ value: 'dns_transfer', label: 'DNS区域传输' }
	];

	// 项目选项（模拟数据）
	const projectOptions = [
		{ value: '', label: '不关联项目' },
		{ value: 'proj1', label: '项目 1' },
		{ value: 'proj2', label: '项目 2' }
	];

	// 处理方法预设选择
	function handleMethodPresetChange(preset: string) {
		if (methodPresets[preset]) {
			formData.enumMethods = [...methodPresets[preset].methods];
		}
	}

	// 处理DNS预设选择
	function handleDnsPresetChange(preset: string) {
		if (dnsPresets[preset] && preset !== 'custom') {
			formData.customDnsServers = [...dnsPresets[preset].servers];
		} else if (preset === 'custom') {
			formData.customDnsServers = [];
		}
	}

	// 验证表单
	function validateForm(): boolean {
		const errors: Record<string, string> = {};

		// 验证任务名称
		if (!formData.name.trim()) {
			errors.name = '任务名称不能为空';
		}

		// 验证目标域名
		if (!formData.target.trim()) {
			errors.target = '目标域名不能为空';
		} else {
			const validation = subdomainApi.validateDomain(formData.target.trim());
			if (!validation.valid) {
				errors.target = validation.message || '域名格式错误';
			}
		}

		// 验证枚举方法
		if (formData.enumMethods.length === 0) {
			errors.enumMethods = '至少选择一种枚举方法';
		}

		// 验证自定义字典
		if (formData.wordlistType === 'custom') {
			if (formData.customWordlist.length === 0) {
				errors.customWordlist = '自定义字典不能为空';
			} else {
				const validation = subdomainApi.validateWordlist(formData.customWordlist);
				if (!validation.valid) {
					errors.customWordlist = validation.message || '自定义字典格式错误';
				}
			}
		}

		// 验证自定义DNS服务器
		if (formData.dnsPreset === 'custom' && formData.customDnsServers.length === 0) {
			errors.customDnsServers = '自定义DNS服务器不能为空';
		}

		// 验证高级设置
		if (formData.advanced.maxWorkers < 1 || formData.advanced.maxWorkers > 200) {
			errors.maxWorkers = '并发数必须在1-200之间';
		}

		if (formData.advanced.timeout < 1 || formData.advanced.timeout > 300) {
			errors.timeout = '超时时间必须在1-300秒之间';
		}

		if (formData.advanced.rateLimit < 1 || formData.advanced.rateLimit > 100) {
			errors.rateLimit = '速率限制必须在1-100之间';
		}

		if (formData.advanced.maxDepth < 1 || formData.advanced.maxDepth > 5) {
			errors.maxDepth = '递归深度必须在1-5之间';
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
				maxWorkers: formData.advanced.maxWorkers,
				timeout: formData.advanced.timeout,
				wordlistPath: formData.wordlistType === 'preset' ? formData.wordlistPreset : '',
				customWordlist: formData.wordlistType === 'custom' ? formData.customWordlist : undefined,
				dnsServers:
					formData.dnsPreset === 'custom'
						? formData.customDnsServers
						: dnsPresets[formData.dnsPreset]?.servers,
				enableWildcard: formData.advanced.enableWildcard,
				maxRetries: formData.advanced.maxRetries,
				enumMethods: formData.enumMethods,
				rateLimit: formData.advanced.rateLimit,
				enableDoH: formData.advanced.enableDoH,
				enableRecursive: formData.advanced.enableRecursive,
				maxDepth: formData.advanced.maxDepth,
				verifySubdomains: formData.advanced.verifySubdomains,
				enableCache: formData.advanced.enableCache,
				cacheTimeout: formData.advanced.cacheTimeout,
				searchEngineAPIs: {
					...(formData.searchEngines.censysApiKey && {
						censys: formData.searchEngines.censysApiKey
					}),
					...(formData.searchEngines.shodanApiKey && {
						shodan: formData.searchEngines.shodanApiKey
					})
				},
				projectId: formData.projectId || undefined
			};

			// 创建任务
			const task = await subdomainStore.actions.createTask(taskData);

			toastStore.success('任务创建成功');
			await goto(`/subdomain/${task.id}`);
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
			methodPreset: 'standard',
			enumMethods: ['dns_brute', 'cert_transparency'],
			wordlistType: 'preset',
			wordlistPreset: 'common',
			customWordlist: [],
			dnsPreset: 'public',
			customDnsServers: [],
			advanced: {
				maxWorkers: 50,
				timeout: 30,
				maxRetries: 3,
				rateLimit: 10,
				enableWildcard: true,
				enableDoH: false,
				enableRecursive: false,
				maxDepth: 2,
				verifySubdomains: true,
				enableCache: true,
				cacheTimeout: 300
			},
			searchEngines: {
				enableGoogle: true,
				enableBing: true,
				censysApiKey: '',
				shodanApiKey: ''
			},
			projectId: undefined
		};
		validationErrors = {};
	}

	// 预估扫描时间
	$: estimatedTime = (() => {
		const wordlistSize =
			formData.wordlistType === 'custom'
				? formData.customWordlist.length
				: wordlistPresets[formData.wordlistPreset]?.size
					? parseInt(wordlistPresets[formData.wordlistPreset].size.replace(/[^\d]/g, '')) || 1000
					: 1000;

		const time = subdomainApi.estimateScanTime({
			wordlistSize,
			enumMethods: formData.enumMethods,
			maxWorkers: formData.advanced.maxWorkers,
			rateLimit: formData.advanced.rateLimit,
			enableRecursive: formData.advanced.enableRecursive
		});

		if (time < 60) {
			return `约${time}秒`;
		} else if (time < 3600) {
			return `约${Math.ceil(time / 60)}分钟`;
		} else {
			return `约${Math.ceil(time / 3600)}小时`;
		}
	})();

	// 处理自定义字典输入
	function handleWordlistChange(value: string) {
		const lines = value
			.split('\n')
			.map((line) => line.trim())
			.filter((line) => line.length > 0);
		formData.customWordlist = lines;
	}

	// 处理DNS服务器输入
	function handleDnsServerAdd(server: string) {
		if (server && !formData.customDnsServers.includes(server)) {
			formData.customDnsServers = [...formData.customDnsServers, server];
		}
	}

	function handleDnsServerRemove(server: string) {
		formData.customDnsServers = formData.customDnsServers.filter((s) => s !== server);
	}
</script>

<svelte:head>
	<title>创建子域名枚举任务 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6">
	<!-- 页面标题 -->
	<div class="flex justify-between items-center mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">创建子域名枚举任务</h1>
			<p class="text-gray-600 mt-1">配置并启动新的子域名枚举任务</p>
		</div>
		<Button variant="outline" onclick={() => goto('/subdomain')}>返回列表</Button>
	</div>

	<!-- 表单 -->
	<div class="max-w-4xl mx-auto">
		<form onsubmit|preventDefault={handleSubmit} class="space-y-6">
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

				<FormField label="目标域名" required error={validationErrors.target}>
					<Input bind:value={formData.target} placeholder="example.com" disabled={isSubmitting} />
					<div class="text-sm text-gray-500 mt-1">输入要枚举子域名的根域名</div>
				</FormField>
			</div>

			<!-- 枚举配置 -->
			<div class="bg-white rounded-lg shadow-sm border p-6">
				<h2 class="text-lg font-semibold text-gray-900 mb-4">枚举配置</h2>

				<FormField label="方法预设" error={validationErrors.methodPreset}>
					<Select
						bind:value={formData.methodPreset}
						options={methodPresetOptions}
						onchange={handleMethodPresetChange}
						disabled={isSubmitting}
					/>
					{#if methodPresets[formData.methodPreset]}
						<div class="text-sm text-gray-500 mt-1">
							{methodPresets[formData.methodPreset].description}
						</div>
					{/if}
				</FormField>

				<FormField label="枚举方法" required error={validationErrors.enumMethods}>
					<div class="space-y-2">
						{#each enumMethodOptions as method}
							<label class="flex items-center">
								<input
									type="checkbox"
									bind:group={formData.enumMethods}
									value={method.value}
									disabled={isSubmitting}
									class="rounded border-gray-300 mr-2"
								/>
								<span class="text-sm text-gray-700">{method.label}</span>
							</label>
						{/each}
					</div>
				</FormField>

				<!-- 字典配置 -->
				<div class="space-y-4">
					<div class="flex items-center gap-4">
						<label class="flex items-center">
							<input
								type="radio"
								bind:group={formData.wordlistType}
								value="preset"
								disabled={isSubmitting}
								class="mr-2"
							/>
							<span class="text-sm font-medium text-gray-700">使用预设字典</span>
						</label>
						<label class="flex items-center">
							<input
								type="radio"
								bind:group={formData.wordlistType}
								value="custom"
								disabled={isSubmitting}
								class="mr-2"
							/>
							<span class="text-sm font-medium text-gray-700">自定义字典</span>
						</label>
					</div>

					{#if formData.wordlistType === 'preset'}
						<FormField label="字典预设" error={validationErrors.wordlistPreset}>
							<Select
								bind:value={formData.wordlistPreset}
								options={wordlistPresetOptions}
								disabled={isSubmitting}
							/>
							{#if wordlistPresets[formData.wordlistPreset]}
								<div class="text-sm text-gray-500 mt-1">
									{wordlistPresets[formData.wordlistPreset].description} ({wordlistPresets[
										formData.wordlistPreset
									].size})
								</div>
							{/if}
						</FormField>
					{:else}
						<FormField label="自定义字典" required error={validationErrors.customWordlist}>
							<Textarea
								value={formData.customWordlist.join('\n')}
								onchange={(e) => handleWordlistChange(e.target.value)}
								placeholder="每行一个子域名前缀，例如：&#10;www&#10;mail&#10;ftp&#10;api"
								rows={8}
								disabled={isSubmitting}
							/>
							<div class="text-sm text-gray-500 mt-1">
								共 {formData.customWordlist.length} 个字典项
							</div>
						</FormField>
					{/if}
				</div>

				<!-- DNS服务器配置 -->
				<FormField label="DNS服务器" error={validationErrors.dnsPreset}>
					<Select
						bind:value={formData.dnsPreset}
						options={dnsPresetOptions}
						onchange={handleDnsPresetChange}
						disabled={isSubmitting}
					/>
					{#if dnsPresets[formData.dnsPreset]}
						<div class="text-sm text-gray-500 mt-1">
							{dnsPresets[formData.dnsPreset].description}
						</div>
					{/if}
				</FormField>

				{#if formData.dnsPreset === 'custom'}
					<FormField label="自定义DNS服务器" required error={validationErrors.customDnsServers}>
						<TagInput
							bind:tags={formData.customDnsServers}
							placeholder="输入DNS服务器IP地址"
							disabled={isSubmitting}
						/>
					</FormField>
				{/if}
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
								min="1"
								max="200"
								disabled={isSubmitting}
							/>
							<div class="text-sm text-gray-500 mt-1">同时进行的DNS查询数量（1-200）</div>
						</FormField>

						<FormField label="超时时间（秒）" error={validationErrors.timeout}>
							<Input
								type="number"
								bind:value={formData.advanced.timeout}
								min="1"
								max="300"
								disabled={isSubmitting}
							/>
						</FormField>

						<FormField label="重试次数" error={validationErrors.maxRetries}>
							<Input
								type="number"
								bind:value={formData.advanced.maxRetries}
								min="0"
								max="10"
								disabled={isSubmitting}
							/>
						</FormField>

						<FormField label="速率限制" error={validationErrors.rateLimit}>
							<Input
								type="number"
								bind:value={formData.advanced.rateLimit}
								min="1"
								max="100"
								disabled={isSubmitting}
							/>
							<div class="text-sm text-gray-500 mt-1">每秒最大DNS查询数（1-100）</div>
						</FormField>

						<FormField label="递归深度" error={validationErrors.maxDepth}>
							<Input
								type="number"
								bind:value={formData.advanced.maxDepth}
								min="1"
								max="5"
								disabled={isSubmitting}
							/>
						</FormField>

						<FormField label="缓存超时（秒）">
							<Input
								type="number"
								bind:value={formData.advanced.cacheTimeout}
								min="60"
								max="3600"
								disabled={isSubmitting}
							/>
						</FormField>
					</div>

					<div class="mt-4 grid grid-cols-1 md:grid-cols-2 gap-6">
						<div class="space-y-3">
							<h4 class="font-medium text-gray-900">DNS选项</h4>
							<div class="flex items-center gap-2">
								<input
									type="checkbox"
									bind:checked={formData.advanced.enableWildcard}
									disabled={isSubmitting}
									class="rounded border-gray-300"
								/>
								<label class="text-sm text-gray-700">启用通配符检测</label>
							</div>

							<div class="flex items-center gap-2">
								<input
									type="checkbox"
									bind:checked={formData.advanced.enableDoH}
									disabled={isSubmitting}
									class="rounded border-gray-300"
								/>
								<label class="text-sm text-gray-700">启用DNS over HTTPS</label>
							</div>

							<div class="flex items-center gap-2">
								<input
									type="checkbox"
									bind:checked={formData.advanced.enableCache}
									disabled={isSubmitting}
									class="rounded border-gray-300"
								/>
								<label class="text-sm text-gray-700">启用DNS缓存</label>
							</div>
						</div>

						<div class="space-y-3">
							<h4 class="font-medium text-gray-900">扫描选项</h4>
							<div class="flex items-center gap-2">
								<input
									type="checkbox"
									bind:checked={formData.advanced.enableRecursive}
									disabled={isSubmitting}
									class="rounded border-gray-300"
								/>
								<label class="text-sm text-gray-700">启用递归枚举</label>
							</div>

							<div class="flex items-center gap-2">
								<input
									type="checkbox"
									bind:checked={formData.advanced.verifySubdomains}
									disabled={isSubmitting}
									class="rounded border-gray-300"
								/>
								<label class="text-sm text-gray-700">验证子域名活跃性</label>
							</div>
						</div>
					</div>
				{/if}
			</div>

			<!-- 搜索引擎API -->
			<div class="bg-white rounded-lg shadow-sm border p-6">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-gray-900">搜索引擎API</h2>
					<Button
						type="button"
						variant="outline"
						onclick={() => (showSearchEngines = !showSearchEngines)}
					>
						{showSearchEngines ? '隐藏' : '显示'}API配置
					</Button>
				</div>

				{#if showSearchEngines}
					<div class="space-y-4">
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<FormField label="Censys API密钥">
								<Input
									bind:value={formData.searchEngines.censysApiKey}
									placeholder="可选，用于证书透明度查询"
									type="password"
									disabled={isSubmitting}
								/>
							</FormField>

							<FormField label="Shodan API密钥">
								<Input
									bind:value={formData.searchEngines.shodanApiKey}
									placeholder="可选，用于网络资产查询"
									type="password"
									disabled={isSubmitting}
								/>
							</FormField>
						</div>

						<div class="text-sm text-gray-500">
							<p>API密钥是可选的，但可以提高查询精度和数量限制。</p>
							<p>不提供API密钥时将使用免费的公开接口。</p>
						</div>
					</div>
				{/if}
			</div>

			<!-- 预估信息 -->
			<div class="bg-blue-50 rounded-lg p-4">
				<h3 class="font-medium text-blue-900 mb-2">扫描预估</h3>
				<div class="text-sm text-blue-800">
					<p>预计扫描时间: {estimatedTime}</p>
					<p>
						使用方法: {formData.enumMethods
							.map((m) => enumMethodOptions.find((o) => o.value === m)?.label)
							.join(', ')}
					</p>
				</div>
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

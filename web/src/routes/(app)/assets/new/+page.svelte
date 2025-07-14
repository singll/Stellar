<script lang="ts">
	import { goto } from '$app/navigation';
	import { assetApi } from '$lib/api/asset';
	import type { CreateAssetRequest } from '$lib/types/asset';
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
	import {
		Select,
		SelectContent,
		SelectItem,
		SelectTrigger,
		SelectValue
	} from '$lib/components/ui/select';
	import Icon from '$lib/components/ui/Icon.svelte';

	// 表单状态
	let formData: CreateAssetRequest = $state({
		type: 'domain',
		domain: '',
		ip: '',
		url: '',
		port: 80,
		appName: '',
		projectId: '',
		tags: [],
		data: {} // 添加必需的 data 字段
	});

	let loading = $state(false);
	let errors = $state<Record<string, string>>({});

	// 资产类型选项
	const assetTypes = [
		{ value: 'domain', label: '域名' },
		{ value: 'ip', label: 'IP地址' },
		{ value: 'url', label: 'URL' },
		{ value: 'port', label: '端口' },
		{ value: 'app', label: '应用' }
	];

	// 表单验证
	const validateForm = (): boolean => {
		const newErrors: Record<string, string> = {};

		if (!formData.type) {
			newErrors.type = '请选择资产类型';
		}

		// 根据类型验证必填字段
		switch (formData.type) {
			case 'domain':
				if (!formData.domain?.trim()) {
					newErrors.domain = '域名不能为空';
				} else if (
					!/^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/.test(
						formData.domain.trim()
					)
				) {
					newErrors.domain = '请输入有效的域名';
				}
				break;
			case 'ip':
				if (!formData.ip?.trim()) {
					newErrors.ip = 'IP地址不能为空';
				} else if (!/^(\d{1,3}\.){3}\d{1,3}$/.test(formData.ip.trim())) {
					newErrors.ip = '请输入有效的IP地址';
				}
				break;
			case 'url':
				if (!formData.url?.trim()) {
					newErrors.url = 'URL不能为空';
				} else {
					try {
						new URL(formData.url.trim());
					} catch {
						newErrors.url = '请输入有效的URL';
					}
				}
				break;
			case 'port':
				if (!formData.ip?.trim()) {
					newErrors.ip = 'IP地址不能为空';
				}
				if (!formData.port || formData.port < 1 || formData.port > 65535) {
					newErrors.port = '端口范围应在1-65535之间';
				}
				break;
			case 'app':
				if (!formData.appName?.trim()) {
					newErrors.appName = '应用名称不能为空';
				}
				break;
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
			// 清理表单数据
			const cleanData: CreateAssetRequest = {
				type: formData.type,
				projectId: formData.projectId || '', // 改为空字符串而不是 undefined
				tags: formData.tags || [],
				data: {} // 添加必需的 data 字段
			};

			// 根据类型添加相应字段
			switch (formData.type) {
				case 'domain':
					cleanData.domain = formData.domain?.trim();
					break;
				case 'ip':
					cleanData.ip = formData.ip?.trim();
					break;
				case 'url':
					cleanData.url = formData.url?.trim();
					break;
				case 'port':
					cleanData.ip = formData.ip?.trim();
					cleanData.port = formData.port;
					break;
				case 'app':
					cleanData.appName = formData.appName?.trim();
					cleanData.ip = formData.ip?.trim();
					cleanData.url = formData.url?.trim();
					break;
			}

			const response = await assetApi.createAsset(cleanData);

			notifications.add({
				type: 'success',
				message: '资产创建成功'
			});

			// 跳转到资产详情页
			await goto(`/assets/${response.data.id}`);
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '创建资产失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		} finally {
			loading = false;
		}
	};

	// 实时验证
	$effect(() => {
		if (formData.type) {
			validateForm();
		}
	});
</script>

<svelte:head>
	<title>创建资产 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6 max-w-2xl">
	<!-- 页面标题 -->
	<div class="mb-6">
		<div class="flex items-center gap-4 mb-4">
			<Button variant="ghost" onclick={() => goto('/assets')} class="flex items-center gap-2">
				<Icon name="chevron-left" class="h-4 w-4" />
				返回资产列表
			</Button>
		</div>

		<h1 class="text-3xl font-bold text-gray-900">创建新资产</h1>
		<p class="text-gray-600 mt-1">添加新的安全资产以进行监控和扫描</p>
	</div>

	<!-- 创建表单 -->
	<Card>
		<CardHeader>
			<CardTitle>资产信息</CardTitle>
			<CardDescription>填写资产的基本信息。标有 * 的字段为必填项。</CardDescription>
		</CardHeader>

		<CardContent>
			<form onsubmit={handleSubmit} class="space-y-6">
				<!-- 资产类型 -->
				<div class="space-y-2">
					<Label for="type" class="text-sm font-medium">
						资产类型 <span class="text-red-500">*</span>
					</Label>
					<Select bind:value={formData.type} disabled={loading}>
						<SelectTrigger>
							<SelectValue placeholder="选择资产类型" />
						</SelectTrigger>
						<SelectContent>
							{#each assetTypes as type}
								<SelectItem value={type.value}>{type.label}</SelectItem>
							{/each}
						</SelectContent>
					</Select>
					{#if errors.type}
						<p class="text-sm text-red-600">{errors.type}</p>
					{/if}
				</div>

				<!-- 根据类型显示相应字段 -->
				{#if formData.type === 'domain'}
					<div class="space-y-2">
						<Label for="domain" class="text-sm font-medium">
							域名 <span class="text-red-500">*</span>
						</Label>
						<Input
							id="domain"
							type="text"
							bind:value={formData.domain}
							placeholder="example.com"
							class={errors.domain ? 'border-red-500 focus:border-red-500' : ''}
							disabled={loading}
							required
						/>
						{#if errors.domain}
							<p class="text-sm text-red-600">{errors.domain}</p>
						{:else}
							<p class="text-xs text-gray-500">输入要监控的域名</p>
						{/if}
					</div>
				{:else if formData.type === 'ip'}
					<div class="space-y-2">
						<Label for="ip" class="text-sm font-medium">
							IP地址 <span class="text-red-500">*</span>
						</Label>
						<Input
							id="ip"
							type="text"
							bind:value={formData.ip}
							placeholder="192.168.1.1"
							class={errors.ip ? 'border-red-500 focus:border-red-500' : ''}
							disabled={loading}
							required
						/>
						{#if errors.ip}
							<p class="text-sm text-red-600">{errors.ip}</p>
						{:else}
							<p class="text-xs text-gray-500">输入要监控的IP地址</p>
						{/if}
					</div>
				{:else if formData.type === 'url'}
					<div class="space-y-2">
						<Label for="url" class="text-sm font-medium">
							URL <span class="text-red-500">*</span>
						</Label>
						<Input
							id="url"
							type="url"
							bind:value={formData.url}
							placeholder="https://example.com/path"
							class={errors.url ? 'border-red-500 focus:border-red-500' : ''}
							disabled={loading}
							required
						/>
						{#if errors.url}
							<p class="text-sm text-red-600">{errors.url}</p>
						{:else}
							<p class="text-xs text-gray-500">输入要监控的完整URL</p>
						{/if}
					</div>
				{:else if formData.type === 'port'}
					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-2">
							<Label for="ip" class="text-sm font-medium">
								IP地址 <span class="text-red-500">*</span>
							</Label>
							<Input
								id="ip"
								type="text"
								bind:value={formData.ip}
								placeholder="192.168.1.1"
								class={errors.ip ? 'border-red-500 focus:border-red-500' : ''}
								disabled={loading}
								required
							/>
							{#if errors.ip}
								<p class="text-sm text-red-600">{errors.ip}</p>
							{/if}
						</div>
						<div class="space-y-2">
							<Label for="port" class="text-sm font-medium">
								端口 <span class="text-red-500">*</span>
							</Label>
							<Input
								id="port"
								type="number"
								bind:value={formData.port}
								placeholder="80"
								min={1}
								max={65535}
								class={errors.port ? 'border-red-500 focus:border-red-500' : ''}
								disabled={loading}
								required
							/>
							{#if errors.port}
								<p class="text-sm text-red-600">{errors.port}</p>
							{/if}
						</div>
					</div>
				{:else if formData.type === 'app'}
					<div class="space-y-4">
						<div class="space-y-2">
							<Label for="appName" class="text-sm font-medium">
								应用名称 <span class="text-red-500">*</span>
							</Label>
							<Input
								id="appName"
								type="text"
								bind:value={formData.appName}
								placeholder="Web应用名称"
								class={errors.appName ? 'border-red-500 focus:border-red-500' : ''}
								disabled={loading}
								required
							/>
							{#if errors.appName}
								<p class="text-sm text-red-600">{errors.appName}</p>
							{/if}
						</div>
						<div class="space-y-2">
							<Label for="appIp" class="text-sm font-medium">IP地址（可选）</Label>
							<Input
								id="appIp"
								type="text"
								bind:value={formData.ip}
								placeholder="192.168.1.1"
								disabled={loading}
							/>
						</div>
						<div class="space-y-2">
							<Label for="appUrl" class="text-sm font-medium">URL（可选）</Label>
							<Input
								id="appUrl"
								type="url"
								bind:value={formData.url}
								placeholder="https://app.example.com"
								disabled={loading}
							/>
						</div>
					</div>
				{/if}

				<!-- 项目关联 -->
				<div class="space-y-2">
					<Label for="projectId" class="text-sm font-medium">关联项目（可选）</Label>
					<Input
						id="projectId"
						type="text"
						bind:value={formData.projectId}
						placeholder="项目ID"
						disabled={loading}
					/>
					<p class="text-xs text-gray-500">将此资产关联到指定项目</p>
				</div>

				<!-- 提交按钮 -->
				<div class="flex items-center gap-4 pt-4 border-t">
					<button
						type="submit"
						disabled={loading || !formData.type}
						class="inline-flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						{#if loading}
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
							创建中...
						{:else}
							<Icon name="check" class="h-4 w-4" />
							创建资产
						{/if}
					</button>

					<button
						type="button"
						onclick={() => goto('/assets')}
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
			<CardTitle class="text-lg">资产类型说明</CardTitle>
		</CardHeader>
		<CardContent class="space-y-3 text-sm text-gray-600">
			<div>
				<strong>域名:</strong> 监控域名的子域名发现、DNS记录等
			</div>
			<div>
				<strong>IP地址:</strong> 监控IP地址的端口扫描、服务识别等
			</div>
			<div>
				<strong>URL:</strong> 监控特定URL的内容变化、安全漏洞等
			</div>
			<div>
				<strong>端口:</strong> 监控特定IP和端口的服务状态
			</div>
			<div>
				<strong>应用:</strong> 监控Web应用的综合安全状态
			</div>
		</CardContent>
	</Card>
</div>

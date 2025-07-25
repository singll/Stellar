<script lang="ts">
	import { goto } from '$app/navigation';
	import { assetApi } from '$lib/api/asset';
	import type { CreateAssetRequest, AssetType } from '$lib/types/asset';
	import { notifications } from '$lib/stores/notifications';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Select } from '$lib/components/ui/select';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import Icon from '$lib/components/ui/Icon.svelte';
	import SearchableProjectSelector from '$lib/components/ui/searchable-project-selector/SearchableProjectSelector.svelte';
	import type { Project } from '$lib/types/project';

	// 表单状态
	let formData = $state({
		type: 'domain' as AssetType,
		projectId: '',
		tags: [] as string[],
		description: '',
		// 具体字段
		domain: '',
		ip: '',
		url: '',
		port: 80,
		appName: '',
		packageName: '',
		host: '',
		subdomain: '',
		service: '',
		protocol: 'tcp'
	});

	let loading = $state(false);
	let errors = $state<Record<string, string>>({});
	
	// 项目选择相关状态
	let selectedProject = $state<Project | null>(null);
	let selectedProjectName = $state('');
	
	// 标签输入
	let tagInput = $state('');

	// 预定义的资产类型
	const assetTypes = [
		{ value: 'domain', label: '域名', description: '监控域名的子域名发现、DNS记录等', icon: 'globe' },
		{ value: 'subdomain', label: '子域名', description: '监控子域名的解析、可用性等', icon: 'globe' },
		{ value: 'ip', label: 'IP地址', description: '监控IP地址的端口扫描、服务识别等', icon: 'server' },
		{ value: 'port', label: '端口', description: '监控特定IP和端口的服务状态', icon: 'wifi' },
		{ value: 'url', label: 'URL', description: '监控特定URL的内容变化、安全漏洞等', icon: 'link' },
		{ value: 'http', label: 'HTTP服务', description: '监控HTTP服务的可用性和安全性', icon: 'globe' },
		{ value: 'app', label: '移动应用', description: '监控移动应用的安全状态', icon: 'smartphone' },
		{ value: 'miniapp', label: '小程序', description: '监控小程序的安全状态', icon: 'smartphone' }
	];

	// 获取当前资产类型信息
	const getCurrentAssetType = () => assetTypes.find(t => t.value === formData.type);

	// 标签管理
	const addTag = () => {
		if (tagInput.trim() && !formData.tags.includes(tagInput.trim())) {
			formData.tags = [...formData.tags, tagInput.trim()];
			tagInput = '';
		}
	};
	
	const removeTag = (tag: string) => {
		formData.tags = formData.tags.filter(t => t !== tag);
	};
	
	const handleTagKeydown = (event: KeyboardEvent) => {
		if (event.key === 'Enter') {
			event.preventDefault();
			addTag();
		}
	};
	
	// 表单验证
	const validateForm = (): boolean => {
		const newErrors: Record<string, string> = {};

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
				} else if (!/^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/.test(formData.ip.trim())) {
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
			case 'subdomain':
				if (!formData.subdomain?.trim()) {
					newErrors.subdomain = '子域名不能为空';
				}
				break;
			case 'app':
				if (!formData.appName?.trim()) {
					newErrors.appName = '应用名称不能为空';
				}
				break;
			case 'http':
				if (!formData.host?.trim()) {
					newErrors.host = '主机地址不能为空';
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
			const requestData: CreateAssetRequest = {
				type: formData.type,
				projectId: formData.projectId || '',
				tags: formData.tags || [],
				data: {}
			};

			// 根据类型设置相应的数据
			switch (formData.type) {
				case 'domain':
					requestData.data = { domain: formData.domain.trim() };
					break;
				case 'ip':
					requestData.data = { ip: formData.ip.trim() };
					break;
				case 'url':
					requestData.data = { url: formData.url.trim() };
					break;
				case 'port':
					requestData.data = { ip: formData.ip.trim(), port: formData.port };
					break;
				case 'subdomain':
					requestData.data = { host: formData.subdomain.trim() };
					break;
				case 'app':
					requestData.data = { appName: formData.appName.trim() };
					break;
				case 'http':
					requestData.data = { host: formData.host.trim() };
					break;
				default:
					requestData.data = {};
			}

			const response = await assetApi.createAsset(requestData);

			notifications.add({
				type: 'success',
				message: '资产创建成功'
			});

			// 跳转到资产列表页
			await goto('/assets');
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '创建资产失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		} finally {
			loading = false;
		}
	};

	// 处理项目选择
	const handleProjectSelect = (project: Project | null) => {
		selectedProject = project;
		formData.projectId = project?.id || '';
		selectedProjectName = project?.name || '';
		console.log('🎯 项目选择变更:', {
			project: project?.name,
			id: project?.id,
			formProjectId: formData.projectId
		});
	};
</script>

<svelte:head>
	<title>创建资产 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6 max-w-4xl">
	<!-- 页面标题 -->
	<div class="mb-8">
		<div class="flex items-center gap-4 mb-6">
			<Button variant="ghost" onclick={() => goto('/assets')} class="flex items-center gap-2">
				<Icon name="chevron-left" class="h-4 w-4" />
				返回资产列表
			</Button>
		</div>

		<div class="text-center mb-8">
			<div class="inline-flex items-center justify-center w-16 h-16 bg-blue-100 rounded-full mb-4">
				<Icon name="plus" class="h-8 w-8 text-blue-600" />
			</div>
			<h1 class="text-3xl font-bold text-gray-900 mb-2">创建新资产</h1>
			<p class="text-gray-600">添加新的安全资产以进行监控和扫描</p>
		</div>
	</div>

	<!-- 单页面表单 -->
	<form onsubmit={handleSubmit}>
		<Card class="max-w-4xl mx-auto">
			<CardHeader>
				<CardTitle class="flex items-center gap-2">
					<Icon name="layers" class="h-5 w-5 text-blue-600" />
					创建资产
				</CardTitle>
				<CardDescription>选择资产类型并填写相关信息</CardDescription>
			</CardHeader>
			<CardContent class="space-y-8">
				<!-- 资产类型选择 -->
				<div class="space-y-4">
					<Label class="text-lg font-medium">资产类型 <span class="text-red-500">*</span></Label>
					<Select 
						bind:value={formData.type} 
						placeholder="选择资产类型"
						options={assetTypes.map(type => ({ value: type.value, label: `${type.label} - ${type.description}` }))}
						class="w-full"
					/>
					{#if getCurrentAssetType()}
						{@const currentType = getCurrentAssetType()}
						<p class="text-sm text-gray-600">📝 {currentType.description}</p>
					{/if}
				</div>

				<!-- 动态表单字段 -->
				<div class="space-y-6">
					<!-- 基于资产类型的动态表单 -->
					{#if formData.type === 'domain'}
						<div class="space-y-2">
							<Label for="domain" class="text-sm font-medium">
								域名地址 <span class="text-red-500">*</span>
							</Label>
							<Input
								id="domain"
								type="text"
								bind:value={formData.domain}
								placeholder="example.com"
								class={errors.domain ? 'border-red-500 focus:border-red-500' : ''}
								disabled={loading}
							/>
							{#if errors.domain}
								<p class="text-sm text-red-600">{errors.domain}</p>
							{:else}
								<p class="text-xs text-gray-500">输入要监控的域名，如 example.com</p>
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
							/>
							{#if errors.ip}
								<p class="text-sm text-red-600">{errors.ip}</p>
							{:else}
								<p class="text-xs text-gray-500">输入要监控的IPv4地址</p>
							{/if}
						</div>

					{:else if formData.type === 'url'}
						<div class="space-y-2">
							<Label for="url" class="text-sm font-medium">
								URL地址 <span class="text-red-500">*</span>
							</Label>
							<Input
								id="url"
								type="url"
								bind:value={formData.url}
								placeholder="https://example.com/path"
								class={errors.url ? 'border-red-500 focus:border-red-500' : ''}
								disabled={loading}
							/>
							{#if errors.url}
								<p class="text-sm text-red-600">{errors.url}</p>
							{:else}
								<p class="text-xs text-gray-500">输入要监控的完整URL地址</p>
							{/if}
						</div>

					{:else if formData.type === 'port'}
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
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
								/>
								{#if errors.ip}
									<p class="text-sm text-red-600">{errors.ip}</p>
								{/if}
							</div>
							<div class="space-y-2">
								<Label for="port" class="text-sm font-medium">
									端口号 <span class="text-red-500">*</span>
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
								/>
								{#if errors.port}
									<p class="text-sm text-red-600">{errors.port}</p>
								{/if}
							</div>
							<div class="space-y-2 md:col-span-2">
								<Label for="service" class="text-sm font-medium">服务类型（可选）</Label>
								<Input
									id="service"
									type="text"
									bind:value={formData.service}
									placeholder="http, ssh, mysql"
									disabled={loading}
								/>
								<p class="text-xs text-gray-500">如果已知端口运行的服务类型可以填写</p>
							</div>
						</div>

					{:else if formData.type === 'subdomain'}
						<div class="space-y-2">
							<Label for="subdomain" class="text-sm font-medium">
								子域名 <span class="text-red-500">*</span>
							</Label>
							<Input
								id="subdomain"
								type="text"
								bind:value={formData.subdomain}
								placeholder="sub.example.com"
								class={errors.subdomain ? 'border-red-500 focus:border-red-500' : ''}
								disabled={loading}
							/>
							{#if errors.subdomain}
								<p class="text-sm text-red-600">{errors.subdomain}</p>
							{:else}
								<p class="text-xs text-gray-500">输入要监控的子域名</p>
							{/if}
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
									placeholder="我的Web应用"
									class={errors.appName ? 'border-red-500 focus:border-red-500' : ''}
									disabled={loading}
								/>
								{#if errors.appName}
									<p class="text-sm text-red-600">{errors.appName}</p>
								{/if}
							</div>
							<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
								<div class="space-y-2">
									<Label for="packageName" class="text-sm font-medium">包名（可选）</Label>
									<Input
										id="packageName"
										type="text"
										bind:value={formData.packageName}
										placeholder="com.example.app"
										disabled={loading}
									/>
								</div>
								<div class="space-y-2">
									<Label for="appUrl" class="text-sm font-medium">应用URL（可选）</Label>
									<Input
										id="appUrl"
										type="url"
										bind:value={formData.url}
										placeholder="https://app.example.com"
										disabled={loading}
									/>
								</div>
							</div>
						</div>

					{:else if formData.type === 'http'}
						<div class="space-y-4">
							<div class="space-y-2">
								<Label for="host" class="text-sm font-medium">
									主机地址 <span class="text-red-500">*</span>
								</Label>
								<Input
									id="host"
									type="text"
									bind:value={formData.host}
									placeholder="example.com 或 192.168.1.1"
									class={errors.host ? 'border-red-500 focus:border-red-500' : ''}
									disabled={loading}
								/>
								{#if errors.host}
									<p class="text-sm text-red-600">{errors.host}</p>
								{:else}
									<p class="text-xs text-gray-500">输入主机域名或IP地址</p>
								{/if}
							</div>
							<div class="space-y-2">
								<Label for="httpPort" class="text-sm font-medium">端口号（可选）</Label>
								<Input
									id="httpPort"
									type="number"
									bind:value={formData.port}
									placeholder="80"
									min={1}
									max={65535}
									disabled={loading}
								/>
								<p class="text-xs text-gray-500">默认为80（HTTP）或443（HTTPS）</p>
							</div>
						</div>
					{/if}
				</div>

				<!-- 通用字段 -->
				<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
					<div class="space-y-2">
						<Label for="description" class="text-sm font-medium">描述（可选）</Label>
						<Textarea
							id="description"
							bind:value={formData.description}
							placeholder="简要描述这个资产的用途或重要信息"
							rows={3}
							disabled={loading}
						/>
						<p class="text-xs text-gray-500">添加有助于识别和管理此资产的描述信息</p>
					</div>

					<div class="space-y-2">
						<Label class="text-sm font-medium">标签（可选）</Label>
						<div class="flex flex-wrap gap-2 mb-2">
							{#each formData.tags as tag}
								<Badge variant="secondary" class="flex items-center gap-1">
									{tag}
									<button 
										type="button" 
										onclick={() => removeTag(tag)} 
										class="ml-1 hover:text-red-500"
										disabled={loading}
									>
										<Icon name="x" class="h-3 w-3" />
									</button>
								</Badge>
							{/each}
						</div>
						<div class="flex gap-2">
							<Input
								bind:value={tagInput}
								placeholder="输入标签名称"
								onkeydown={handleTagKeydown}
								disabled={loading}
								class="flex-1"
							/>
							<Button type="button" variant="outline" onclick={addTag} disabled={!tagInput.trim() || loading}>
								<Icon name="plus" class="h-4 w-4" />
							</Button>
						</div>
						<p class="text-xs text-gray-500">按Enter键或点击+按钮添加标签</p>
					</div>
				</div>

				<!-- 项目关联 -->
				<div class="space-y-4">
					<SearchableProjectSelector 
						bind:selectedProjectId={formData.projectId}
						bind:selectedProjectName={selectedProjectName}
						placeholder="搜索项目名称、ID或标签..."
						disabled={loading}
						onProjectSelect={handleProjectSelect}
						class="w-full"
					/>
				</div>

				<!-- 操作按钮 -->
				<div class="flex justify-between pt-6 border-t">
					<Button type="button" variant="outline" onclick={() => goto('/assets')} disabled={loading}>
						取消
					</Button>
					<Button type="submit" disabled={loading}>
						{#if loading}
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
							创建中...
						{:else}
							<Icon name="check" class="h-4 w-4 mr-2" />
							创建资产
						{/if}
					</Button>
				</div>
			</CardContent>
		</Card>
	</form>
</div>
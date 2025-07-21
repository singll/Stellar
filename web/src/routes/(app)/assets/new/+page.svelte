<script lang="ts">
	import { goto } from '$app/navigation';
	import { assetApi } from '$lib/api/asset';
	import type { CreateAssetRequest, AssetType } from '$lib/types/asset';
	import { notifications } from '$lib/stores/notifications';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import Icon from '$lib/components/ui/Icon.svelte';
	import ProjectSelector from '$lib/components/ui/project-selector/ProjectSelector.svelte';
	import { onMount } from 'svelte';
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
	let currentStep = $state(1); // 分步表单：1-选择类型，2-填写信息，3-确认提交
	let showProjectSelector = $state(false);
	
	// 项目选择相关状态
	let selectedProject = $state<Project | null>(null);
	let selectedProjectName = $state('');
	let selectedProjectId = $state('');
	
	// 标签输入
	let tagInput = $state('');

	// 预定义的资产类型
	const assetTypes = [
		{ value: 'domain', label: '域名', description: '监控域名的子域名发现、DNS记录等', icon: 'globe', color: 'blue' },
		{ value: 'subdomain', label: '子域名', description: '监控子域名的解析、可用性等', icon: 'globe', color: 'cyan' },
		{ value: 'ip', label: 'IP地址', description: '监控IP地址的端口扫描、服务识别等', icon: 'server', color: 'green' },
		{ value: 'port', label: '端口', description: '监控特定IP和端口的服务状态', icon: 'wifi', color: 'orange' },
		{ value: 'url', label: 'URL', description: '监控特定URL的内容变化、安全漏洞等', icon: 'link', color: 'purple' },
		{ value: 'http', label: 'HTTP服务', description: '监控HTTP服务的可用性和安全性', icon: 'globe', color: 'indigo' },
		{ value: 'app', label: '移动应用', description: '监控移动应用的安全状态', icon: 'smartphone', color: 'red' },
		{ value: 'miniapp', label: '小程序', description: '监控小程序的安全状态', icon: 'smartphone', color: 'pink' }
	];
	
	// 根据选中状态获取样式类
	const getAssetTypeStyles = (type: any, isSelected: boolean) => {
		if (!isSelected) {
			return {
				container: 'border-gray-200 hover:border-gray-300 bg-white hover:bg-gray-50',
				icon: 'bg-gray-100 text-gray-600',
				radio: 'border-gray-300 bg-white',
				dot: 'opacity-0'
			};
		}
		
		const colorMap = {
			'domain': { border: 'border-blue-500', bg: 'bg-blue-50', iconBg: 'bg-blue-100', iconText: 'text-blue-600', radioBorder: 'border-blue-500', radioBg: 'bg-blue-500' },
			'subdomain': { border: 'border-cyan-500', bg: 'bg-cyan-50', iconBg: 'bg-cyan-100', iconText: 'text-cyan-600', radioBorder: 'border-cyan-500', radioBg: 'bg-cyan-500' },
			'ip': { border: 'border-green-500', bg: 'bg-green-50', iconBg: 'bg-green-100', iconText: 'text-green-600', radioBorder: 'border-green-500', radioBg: 'bg-green-500' },
			'port': { border: 'border-orange-500', bg: 'bg-orange-50', iconBg: 'bg-orange-100', iconText: 'text-orange-600', radioBorder: 'border-orange-500', radioBg: 'bg-orange-500' },
			'url': { border: 'border-purple-500', bg: 'bg-purple-50', iconBg: 'bg-purple-100', iconText: 'text-purple-600', radioBorder: 'border-purple-500', radioBg: 'bg-purple-500' },
			'http': { border: 'border-indigo-500', bg: 'bg-indigo-50', iconBg: 'bg-indigo-100', iconText: 'text-indigo-600', radioBorder: 'border-indigo-500', radioBg: 'bg-indigo-500' },
			'app': { border: 'border-red-500', bg: 'bg-red-50', iconBg: 'bg-red-100', iconText: 'text-red-600', radioBorder: 'border-red-500', radioBg: 'bg-red-500' },
			'miniapp': { border: 'border-pink-500', bg: 'bg-pink-50', iconBg: 'bg-pink-100', iconText: 'text-pink-600', radioBorder: 'border-pink-500', radioBg: 'bg-pink-500' }
		};
		
		const colors = colorMap[type.value as keyof typeof colorMap];
		return {
			container: `${colors.border} ${colors.bg}`,
			icon: `${colors.iconBg} ${colors.iconText}`,
			radio: `${colors.radioBorder} ${colors.radioBg}`,
			dot: 'opacity-100'
		};
	};
	
	// 获取当前选择的资产类型配置
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
	
	// 步骤导航
	const nextStep = () => {
		if (currentStep < 3) {
			currentStep++;
		}
	};
	
	const prevStep = () => {
		if (currentStep > 1) {
			currentStep--;
		}
	};
	
	// 表单验证
	const validateCurrentStep = (): boolean => {
		const newErrors: Record<string, string> = {};

		if (currentStep === 1) {
			if (!formData.type) {
				newErrors.type = '请选择资产类型';
			}
		}
		
		if (currentStep === 2) {
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
		selectedProjectId = project?.id || '';
		formData.projectId = project?.id || '';
		selectedProjectName = project?.name || '';
		showProjectSelector = false;
	};
	
	// 步骤导航处理
	const handleNext = () => {
		if (validateCurrentStep()) {
			nextStep();
		}
	};
	
	// 重置表单
	const resetForm = () => {
		formData.type = 'domain';
		formData.projectId = '';
		formData.tags = [];
		formData.description = '';
		formData.domain = '';
		formData.ip = '';
		formData.url = '';
		formData.port = 80;
		formData.appName = '';
		formData.packageName = '';
		formData.host = '';
		formData.subdomain = '';
		formData.service = '';
		formData.protocol = 'tcp';
		selectedProject = null;
		selectedProjectName = '';
		selectedProjectId = '';
		currentStep = 1;
		errors = {};
	};
</script>

<svelte:head>
	<title>创建资产 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6 max-w-4xl">
	<!-- 页面标题和进度 -->
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

		<!-- 步骤指示器 -->
		<div class="flex items-center justify-center mb-8">
			<div class="flex items-center space-x-4">
				{#each [1, 2, 3] as step, index}
					<div class="flex items-center">
						<div class="flex items-center justify-center w-10 h-10 rounded-full {currentStep >= step ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-600'} transition-colors">
							{#if currentStep > step}
								<Icon name="check" class="h-5 w-5" />
							{:else}
								{step}
							{/if}
						</div>
						{#if index < 2}
							<div class="w-16 h-px {currentStep > step + 1 ? 'bg-blue-600' : 'bg-gray-300'} mx-4"></div>
						{/if}
					</div>
				{/each}
			</div>
		</div>
		
		<div class="text-center mb-6">
			<p class="text-sm text-gray-600">
				{#if currentStep === 1}
					第1步：选择资产类型
				{:else if currentStep === 2}
					第2步：填写资产信息
				{:else}
					第3步：确认并创建
				{/if}
			</p>
		</div>
	</div>

	<!-- 步骤内容 -->
	{#if currentStep === 1}
		<!-- 第一步：选择资产类型 -->
		<Card class="max-w-3xl mx-auto">
			<CardHeader>
				<CardTitle class="flex items-center gap-2">
					<Icon name="layers" class="h-5 w-5 text-blue-600" />
					选择资产类型
				</CardTitle>
				<CardDescription>请选择要创建的安全资产类型</CardDescription>
			</CardHeader>
			<CardContent>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					{#each assetTypes as type}
						{@const isSelected = formData.type === type.value}
						{@const styles = getAssetTypeStyles(type, isSelected)}
						<button
							type="button"
							class="relative cursor-pointer text-left h-32 w-full"
							onclick={() => { formData.type = type.value as any; }}
							disabled={loading}
						>
							<div class="flex items-start gap-3 p-4 rounded-xl border-2 transition-all duration-200 h-full {styles.container}">
								<div class="flex items-center justify-center w-10 h-10 rounded-lg flex-shrink-0 {styles.icon}">
									<Icon name={type.icon} class="h-5 w-5" />
								</div>
								<div class="flex-1 min-w-0 flex flex-col justify-between h-full py-1">
									<div class="flex-1">
										<div class="flex items-center gap-2 mb-1">
											<h3 class="font-semibold text-gray-900 text-sm truncate">{type.label}</h3>
											<Badge variant="outline" class="text-xs flex-shrink-0">{type.value}</Badge>
										</div>
										<p class="text-xs text-gray-600 leading-tight overflow-hidden" style="display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical;">{type.description}</p>
									</div>
								</div>
								<div class="flex items-start justify-center pt-2 flex-shrink-0">
									<div class="flex items-center justify-center w-5 h-5 rounded-full border-2 transition-colors {styles.radio}">
										<div class="w-2.5 h-2.5 rounded-full bg-white transition-opacity {styles.dot}"></div>
									</div>
								</div>
							</div>
						</button>
					{/each}
				</div>
				{#if errors.type}
					<p class="text-sm text-red-600 mt-4">{errors.type}</p>
				{/if}
				
				<div class="flex justify-end mt-8">
					<Button onclick={handleNext} disabled={!formData.type}>
						下一步：填写信息
						<Icon name="chevron-right" class="h-4 w-4 ml-2" />
					</Button>
				</div>
			</CardContent>
		</Card>

	{:else if currentStep === 2}
		<!-- 第二步：填写资产信息 -->
		<Card class="max-w-2xl mx-auto">
			{@const currentType = getCurrentAssetType()}
			{@const iconStyles = currentType ? getAssetTypeStyles(currentType, true) : null}
			<CardHeader>
				<CardTitle class="flex items-center gap-2">
					<Icon name={currentType?.icon || 'settings'} class="h-5 w-5 text-blue-600" />
					填写{currentType?.label || '资产'}信息
				</CardTitle>
				<CardDescription>请填写{currentType?.label || '资产'}的详细信息，标有 * 的字段为必填项</CardDescription>
			</CardHeader>
			<CardContent class="space-y-6">
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

				<!-- 通用字段 -->
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

				<!-- 标签管理 -->
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
					<p class="text-xs text-gray-500">按Enter键或点击+按钮添加标签，用于分类和筛选</p>
				</div>

				<div class="flex justify-between mt-8">
					<Button variant="outline" onclick={prevStep}>
						<Icon name="chevron-left" class="h-4 w-4 mr-2" />
						上一步
					</Button>
					<Button onclick={handleNext}>
						下一步：确认信息
						<Icon name="chevron-right" class="h-4 w-4 ml-2" />
					</Button>
				</div>
			</CardContent>
		</Card>

	{:else}
		<!-- 第三步：确认并创建 -->
		<Card class="max-w-2xl mx-auto">
			<CardHeader>
				<CardTitle class="flex items-center gap-2">
					<Icon name="check-circle" class="h-5 w-5 text-green-600" />
					确认资产信息
				</CardTitle>
				<CardDescription>请确认以下信息无误，然后点击创建按钮</CardDescription>
			</CardHeader>
			<CardContent class="space-y-6">
				{@const currentType = getCurrentAssetType()}
				{@const iconStyles = currentType ? getAssetTypeStyles(currentType, true) : null}
				
				<!-- 资产类型显示 -->
				<div class="flex items-center gap-4 p-4 bg-gray-50 rounded-lg">
					<div class="flex items-center justify-center w-12 h-12 rounded-lg {iconStyles?.icon || 'bg-blue-100 text-blue-600'}">
						<Icon name={currentType?.icon || 'layers'} class="h-6 w-6" />
					</div>
					<div>
						<h3 class="font-semibold text-gray-900">{currentType?.label}</h3>
						<p class="text-sm text-gray-600">{currentType?.description}</p>
					</div>
				</div>

				<!-- 详细信息 -->
				<div class="space-y-4">
					{#if formData.type === 'domain' && formData.domain}
						<div class="flex justify-between py-2 border-b">
							<span class="text-sm font-medium text-gray-600">域名地址</span>
							<span class="text-sm text-gray-900">{formData.domain}</span>
						</div>
					{/if}

					{#if formData.type === 'ip' && formData.ip}
						<div class="flex justify-between py-2 border-b">
							<span class="text-sm font-medium text-gray-600">IP地址</span>
							<span class="text-sm text-gray-900">{formData.ip}</span>
						</div>
					{/if}

					{#if formData.type === 'url' && formData.url}
						<div class="flex justify-between py-2 border-b">
							<span class="text-sm font-medium text-gray-600">URL地址</span>
							<span class="text-sm text-gray-900 break-all">{formData.url}</span>
						</div>
					{/if}

					{#if formData.type === 'port'}
						<div class="flex justify-between py-2 border-b">
							<span class="text-sm font-medium text-gray-600">目标地址</span>
							<span class="text-sm text-gray-900">{formData.ip}:{formData.port}</span>
						</div>
						{#if formData.service}
							<div class="flex justify-between py-2 border-b">
								<span class="text-sm font-medium text-gray-600">服务类型</span>
								<span class="text-sm text-gray-900">{formData.service}</span>
							</div>
						{/if}
					{/if}

					{#if formData.type === 'subdomain' && formData.subdomain}
						<div class="flex justify-between py-2 border-b">
							<span class="text-sm font-medium text-gray-600">子域名</span>
							<span class="text-sm text-gray-900">{formData.subdomain}</span>
						</div>
					{/if}

					{#if formData.type === 'app' && formData.appName}
						<div class="flex justify-between py-2 border-b">
							<span class="text-sm font-medium text-gray-600">应用名称</span>
							<span class="text-sm text-gray-900">{formData.appName}</span>
						</div>
						{#if formData.packageName}
							<div class="flex justify-between py-2 border-b">
								<span class="text-sm font-medium text-gray-600">包名</span>
								<span class="text-sm text-gray-900">{formData.packageName}</span>
							</div>
						{/if}
						{#if formData.url}
							<div class="flex justify-between py-2 border-b">
								<span class="text-sm font-medium text-gray-600">应用URL</span>
								<span class="text-sm text-gray-900 break-all">{formData.url}</span>
							</div>
						{/if}
					{/if}

					{#if formData.type === 'http' && formData.host}
						<div class="flex justify-between py-2 border-b">
							<span class="text-sm font-medium text-gray-600">主机地址</span>
							<span class="text-sm text-gray-900">{formData.host}{formData.port !== 80 ? ':' + formData.port : ''}</span>
						</div>
					{/if}

					{#if formData.description}
						<div class="py-2 border-b">
							<span class="text-sm font-medium text-gray-600">描述</span>
							<p class="text-sm text-gray-900 mt-1">{formData.description}</p>
						</div>
					{/if}

					{#if formData.tags.length > 0}
						<div class="py-2 border-b">
							<span class="text-sm font-medium text-gray-600">标签</span>
							<div class="flex flex-wrap gap-1 mt-1">
								{#each formData.tags as tag}
									<Badge variant="secondary">{tag}</Badge>
								{/each}
							</div>
						</div>
					{/if}

					<!-- 项目关联 -->
					<div class="py-4 border-t">
						<div class="flex items-center justify-between mb-4">
							<Label class="text-sm font-medium">项目关联（可选）</Label>
							<Button 
								variant="outline" 
								size="sm"
								onclick={() => { showProjectSelector = !showProjectSelector; }}
								disabled={loading}
							>
								{selectedProject ? '更改项目' : '选择项目'}
							</Button>
						</div>
						
						{#if selectedProject}
							<div class="flex items-center gap-3 p-3 bg-blue-50 rounded-lg">
								<Icon name="folder" class="h-5 w-5 text-blue-600" />
								<div class="flex-1">
									<p class="font-medium text-blue-900">{selectedProject.name}</p>
									<p class="text-xs text-blue-700">ID: {selectedProject.id}</p>
								</div>
								<Button 
									variant="ghost" 
									size="sm" 
									onclick={() => handleProjectSelect(null)}
									disabled={loading}
								>
									<Icon name="x" class="h-4 w-4" />
								</Button>
							</div>
						{:else}
							<p class="text-sm text-gray-500">未关联项目，资产将添加到默认分组</p>
						{/if}

						{#if showProjectSelector}
							<div class="mt-4">
								<ProjectSelector 
									bind:selectedProjectId={selectedProjectId}
									bind:selectedProjectName={selectedProjectName}
									placeholder="搜索并选择项目"
									disabled={loading}
									onProjectSelect={handleProjectSelect}
								/>
							</div>
						{/if}
					</div>
				</div>

				<div class="flex justify-between mt-8">
					<Button variant="outline" onclick={prevStep} disabled={loading}>
						<Icon name="chevron-left" class="h-4 w-4 mr-2" />
						上一步修改
					</Button>
					<div class="flex gap-3">
						<Button variant="outline" onclick={resetForm} disabled={loading}>
							重新开始
						</Button>
						<Button onclick={handleSubmit} disabled={loading}>
							{#if loading}
								<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
								创建中...
							{:else}
								<Icon name="check" class="h-4 w-4 mr-2" />
								创建资产
							{/if}
						</Button>
					</div>
				</div>
			</CardContent>
		</Card>
	{/if}
</div>

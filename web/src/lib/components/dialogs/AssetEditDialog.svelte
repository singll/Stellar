<!-- 资产编辑对话框 -->
<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { DialogContent, DialogHeader, DialogTitle, DialogFooter } from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import Icon from '$lib/components/ui/Icon.svelte';
	import type { Asset, AssetType, UpdateAssetRequest } from '$lib/types/asset';
	
	interface Props {
		open: boolean;
		asset: Asset | null;
		loading?: boolean;
		onSave: (data: UpdateAssetRequest) => Promise<void>;
		onCancel: () => void;
	}

	let {
		open = $bindable(),
		asset,
		loading = false,
		onSave,
		onCancel
	}: Props = $props();

	// 资产类型选项
	const assetTypeOptions = [
		{ value: 'domain', label: '域名', icon: 'globe' },
		{ value: 'subdomain', label: '子域名', icon: 'globe' },
		{ value: 'ip', label: 'IP地址', icon: 'server' },
		{ value: 'port', label: '端口', icon: 'wifi' },
		{ value: 'url', label: 'URL', icon: 'link' },
		{ value: 'http', label: 'HTTP服务', icon: 'globe' },
		{ value: 'app', label: '应用', icon: 'smartphone' },
		{ value: 'miniapp', label: '小程序', icon: 'smartphone' }
	];

	// 风险等级选项
	const riskLevelOptions = [
		{ value: 'low', label: '低风险', color: 'bg-green-500' },
		{ value: 'medium', label: '中风险', color: 'bg-yellow-500' },
		{ value: 'high', label: '高风险', color: 'bg-orange-500' },
		{ value: 'critical', label: '严重', color: 'bg-red-500' }
	];

	// 表单数据
	let formData = $state({
		name: '',
		description: '',
		value: '', // 资产的主要值
		riskLevel: 'low' as const,
		tags: [] as string[],
		// 根据不同资产类型的特定字段
		domain: '',
		ip: '',
		host: '',
		url: '',
		port: 0,
		appName: '',
		appId: ''
	});

	// 标签输入
	let tagInput = $state('');

	// 表单验证错误
	let errors = $state({
		name: '',
		value: ''
	});

	// 获取资产类型图标
	function getAssetTypeIcon(type?: AssetType): string {
		const option = assetTypeOptions.find(opt => opt.value === type);
		return option?.icon || 'layers';
	}

	// 获取资产类型标签
	function getAssetTypeLabel(type?: AssetType): string {
		const option = assetTypeOptions.find(opt => opt.value === type);
		return option?.label || '未知';
	}

	// 获取资产主要值
	function getAssetMainValue(asset: Asset): string {
		if (asset.value) return asset.value;
		if (asset.name) return asset.name;
		
		const assetData = asset as any;
		switch (asset.type) {
			case 'domain':
				return assetData.domain || '';
			case 'subdomain':
				return assetData.host || assetData.subdomain || '';
			case 'ip':
				return assetData.ip || assetData.address || '';
			case 'port':
				return assetData.ip && assetData.port ? `${assetData.ip}:${assetData.port}` : '';
			case 'url':
				return assetData.url || '';
			case 'http':
				return assetData.url || assetData.host || '';
			case 'app':
			case 'miniapp':
				return assetData.appName || assetData.name || '';
			default:
				return '';
		}
	}

	// 监听资产数据变化，自动填充表单
	$effect(() => {
		if (asset && open) {
			const assetData = asset as any;
			formData = {
				name: asset.name || '',
				description: asset.description || '',
				value: getAssetMainValue(asset),
				riskLevel: asset.riskLevel || 'low',
				tags: asset.tags || [],
				domain: assetData.domain || '',
				ip: assetData.ip || assetData.address || '',
				host: assetData.host || '',
				url: assetData.url || '',
				port: assetData.port || 0,
				appName: assetData.appName || '',
				appId: assetData.appId || ''
			};
			// 清除错误
			errors = { name: '', value: '' };
		}
	});

	// 表单验证
	function validateForm(): boolean {
		let isValid = true;
		errors = { name: '', value: '' };

		if (!formData.value.trim()) {
			errors.value = '资产值不能为空';
			isValid = false;
		}

		return isValid;
	}

	// 添加标签
	function addTag() {
		const tag = tagInput.trim();
		if (tag && !formData.tags.includes(tag)) {
			formData.tags = [...formData.tags, tag];
			tagInput = '';
		}
	}

	// 删除标签
	function removeTag(index: number) {
		formData.tags = formData.tags.filter((_, i) => i !== index);
	}

	// 处理标签输入回车
	function handleTagKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			addTag();
		}
	}

	// 构建更新数据
	function buildUpdateData(): UpdateAssetRequest {
		const baseData = {
			name: formData.name.trim() || undefined,
			description: formData.description.trim() || undefined,
			value: formData.value.trim(),
			riskLevel: formData.riskLevel,
			tags: formData.tags
		};

		// 根据资产类型构建特定数据
		const specificData: Record<string, any> = { ...baseData };

		if (asset) {
			switch (asset.type) {
				case 'domain':
					specificData.domain = formData.value.trim();
					break;
				case 'subdomain':
					specificData.host = formData.value.trim();
					break;
				case 'ip':
					specificData.ip = formData.value.trim();
					break;
				case 'port':
					// 解析 IP:Port 格式
					const parts = formData.value.split(':');
					if (parts.length === 2) {
						specificData.ip = parts[0].trim();
						specificData.port = parseInt(parts[1].trim()) || 0;
					}
					break;
				case 'url':
				case 'http':
					specificData.url = formData.value.trim();
					break;
				case 'app':
				case 'miniapp':
					specificData.appName = formData.value.trim();
					if (formData.appId) {
						specificData.appId = formData.appId.trim();
					}
					break;
			}
		}

		return {
			type: asset!.type,
			data: specificData
		};
	}

	// 提交表单
	async function handleSave() {
		if (!validateForm()) {
			return;
		}

		try {
			await onSave(buildUpdateData());
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
				<div class="flex items-center justify-center w-10 h-10 rounded-full bg-purple-100">
					<Icon name={getAssetTypeIcon(asset?.type)} class="h-5 w-5 text-purple-600" />
				</div>
				<div>
					<DialogTitle>编辑{getAssetTypeLabel(asset?.type)}</DialogTitle>
					<p class="text-sm text-gray-500 mt-1">资产ID: {asset?.id}</p>
				</div>
			</div>
		</DialogHeader>

		<div class="space-y-6 py-4">
			<!-- 资产值 -->
			<div class="space-y-2">
				<Label for="asset-value" class="text-sm font-medium">
					{getAssetTypeLabel(asset?.type)}值 <span class="text-red-500">*</span>
				</Label>
				<Input
					id="asset-value"
					bind:value={formData.value}
					placeholder="请输入资产值"
					class={errors.value ? 'border-red-500' : ''}
					disabled={loading}
				/>
				{#if errors.value}
					<p class="text-sm text-red-500">{errors.value}</p>
				{/if}
			</div>

			<!-- 资产名称 -->
			<div class="space-y-2">
				<Label for="asset-name" class="text-sm font-medium">
					资产名称
				</Label>
				<Input
					id="asset-name"
					bind:value={formData.name}
					placeholder="请输入资产名称（可选）"
					disabled={loading}
				/>
			</div>

			<!-- App ID (仅对应用和小程序显示) -->
			{#if asset?.type === 'app' || asset?.type === 'miniapp'}
				<div class="space-y-2">
					<Label for="asset-app-id" class="text-sm font-medium">
						{asset.type === 'app' ? '包名' : '小程序ID'}
					</Label>
					<Input
						id="asset-app-id"
						bind:value={formData.appId}
						placeholder="请输入{asset.type === 'app' ? '应用包名' : '小程序ID'}"
						disabled={loading}
					/>
				</div>
			{/if}

			<!-- 资产描述 -->
			<div class="space-y-2">
				<Label for="asset-description" class="text-sm font-medium">
					资产描述
				</Label>
				<Textarea
					id="asset-description"
					bind:value={formData.description}
					placeholder="请输入资产描述（可选）"
					rows={3}
					disabled={loading}
				/>
			</div>

			<!-- 风险等级 -->
			<div class="space-y-2">
				<Label class="text-sm font-medium">风险等级</Label>
				<div class="flex flex-wrap gap-2">
					{#each riskLevelOptions as risk}
						<button
							type="button"
							class="flex items-center gap-2 px-3 py-2 text-sm border rounded-md transition-colors hover:bg-gray-50 {formData.riskLevel === risk.value ? 'border-blue-500 bg-blue-50' : 'border-gray-300'}"
							onclick={() => formData.riskLevel = risk.value as any}
							disabled={loading}
						>
							<div class="w-4 h-4 rounded-full {risk.color}"></div>
							{risk.label}
						</button>
					{/each}
				</div>
			</div>

			<!-- 标签 -->
			<div class="space-y-2">
				<Label class="text-sm font-medium">标签</Label>
				
				<!-- 现有标签 -->
				{#if formData.tags.length > 0}
					<div class="flex flex-wrap gap-2 mb-2">
						{#each formData.tags as tag, index}
							<span class="inline-flex items-center gap-1 px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded">
								{tag}
								<button
									type="button"
									onclick={() => removeTag(index)}
									class="text-blue-600 hover:text-blue-800"
									disabled={loading}
								>
									<Icon name="x" class="h-3 w-3" />
								</button>
							</span>
						{/each}
					</div>
				{/if}

				<!-- 添加标签 -->
				<div class="flex gap-2">
					<Input
						bind:value={tagInput}
						placeholder="输入标签名称"
						onkeydown={handleTagKeydown}
						disabled={loading}
						class="flex-1"
					/>
					<Button
						type="button"
						variant="outline"
						onclick={addTag}
						disabled={loading || !tagInput.trim()}
					>
						<Icon name="plus" class="h-4 w-4" />
					</Button>
				</div>
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
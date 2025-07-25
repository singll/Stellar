<!-- 资产列表页面 -->
<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import PageLayout from '$lib/components/ui/page-layout/PageLayout.svelte';
	import StatsGrid from '$lib/components/ui/stats-grid/StatsGrid.svelte';
	import DataList from '$lib/components/ui/data-list/DataList.svelte';
	import DeleteConfirmDialog from '$lib/components/dialogs/DeleteConfirmDialog.svelte';
	import AssetEditDialog from '$lib/components/dialogs/AssetEditDialog.svelte';
	import { assetApi } from '$lib/api/asset';
	import type { Asset } from '$lib/types/asset';
	import { notifications } from '$lib/stores/notifications';
	import { goto } from '$app/navigation';

	let assets: Asset[] = $state([]);
	let loading = $state(true);
	let searchQuery = $state('');

	// 弹窗状态
	let deleteDialogOpen = $state(false);
	let editDialogOpen = $state(false);
	let selectedAsset = $state<Asset | null>(null);
	let dialogLoading = $state(false);

	// 搜索处理
	const handleSearch = async (query?: string) => {
		if (query !== undefined) {
			searchQuery = query;
		}
		// 重新加载资产数据（这里可以将搜索传递给API）
		await loadAssets();
	};

	// 获取资产显示名称
	function getAssetDisplayName(asset: Asset): string {
		console.log(`获取资产显示名称 - 资产ID: ${asset.id}, 类型: ${asset.type}`, asset);
		
		// 首先尝试使用通用字段
		if (asset.name) {
			console.log(`使用name字段: ${asset.name}`);
			return asset.name;
		}
		if (asset.value) {
			console.log(`使用value字段: ${asset.value}`);
			return asset.value;
		}
		
		// 然后根据类型尝试特定字段
		const assetData = asset as any;
		let result = '';
		
		switch (asset.type) {
			case 'domain':
				result = assetData.domain || assetData.Domain || asset.id;
				break;
			case 'subdomain':
				result = assetData.host || assetData.Host || assetData.subdomain || assetData.Subdomain || asset.id;
				break;
			case 'ip':
				result = assetData.ip || assetData.IP || assetData.address || assetData.Address || asset.id;
				break;
			case 'port':
				if (assetData.ip && assetData.port) {
					result = `${assetData.ip}:${assetData.port}`;
				} else if (assetData.IP && assetData.Port) {
					result = `${assetData.IP}:${assetData.Port}`;
				} else {
					result = asset.id;
				}
				break;
			case 'url':
				result = assetData.url || assetData.URL || assetData.link || assetData.Link || asset.id;
				break;
			case 'http':
				result = assetData.url || assetData.URL || 
					(assetData.host && assetData.port ? `${assetData.host}:${assetData.port}` : '') ||
					(assetData.Host && assetData.Port ? `${assetData.Host}:${assetData.Port}` : '') ||
					asset.id;
				break;
			case 'app':
			case 'miniapp':
				result = assetData.appName || assetData.AppName || assetData.name || assetData.Name || asset.id;
				break;
			default:
				result = assetData.title || assetData.Title || asset.id;
		}
		
		console.log(`最终显示名称: ${result}`);
		return result;
	}

	// 获取资产URL/IP信息
	function getAssetUrlOrIp(asset: Asset): string {
		const assetData = asset as any;
		
		switch (asset.type) {
			case 'domain':
				const domainIps = assetData.ips || assetData.IPs || assetData.ipAddresses;
				if (Array.isArray(domainIps) && domainIps.length > 0) {
					return domainIps[0];
				}
				return assetData.domain || assetData.Domain || '-';
			case 'subdomain':
				const subdomainIps = assetData.ips || assetData.IPs || assetData.ipAddresses;
				if (Array.isArray(subdomainIps) && subdomainIps.length > 0) {
					return subdomainIps[0];
				}
				return assetData.host || assetData.Host || assetData.subdomain || assetData.Subdomain || '-';
			case 'ip':
				return assetData.ip || assetData.IP || assetData.address || assetData.Address || '-';
			case 'port':
				return assetData.ip || assetData.IP || assetData.address || assetData.Address || '-';
			case 'url':
				return assetData.url || assetData.URL || assetData.link || assetData.Link || '-';
			case 'http':
				if (assetData.url || assetData.URL) {
					return assetData.url || assetData.URL;
				}
				if (assetData.host && assetData.port) {
					return `http://${assetData.host}:${assetData.port}`;
				}
				if (assetData.Host && assetData.Port) {
					return `http://${assetData.Host}:${assetData.Port}`;
				}
				return '-';
			case 'app':
				return assetData.downloadUrl || assetData.packageName || assetData.appUrl || '-';
			case 'miniapp':
				return assetData.qrCodeUrl || assetData.appId || assetData.miniAppId || '-';
			default:
				return assetData.value || assetData.url || assetData.address || '-';
		}
	}

	// 获取资产状态
	function getAssetStatus(asset: Asset): { text: string; color: string } {
		const assetData = asset as any;
		const status = asset.status || assetData.Status || 'active';
		
		// 检查各种可能的时间字段
		const lastScanTime = asset.lastScanTime || assetData.lastScanTime || assetData.LastScanTime || 
							assetData.lastScan || assetData.LastScan || assetData.scanTime;
		
		const lastScan = lastScanTime ? new Date(lastScanTime) : null;
		const now = new Date();
		const daysSinceLastScan = lastScan ? Math.floor((now.getTime() - lastScan.getTime()) / (1000 * 60 * 60 * 24)) : null;

		// 如果超过7天没有扫描，标记为过期
		if (daysSinceLastScan !== null && daysSinceLastScan > 7) {
			return { text: '需要扫描', color: 'text-yellow-700 border-yellow-200 bg-yellow-50' };
		}

		switch (status.toLowerCase()) {
			case 'active':
				return { text: '正常', color: 'text-green-700 border-green-200 bg-green-50' };
			case 'inactive':
				return { text: '未激活', color: 'text-gray-700 border-gray-200 bg-gray-50' };
			case 'deleted':
				return { text: '已删除', color: 'text-red-700 border-red-200 bg-red-50' };
			default:
				return { text: '未知', color: 'text-gray-700 border-gray-200 bg-gray-50' };
		}
	}

	// 获取风险等级
	function getRiskLevel(asset: Asset): { text: string; color: string } {
		const assetData = asset as any;
		const riskLevel = asset.riskLevel || assetData.riskLevel || assetData.RiskLevel || assetData.risk || assetData.Risk || 'low';
		
		switch (riskLevel.toLowerCase()) {
			case 'critical':
				return { text: '严重', color: 'text-red-700 border-red-200 bg-red-50' };
			case 'high':
				return { text: '高风险', color: 'text-orange-700 border-orange-200 bg-orange-50' };
			case 'medium':
				return { text: '中风险', color: 'text-yellow-700 border-yellow-200 bg-yellow-50' };
			case 'low':
				return { text: '低风险', color: 'text-green-700 border-green-200 bg-green-50' };
			default:
				return { text: '未评估', color: 'text-gray-700 border-gray-200 bg-gray-50' };
		}
	}

	// 格式化最后扫描时间
	function formatLastScanTime(asset: Asset): string {
		const assetData = asset as any;
		const timestamp = asset.lastScanTime || assetData.lastScanTime || assetData.LastScanTime || 
						 assetData.lastScan || assetData.LastScan || assetData.scanTime ||
						 asset.updatedAt || assetData.updatedAt || assetData.UpdatedAt;
		
		if (!timestamp) return '从未';
		
		const date = new Date(timestamp);
		if (isNaN(date.getTime())) return '从未';
		
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		const diffMinutes = Math.floor(diffMs / (1000 * 60));

		if (diffDays > 0) {
			return `${diffDays}天前`;
		} else if (diffHours > 0) {
			return `${diffHours}小时前`;
		} else if (diffMinutes > 0) {
			return `${diffMinutes}分钟前`;
		} else {
			return '刚刚';
		}
	}

	// 获取资产类型图标
	function getAssetTypeIcon(type: string): string {
		switch (type) {
			case 'domain': return 'globe';
			case 'subdomain': return 'globe';
			case 'ip': return 'server';
			case 'port': return 'wifi';
			case 'url': return 'link';
			case 'http': return 'globe';
			case 'app': return 'smartphone';
			case 'miniapp': return 'smartphone';
			default: return 'layers';
		}
	}

	// 获取资产类型标签颜色
	function getAssetTypeColor(type: string): string {
		switch (type) {
			case 'domain': return 'bg-blue-100 text-blue-800';
			case 'subdomain': return 'bg-blue-100 text-blue-800';
			case 'ip': return 'bg-green-100 text-green-800';
			case 'port': return 'bg-orange-100 text-orange-800';
			case 'url': return 'bg-purple-100 text-purple-800';
			case 'http': return 'bg-purple-100 text-purple-800';
			case 'app': return 'bg-red-100 text-red-800';
			case 'miniapp': return 'bg-red-100 text-red-800';
			default: return 'bg-gray-100 text-gray-800';
		}
	}

	async function loadAssets() {
		try {
			loading = true;
			const response = await assetApi.getAllAssets();
			console.log('完整API响应:', response);
			
			// API返回的是 { data: Asset[], message: string, success: boolean }
			assets = Array.isArray(response.data) ? response.data : [];
			
			// 详细的调试信息
			console.log('解析后的资产数组:', assets);
			console.log('资产数量:', assets.length);
			
			if (assets.length > 0) {
				console.log('第一个资产的结构:', assets[0]);
				console.log('第一个资产的字段:', Object.keys(assets[0]));
			} else {
				console.log('资产数据为空，可能是后端没有返回资产数据');
			}
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '加载资产列表失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
			console.error('加载资产列表失败:', error);
			assets = [];
		} finally {
			loading = false;
		}
	}

	function handleAssetClick(id: string) {
		goto(`/assets/${id}`);
	}

	// 处理编辑资产
	const handleEditAsset = (assetId: string) => {
		const asset = assets.find(a => a.id === assetId);
		if (asset) {
			selectedAsset = asset;
			editDialogOpen = true;
		}
	};

	// 处理删除资产
	const handleDeleteAsset = (assetId: string) => {
		const asset = assets.find(a => a.id === assetId);
		if (asset) {
			selectedAsset = asset;
			deleteDialogOpen = true;
		}
	};

	// 确认删除资产
	const confirmDeleteAsset = async () => {
		if (!selectedAsset) return;

		try {
			dialogLoading = true;
			await assetApi.deleteAsset(selectedAsset.id);
			assets = assets.filter((a) => a.id !== selectedAsset.id);
			notifications.add({
				type: 'success',
				message: '资产删除成功'
			});
			deleteDialogOpen = false;
			selectedAsset = null;
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '删除资产失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		} finally {
			dialogLoading = false;
		}
	};

	// 保存资产编辑
	const saveAssetEdit = async (data: any) => {
		if (!selectedAsset) return;

		try {
			dialogLoading = true;
			await assetApi.updateAsset(selectedAsset.id, data);
			
			// 重新加载资产列表以获取最新数据
			await loadAssets();
			
			notifications.add({
				type: 'success',
				message: '资产更新成功'
			});
			editDialogOpen = false;
			selectedAsset = null;
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '更新资产失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
			throw error; // 重新抛出错误，让弹窗保持打开状态
		} finally {
			dialogLoading = false;
		}
	};

	// 取消弹窗操作
	const handleDialogCancel = () => {
		selectedAsset = null;
	};

	// 准备统计数据
	const statsData = $derived(assets.length > 0 ? [
		{
			title: '总资产',
			value: assets.length,
			icon: 'database',
			color: 'blue' as const
		},
		{
			title: '域名资产',
			value: assets.filter(a => a.type === 'domain' || a.type === 'subdomain').length,
			icon: 'globe',
			color: 'green' as const
		},
		{
			title: '网络资产',
			value: assets.filter(a => a.type === 'ip' || a.type === 'port' || a.type === 'url').length,
			icon: 'server',
			color: 'purple' as const
		},
		{
			title: '应用资产',
			value: assets.filter(a => a.type === 'app' || a.type === 'miniapp').length,
			icon: 'smartphone',
			color: 'orange' as const
		}
	] : []);

	// 准备表格列配置
	const columns = [
		{
			key: 'name',
			title: '名称',
			render: (value: any, row: Asset) => {
				const displayName = getAssetDisplayName(row);
				// 根据资产类型获取图标路径
				const getIconSvg = (type: string) => {
					switch (type) {
						case 'domain':
						case 'subdomain':
							return '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9v-9m0-9v9"></path>';
						case 'ip':
							return '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h6a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h6a2 2 0 002-2v-4a2 2 0 00-2-2m8 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v3m2 4h1.01M21 16h1.01"></path>';
						case 'port':
							return '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"></path>';
						case 'url':
						case 'http':
							return '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path>';
						case 'app':
						case 'miniapp':
							return '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z"></path>';
						default:
							return '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>';
					}
				};
				return `
					<div class="flex items-center gap-3">
						<div class="flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 group-hover:bg-gray-200 transition-colors">
							<svg class="h-4 w-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								${getIconSvg(row.type)}
							</svg>
						</div>
						<span class="font-medium">${displayName}</span>
					</div>
				`;
			}
		},
		{
			key: 'type',
			title: '类型',
			render: (value: any, row: Asset) => {
				const colorClass = getAssetTypeColor(row.type);
				return `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${colorClass}">${row.type}</span>`;
			}
		},
		{
			key: 'urlOrIp',
			title: 'URL/IP',
			render: (value: any, row: Asset) => {
				const urlOrIp = getAssetUrlOrIp(row);
				return `<span class="font-mono text-xs bg-gray-100 px-2 py-1 rounded">${urlOrIp}</span>`;
			}
		},
		{
			key: 'status',
			title: '状态',
			render: (value: any, row: Asset) => {
				const status = getAssetStatus(row);
				return `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${status.color}">${status.text}</span>`;
			}
		},
		{
			key: 'lastScan',
			title: '最后扫描',
			render: (value: any, row: Asset) => {
				const timeText = formatLastScanTime(row);
				return `
					<div class="flex items-center gap-2 text-gray-500 text-sm">
						<svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<circle cx="12" cy="12" r="10"></circle>
							<polyline points="12,6 12,12 16,14"></polyline>
						</svg>
						${timeText}
					</div>
				`;
			}
		},
		{
			key: 'risk',
			title: '风险等级',
			render: (value: any, row: Asset) => {
				const risk = getRiskLevel(row);
				return `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${risk.color}">${risk.text}</span>`;
			}
		}
	];

	onMount(() => {
		loadAssets();
	});
</script>

<PageLayout
	title="资产管理"
	description="管理和监控您的所有安全资产"
	icon="shield"
	showStats={!loading && assets.length > 0}
	actions={[
		{
			text: '添加资产',
			icon: 'plus',
			variant: 'default',
			onClick: () => goto('/assets/new')
		}
	]}
>
	{#snippet stats()}
		<StatsGrid stats={statsData} />
	{/snippet}

	<DataList
		title=""
		{columns}
		data={assets}
		{loading}
		searchPlaceholder="搜索资产..."
		searchValue={searchQuery}
		onSearch={handleSearch}
		emptyStateTitle="暂无资产"
		emptyStateDescription="您还没有添加任何资产，开始添加第一个资产吧"
		emptyStateAction={{
			text: '创建第一个资产',
			icon: 'plus',
			onClick: () => goto('/assets/new')
		}}
		onRowClick={(asset) => handleAssetClick(asset.id)}
		rowActions={(row) => [
			{
				icon: 'edit',
				title: '编辑资产',
				variant: 'ghost',
				onClick: () => handleEditAsset(row.id)
			},
			{
				icon: 'trash',
				title: '删除资产',
				variant: 'ghost',
				color: 'red',
				onClick: () => handleDeleteAsset(row.id)
			}
		]}
	/>
</PageLayout>

<!-- 弹窗组件 -->
<DeleteConfirmDialog
	bind:open={deleteDialogOpen}
	itemName={selectedAsset ? getAssetDisplayName(selectedAsset) : ''}
	itemType="资产"
	loading={dialogLoading}
	onConfirm={confirmDeleteAsset}
	onCancel={handleDialogCancel}
/>

<AssetEditDialog
	bind:open={editDialogOpen}
	asset={selectedAsset}
	loading={dialogLoading}
	onSave={saveAssetEdit}
	onCancel={handleDialogCancel}
/>

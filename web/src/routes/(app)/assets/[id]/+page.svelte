<!-- 资产详情页面 -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { assetApi } from '$lib/api/asset';
	import type { Asset } from '$lib/types/asset';
	import { notifications } from '$lib/stores/notifications';
	import { goto } from '$app/navigation';
	import Icon from '@iconify/svelte';
	import AssetDetails from '$lib/components/assets/AssetDetails.svelte';

	let asset: Asset | null = $state(null);
	let loading = $state(true);
	let assetType = $derived($page.url.searchParams.get('type') || '');

	async function loadAsset() {
		if (!assetType) {
			notifications.add({
				type: 'error',
				message: '缺少资产类型参数'
			});
			goto('/assets');
			return;
		}

		try {
			loading = true;
			const response = await assetApi.getAssetById($page.params.id, assetType);
			asset = response.data;
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '加载资产详情失败'
			});
			goto('/assets');
		} finally {
			loading = false;
		}
	}

	async function handleDelete() {
		if (!asset || !confirm('确定要删除此资产吗？')) {
			return;
		}

		try {
			await assetApi.deleteAsset($page.params.id, assetType);
			notifications.add({
				type: 'success',
				message: '资产已删除'
			});
			goto('/assets');
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '删除资产失败'
			});
		}
	}

	function handleEdit() {
		if (!asset) return;
		goto(`/assets/${asset.id}/edit?type=${assetType}`);
	}

	function handleScan() {
		// TODO: 实现扫描功能
		notifications.add({
			type: 'info',
			message: '扫描功能正在开发中'
		});
	}

	onMount(() => {
		loadAsset();
	});
</script>

<svelte:head>
	<title>资产详情 - Stellar</title>
</svelte:head>

<div class="container mx-auto p-4 space-y-6">
	<!-- 头部 -->
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-4">
			<Button variant="ghost" size="sm" onclick={() => goto('/assets')}>
				<Icon icon="tabler:arrow-left" width={16} class="h-4 w-4" />
				返回
			</Button>
			<div>
				<h1 class="text-2xl font-bold">资产详情</h1>
				<p class="text-muted-foreground">查看资产的详细信息和变更历史</p>
			</div>
		</div>

		{#if asset}
			<div class="flex items-center gap-2">
				<Button variant="outline" size="sm" onclick={handleEdit}>
					<Icon icon="tabler:edit" width={16} class="h-4 w-4 mr-2" />
					编辑
				</Button>
				<Button variant="outline" size="sm" onclick={handleScan}>
					<Icon icon="tabler:wifi" width={16} class="h-4 w-4 mr-2" />
					扫描
				</Button>
				<Button variant="destructive" size="sm" onclick={handleDelete}>
					<Icon icon="tabler:trash" width={16} class="h-4 w-4 mr-2" />
					删除
				</Button>
			</div>
		{/if}
	</div>

	<!-- 内容 -->
	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="text-center">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
				<p class="text-muted-foreground">加载中...</p>
			</div>
		</div>
	{:else if asset}
		<AssetDetails {asset} />
	{:else}
		<div class="flex items-center justify-center py-12">
			<div class="text-center">
				<h3 class="text-lg font-semibold mb-2">未找到资产</h3>
				<p class="text-muted-foreground mb-4">请检查资产ID和类型参数是否正确</p>
				<Button onclick={() => goto('/assets')}>返回资产列表</Button>
			</div>
		</div>
	{/if}
</div>
"

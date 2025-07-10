<!-- 资产列表页面 -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { assetApi } from '$lib/api/asset';
	import type { Asset } from '$lib/types/asset';
	import { notifications } from '$lib/stores/notifications';
	import { goto } from '$app/navigation';

	let assets: Asset[] = $state([]);
	let loading = $state(true);
	let searchQuery = $state('');

	async function loadAssets() {
		try {
			loading = true;
			const response = await assetApi.getAssets();
			assets = response.data;
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '加载资产列表失败'
			});
		} finally {
			loading = false;
		}
	}

	function handleAssetClick(id: string) {
		goto(`/assets/${id}`);
	}

	onMount(() => {
		loadAssets();
	});
</script>

<div class="container mx-auto p-4 space-y-4">
	<Card>
		<CardHeader>
			<CardTitle>资产管理</CardTitle>
			<CardDescription>管理和监控您的所有安全资产</CardDescription>
		</CardHeader>
		<CardContent>
			<!-- 工具栏 -->
			<div class="flex justify-between items-center mb-4">
				<div class="flex gap-2 items-center">
					<div class="relative">
						<Icon name="search" class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
						<Input type="search" placeholder="搜索资产..." class="pl-8" bind:value={searchQuery} />
					</div>
				</div>
				<Button onclick={() => goto('/assets/new')}>添加资产</Button>
			</div>

			<!-- 资产列表 -->
			<div class="rounded-md border">
				<Table>
					<TableHeader>
						<TableRow>
							<TableHead>名称</TableHead>
							<TableHead>类型</TableHead>
							<TableHead>URL/IP</TableHead>
							<TableHead>状态</TableHead>
							<TableHead>最后扫描</TableHead>
							<TableHead>风险等级</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{#if loading}
							<TableRow>
								<td colspan="6" class="text-center py-4">加载中...</td>
							</TableRow>
						{:else if assets.length === 0}
							<TableRow>
								<td colspan="6" class="text-center py-4">暂无资产</td>
							</TableRow>
						{:else}
							{#each assets as asset}
								<tr
									class="cursor-pointer hover:bg-muted/50 border-b transition-colors"
									onclick={() => handleAssetClick(asset.id)}
									onkeydown={(e) => {
										if (e.key === 'Enter' || e.key === ' ') {
											handleAssetClick(asset.id);
										}
									}}
									role="button"
									tabindex="0"
								>
									<TableCell>
										{#if asset.type === 'domain'}
											{(asset as any).domain}
										{:else if asset.type === 'ip'}
											{(asset as any).ip}
										{:else if asset.type === 'url'}
											{(asset as any).url}
										{:else if asset.type === 'app'}
											{(asset as any).appName}
										{:else}
											{asset.id}
										{/if}
									</TableCell>
									<TableCell>{asset.type}</TableCell>
									<TableCell>
										{#if asset.type === 'url' || asset.type === 'http'}
											{(asset as any).url}
										{:else if asset.type === 'ip' || asset.type === 'port'}
											{(asset as any).ip}
										{:else}
											-
										{/if}
									</TableCell>
									<TableCell>
										<Badge variant="default">正常</Badge>
									</TableCell>
									<TableCell>{new Date(asset.lastScanTime).toLocaleString()}</TableCell>
									<TableCell>
										<Badge variant="default">正常</Badge>
									</TableCell>
								</tr>
							{/each}
						{/if}
					</TableBody>
				</Table>
			</div>
		</CardContent>
	</Card>
</div>
